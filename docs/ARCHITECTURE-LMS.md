# LMS Backend Architecture (BuildWithAngga-style)

Scalable REST API for an LMS with **frictionless checkout**: users can purchase courses without logging in first. The system creates or attaches the user when they enter email/name at checkout.

---

## Tech Stack

| Layer        | Choice                |
|-------------|------------------------|
| API         | REST (Go + Chi)        |
| Auth        | JWT                    |
| Payment     | Midtrans / Stripe      |
| Database    | PostgreSQL             |
| Roles       | student, instructor, admin |

---

## User Flow (Checkout)

```
1. User visits landing page
2. User opens course detail (GET /courses/{slug})
3. User clicks "Buy Course"
4. User enters email + name (no password)
5. System:
   - If email exists → attach order to existing account
   - If email not exists → create account (password nullable), send magic-link for later login
6. System creates pending order
7. User selects payment method
8. Payment gateway session created → return payment URL
9. User completes payment on gateway
10. Webhook confirms payment → mark order paid → create course enrollments
11. Course access granted automatically
```

---

## Database Models

### Users

| Column           | Type        | Notes                          |
|------------------|------------|---------------------------------|
| id               | UUID PK    |                                 |
| name             | VARCHAR    |                                 |
| email            | VARCHAR    | UNIQUE                          |
| password_hash    | VARCHAR    | **Nullable** (auto-created at checkout) |
| role             | ENUM       | student, instructor, admin      |
| email_verified_at| TIMESTAMPTZ| Set when magic-link used       |
| created_at       | TIMESTAMPTZ|                                 |
| updated_at       | TIMESTAMPTZ|                                 |

### Courses

| Column       | Type        | Notes                |
|-------------|-------------|----------------------|
| id          | UUID PK     |                      |
| title       | VARCHAR     |                      |
| slug        | VARCHAR     | UNIQUE, for URL     |
| price       | INTEGER     | Price in cents       |
| instructor_id| UUID FK    | → users.id           |
| thumbnail   | VARCHAR     | URL                  |
| description | TEXT        |                      |
| created_at  | TIMESTAMPTZ |                      |
| updated_at  | TIMESTAMPTZ |                      |

### Orders

| Column            | Type        | Notes                    |
|-------------------|------------|---------------------------|
| id                | UUID PK    |                           |
| user_id           | UUID FK    | → users.id                |
| status            | ENUM       | pending, paid, failed      |
| total_price       | INTEGER    | Cents                     |
| payment_method    | VARCHAR    | midtrans, stripe, etc.    |
| payment_reference | VARCHAR    | Order ID for gateway      |
| created_at        | TIMESTAMPTZ|                           |

### OrderItems

| Column   | Type     | Notes      |
|----------|----------|------------|
| id       | UUID PK  |            |
| order_id | UUID FK  | → orders   |
| course_id| UUID FK  | → courses  |
| price    | INTEGER  | Cents      |

### CourseEnrollments

| Column     | Type      | Notes   |
|------------|-----------|---------|
| id         | UUID PK   |         |
| user_id    | UUID FK   | → users |
| course_id  | UUID FK   | → courses |
| enrolled_at| TIMESTAMPTZ |       |

(Existing `course_enrollments` table is used.)

### Payments (gateway record)

| Column         | Type      | Notes                    |
|----------------|-----------|--------------------------|
| id             | UUID PK   |                          |
| order_id       | UUID FK   | → orders                 |
| gateway        | VARCHAR   | midtrans, stripe         |
| status         | VARCHAR   | pending, paid, failed    |
| transaction_id | VARCHAR   | Gateway transaction ID   |
| paid_at        | TIMESTAMPTZ |                        |

---

## API Endpoints

### Public

| Method | Path             | Description        |
|--------|------------------|--------------------|
| GET    | /api/v1/courses  | List courses       |
| GET    | /api/v1/courses/slug/{slug} | Course by slug |

