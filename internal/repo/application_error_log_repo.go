package repo

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type ApplicationErrorLogRepo interface {
	Insert(ctx context.Context, e domain.ApplicationErrorLog) error
	List(ctx context.Context, filter ApplicationErrorLogListFilter) ([]domain.ApplicationErrorLog, int, error)
	GetByID(ctx context.Context, id string) (domain.ApplicationErrorLog, error)
	MarkResolved(ctx context.Context, id, adminUserID string, note *string) error
	MarkUnresolved(ctx context.Context, id string) error
	UpdateAdminNote(ctx context.Context, id string, note string) error
	AnalyticsSummary(ctx context.Context, from, to time.Time) (ApplicationErrorAnalytics, error)
}

type ApplicationErrorLogListFilter struct {
	ErrorType  string
	HTTPStatus int
	Resolved   *bool
	From       *time.Time
	To         *time.Time
	Search     string
	Offset     int
	Limit      int
}

type ApplicationErrorAnalytics struct {
	Total        int64                         `json:"total"`
	Unresolved   int64                         `json:"unresolved"`
	ByType       []ApplicationErrorTypeCount   `json:"byType"`
	ByHTTPStatus []ApplicationErrorStatusCount `json:"byHttpStatus"`
	ByDay        []ApplicationErrorDayCount    `json:"byDay"`
}

type ApplicationErrorTypeCount struct {
	ErrorType string `json:"errorType"`
	Count     int64  `json:"count"`
}

type ApplicationErrorStatusCount struct {
	HTTPStatus int   `json:"httpStatus"`
	Count      int64 `json:"count"`
}

type ApplicationErrorDayCount struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

type applicationErrorLogRepo struct{ pool *pgxpool.Pool }

func NewApplicationErrorLogRepo(pool *pgxpool.Pool) ApplicationErrorLogRepo {
	return &applicationErrorLogRepo{pool: pool}
}

func (r *applicationErrorLogRepo) Insert(ctx context.Context, e domain.ApplicationErrorLog) error {
	metaJSON := []byte("{}")
	if len(e.Meta) > 0 {
		if b, err := json.Marshal(e.Meta); err == nil {
			metaJSON = b
		}
	}
	var userID interface{}
	if e.UserID != nil && strings.TrimSpace(*e.UserID) != "" {
		userID = strings.TrimSpace(*e.UserID)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO application_error_logs (
			error_type, error_code, message, http_status, method, path, query_string,
			user_id, user_role, request_id, ip_address, user_agent, stack_trace, meta
		) VALUES (
			$1, NULLIF(trim($2), ''), $3, $4, $5, $6, NULLIF(trim($7), ''),
			$8::uuid, NULLIF(trim($9), ''), NULLIF(trim($10), ''),
			NULLIF(trim($11), ''), NULLIF(trim($12), ''), NULLIF(trim($13), ''), $14::jsonb
		)
	`, e.ErrorType, ptrStr(e.ErrorCode), e.Message, e.HTTPStatus, e.Method, e.Path, ptrStr(e.QueryString),
		userID, ptrStr(e.UserRole), ptrStr(e.RequestID), ptrStr(e.IPAddress), ptrStr(e.UserAgent), ptrStr(e.StackTrace), metaJSON)
	return err
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (r *applicationErrorLogRepo) listWhereArgs(filter ApplicationErrorLogListFilter) (string, []any) {
	var parts []string
	var args []any
	n := 1
	if filter.ErrorType != "" {
		parts = append(parts, "error_type = $"+strconv.Itoa(n))
		args = append(args, filter.ErrorType)
		n++
	}
	if filter.HTTPStatus > 0 {
		parts = append(parts, "http_status = $"+strconv.Itoa(n))
		args = append(args, filter.HTTPStatus)
		n++
	}
	if filter.Resolved != nil {
		if *filter.Resolved {
			parts = append(parts, "resolved_at IS NOT NULL")
		} else {
			parts = append(parts, "resolved_at IS NULL")
		}
	}
	if filter.From != nil {
		parts = append(parts, "created_at >= $"+strconv.Itoa(n))
		args = append(args, *filter.From)
		n++
	}
	if filter.To != nil {
		parts = append(parts, "created_at < $"+strconv.Itoa(n))
		args = append(args, *filter.To)
		n++
	}
	if s := strings.TrimSpace(filter.Search); s != "" {
		pat := "%" + s + "%"
		parts = append(parts, "(message ILIKE $"+strconv.Itoa(n)+" OR path ILIKE $"+strconv.Itoa(n+1)+")")
		args = append(args, pat, pat)
		n += 2
	}
	if len(parts) == 0 {
		return "TRUE", args
	}
	return strings.Join(parts, " AND "), args
}

func (r *applicationErrorLogRepo) List(ctx context.Context, filter ApplicationErrorLogListFilter) ([]domain.ApplicationErrorLog, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	wclause, args := r.listWhereArgs(filter)

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM application_error_logs WHERE `+wclause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	qargs := append(append([]any{}, args...), filter.Limit, filter.Offset)
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, created_at, error_type, error_code, message, http_status, method, path, query_string,
			user_id::text, user_role, request_id, ip_address, user_agent, stack_trace, meta,
			resolved_at, resolved_by::text, admin_note
		FROM application_error_logs
		WHERE `+wclause+`
		ORDER BY created_at DESC
		LIMIT $`+strconv.Itoa(limitArg)+` OFFSET $`+strconv.Itoa(offsetArg),
		qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []domain.ApplicationErrorLog
	for rows.Next() {
		e, err := scanApplicationErrorLog(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanApplicationErrorLog(row rowScanner) (domain.ApplicationErrorLog, error) {
	var e domain.ApplicationErrorLog
	var errCode, q, uid, urole, rid, ip, ua, stack, rb, note *string
	var metaBytes []byte
	var resolvedAt *time.Time
	if err := row.Scan(&e.ID, &e.CreatedAt, &e.ErrorType, &errCode, &e.Message, &e.HTTPStatus, &e.Method, &e.Path,
		&q, &uid, &urole, &rid, &ip, &ua, &stack, &metaBytes, &resolvedAt, &rb, &note); err != nil {
		return e, err
	}
	e.ErrorCode = errCode
	e.QueryString = q
	e.UserID = uid
	e.UserRole = urole
	e.RequestID = rid
	e.IPAddress = ip
	e.UserAgent = ua
	e.StackTrace = stack
	e.ResolvedAt = resolvedAt
	e.ResolvedBy = rb
	e.AdminNote = note
	if len(metaBytes) > 0 {
		_ = json.Unmarshal(metaBytes, &e.Meta)
	}
	return e, nil
}

func (r *applicationErrorLogRepo) GetByID(ctx context.Context, id string) (domain.ApplicationErrorLog, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, created_at, error_type, error_code, message, http_status, method, path, query_string,
			user_id::text, user_role, request_id, ip_address, user_agent, stack_trace, meta,
			resolved_at, resolved_by::text, admin_note
		FROM application_error_logs WHERE id = $1::uuid
	`, strings.TrimSpace(id))
	e, err := scanApplicationErrorLog(row)
	return e, err
}

