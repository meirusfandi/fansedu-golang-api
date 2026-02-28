package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authReq struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResp struct {
	User  map[string]any `json:"user"`
	Token string         `json:"token"`
}

func AuthRegister(_ *pgxpool.Pool, jwtSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// MVP stub: wire to service+repo later
		token, _ := signJWT(jwtSecret, "user-1", "student")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(authResp{
			User:  map[string]any{"id": "user-1", "name": req.Name, "email": req.Email, "role": "student"},
			Token: token,
		})
	}
}

func AuthLogin(_ *pgxpool.Pool, jwtSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// MVP stub: wire to service+repo later
		token, _ := signJWT(jwtSecret, "user-1", "student")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(authResp{
			User:  map[string]any{"id": "user-1", "email": req.Email, "role": "student"},
			Token: token,
		})
	}
}

func signJWT(secret []byte, userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

