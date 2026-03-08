# FANSEDU Production Schema

Production-ready PostgreSQL schema for the FANSEDU LMS.

## Roles

| Role     | Slug     | Purpose |
|----------|----------|---------|
| Student  | student  | Register, tryouts, purchase classes, modules, quizzes, chat, forum, rankings |
| Teacher  | teacher  | Schools, school profile, purchase packages, add students, enroll students, view rankings |
| Trainer  | trainer  | Create classes, capacity, modules, materials, quizzes, question bank, analytics |
| Admin    | admin    | Full system management, payments, question bank, notifications |

## Main Tables

| Table                 | Purpose |
|-----------------------|---------|
| roles                 | System roles |
| users                 | All users (role_id, optional school_id) |
| schools               | Schools (npsn, address, telepon, etc.) |
| teacher_schools       | Teacher ↔ School (many-to-many) |
| teacher_students      | Students under a teacher |
| trainer_slots         | Paid slots per trainer (add students) |
| subjects / topics     | For question bank and topic-based analytics |
| classes               | Created by trainer; capacity, price |
| class_enrollments     | Student enrollments in classes |
| modules               | Learning content per class |
| quizzes               | Quizzes per class/module |
| questions             | Question bank (topic_id for weakness) |
| question_options      | Multiple choice options |
| quiz_questions        | Questions in a quiz (N:M) |
| quiz_attempts         | Student attempt per quiz |
| student_answers       | Per-question answers (for analytics) |
| student_scores        | Denormalized score per attempt |
| tryouts               | Tryout events |
| tryout_registrations  | Student registration for tryout |
| tryout_questions      | Questions in a tryout |
| tryout_attempts       | Tryout attempt per user |
| tryout_answer         | Answers in tryout attempt |
| orders                | Order header |
| order_items           | Line items (class, tryout, teacher_slots) |
| payments              | Payment + proof_url, confirmation |
| notifications         | User notifications |
| class_messages        | Class chat |
| discussions           | Forum threads per class |
| discussion_replies     | Replies to discussions |

## Views

- **class_ranking** — Per-class student ranking by total quiz score.
- **student_weakness_analysis** — Per-student per-topic: total_answered, correct_count, wrong_count, error_rate_pct.
- **student_quiz_scores** — Student score per quiz (score, max_score, score_pct).

## How to Run

```bash
# Create database
createdb fansedu_production

# Run schema (from repo root)
psql -d fansedu_production -f internal/db/schema_production/fansedu_production_schema.sql
```

Or use your migration runner; run the SQL in dependency order (file is already ordered).

## Scale

Designed for:

- 100,000+ students
- 10,000+ classes
- Millions of quiz answers

Indexes are added on FKs, status, and common filters (created_at, user_id, class_id, etc.).

## SQLC / GORM

- UUID primary keys throughout.
- Standard `created_at`, `updated_at` timestamps.
- Foreign keys and enums are SQLC/GORM compatible.