func (r *applicationErrorLogRepo) MarkResolved(ctx context.Context, id, adminUserID string, note *string) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE application_error_logs
		SET resolved_at = NOW(), resolved_by = $2::uuid, admin_note = COALESCE($3, admin_note)
		WHERE id = $1::uuid AND resolved_at IS NULL
	`, strings.TrimSpace(id), strings.TrimSpace(adminUserID), note)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *applicationErrorLogRepo) MarkUnresolved(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE application_error_logs
		SET resolved_at = NULL, resolved_by = NULL, admin_note = NULL
		WHERE id = $1::uuid
	`, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *applicationErrorLogRepo) UpdateAdminNote(ctx context.Context, id string, note string) error {
	ct, err := r.pool.Exec(ctx, `
		UPDATE application_error_logs SET admin_note = $2 WHERE id = $1::uuid
	`, strings.TrimSpace(id), note)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *applicationErrorLogRepo) AnalyticsSummary(ctx context.Context, from, to time.Time) (ApplicationErrorAnalytics, error) {
	var out ApplicationErrorAnalytics
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::bigint, COUNT(*) FILTER (WHERE resolved_at IS NULL)::bigint
		FROM application_error_logs
		WHERE created_at >= $1 AND created_at < $2
	`, from, to).Scan(&out.Total, &out.Unresolved)
	if err != nil {
		return out, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT error_type, COUNT(*)::bigint FROM application_error_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY error_type ORDER BY COUNT(*) DESC
	`, from, to)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		var c int64
		if err := rows.Scan(&t, &c); err != nil {
			return out, err
		}
		out.ByType = append(out.ByType, ApplicationErrorTypeCount{ErrorType: t, Count: c})
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	rows2, err := r.pool.Query(ctx, `
		SELECT http_status, COUNT(*)::bigint FROM application_error_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY http_status ORDER BY COUNT(*) DESC
	`, from, to)
	if err != nil {
		return out, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var s int
		var c int64
		if err := rows2.Scan(&s, &c); err != nil {
			return out, err
		}
		out.ByHTTPStatus = append(out.ByHTTPStatus, ApplicationErrorStatusCount{HTTPStatus: s, Count: c})
	}
	if err := rows2.Err(); err != nil {
		return out, err
	}
	rows3, err := r.pool.Query(ctx, `
		SELECT to_char(date_trunc('day', created_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS d, COUNT(*)::bigint
		FROM application_error_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY 1 ORDER BY 1
	`, from, to)
	if err != nil {
		return out, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var d string
		var c int64
		if err := rows3.Scan(&d, &c); err != nil {
			return out, err
		}
		out.ByDay = append(out.ByDay, ApplicationErrorDayCount{Day: d, Count: c})
	}
	return out, rows3.Err()
}