### Checkout (no auth required for initiate)

| Method | Path                          | Description                    |
|--------|-------------------------------|--------------------------------|
| POST   | /api/v1/checkout/initiate     | Validate course, create/find user, create order; return order_id |
| POST   | /api/v1/checkout/payment-session | Create gateway session; return payment URL / client_key |
| POST   | /api/v1/webhook/payment       | Gateway webhook (Midtrans/Stripe) |

### Auth

| Method | Path                  | Description        |
|--------|-----------------------|--------------------|
| POST   | /api/v1/auth/login    | Email + password   |
| POST   | /api/v1/auth/register | Full registration  |
| POST   | /api/v1/auth/magic-login | Request magic link (email) |
| GET    | /api/v1/auth/magic-login/verify?token= | Verify token, set session |

### Student (JWT required)

| Method | Path                    | Description           |
|--------|-------------------------|-----------------------|
| GET    | /api/v1/my-courses      | Enrolled courses      |
| GET    | /api/v1/courses/{id}/lessons | Course lessons/content |

---

## Business Logic

### Checkout Initiate

**Input:** `{ "course_slug": "...", "email": "...", "name": "..." }`

1. Validate course exists and is purchasable (slug).
2. Find user by email:
   - **Exists:** use `user_id`.
   - **Not exists:** create user (name, email, password_hash = NULL, role = student).
3. Create order: status = pending, user_id, total_price = course.price.
4. Create order_item: order_id, course_id, price.
5. Return `{ "order_id": "...", "user_id": "...", "total_price": ..., "is_new_user": bool }`.

### Payment Session

**Input:** `{ "order_id": "...", "payment_method": "midtrans" | "stripe" }`

1. Load order (must be pending).
2. Create transaction at gateway (Midtrans Snap / Stripe Checkout).
3. Store payment row: order_id, gateway, transaction_id, status = pending.
4. Return `{ "payment_url": "..." }` or `{ "client_key": "...", "snap_token": "..." }` for Snap.

### Payment Webhook

1. **Verify signature** (Midtrans/Stripe secret).
2. **Idempotency:** if order already paid, return 200 and skip.
3. Update payment: status = paid, paid_at = now.
4. Update order: status = paid.
5. For each order_item: create course_enrollment (user_id, course_id) if not exists.
6. (Optional) Send email: receipt + magic-link for new users.

---

## Security

- **Webhook verification:** Validate signature using gateway secret; reject invalid requests.
- **Idempotent payment:** Before creating enrollments, check order.status; if already paid, skip.
- **Email verification:** Auto-created accounts get magic-link email; password set on first login or via magic-link.
- **Rate limit:** Apply to checkout/initiate and auth endpoints to prevent abuse.

---

## Optional Features

- **Coupons:** `coupon_code` on order; validate and apply discount before creating payment session.
- **Bundles:** Multiple course_ids in initiate; order_items for each.
- **Affiliate:** `affiliate_id` or ref on order; track in order or separate table.
- **Invoices:** Generate PDF from order + payment; store URL or send email.

---

## File Layout (Implementation)

```
internal/
  domain/         order.go, order_item.go, course.go (slug, price, ...)
  repo/           order_repo.go, order_item_repo.go, payment_gateway.go (interface)
  service/        checkout_service.go (Initiate, CreatePaymentSession, HandleWebhook)
  app/http/
    handlers/     checkout_handlers.go, webhook_handlers.go
    dto/          checkout.go
  payment/        midtrans.go, stripe.go (optional adapter)
```

---

## Migration Summary

- `users.password_hash` → nullable.
- `courses`: add slug (UNIQUE), price_cents, thumbnail; instructor_id = created_by or new column.
- New: `orders`, `order_items`.
- `payments`: add order_id (FK), gateway, transaction_id (or new table `order_payments`).
- Role enum: add `instructor` if not present.

This document is the single source of truth for the frictionless checkout LMS flow.
