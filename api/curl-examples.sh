#!/usr/bin/env bash
# Fansedu API - Contoh curl untuk semua endpoint
# Pakai: export BASE_URL=http://localhost:8080
# Login dulu, lalu: export TOKEN="<token-dari-response-login>"

BASE_URL="${BASE_URL:-http://localhost:8080}"
V1="${BASE_URL}/api/v1"

echo "=== Health ==="
curl -s -X GET "$V1/health" | jq .

echo -e "\n=== Auth: Register ==="
curl -s -X POST "$V1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"StrongP@ssw0rd"}' | jq .

echo -e "\n=== Auth: Login (simpan token ke \$TOKEN) ==="
curl -s -X POST "$V1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"jane@example.com","password":"StrongP@ssw0rd"}' | jq .
# Set token manual: export TOKEN="<token-dari-output-di-atas>"

echo -e "\n=== Auth: Login Admin ==="
curl -s -X POST "$V1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"me.rusfandi@gmail.com","password":"mr@Condong1105"}' | jq .

echo -e "\n=== Auth: Logout (butuh Bearer) ==="
curl -s -X POST "$V1/auth/logout" -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Auth: Forgot Password ==="
curl -s -X POST "$V1/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d '{"email":"jane@example.com"}' | jq .

echo -e "\n=== Auth: Reset Password ==="
curl -s -X POST "$V1/auth/reset-password" \
  -H "Content-Type: application/json" \
  -d '{"token":"reset-token","new_password":"NewPass123"}' | jq .

echo -e "\n=== Tryouts: List Open ==="
curl -s -X GET "$V1/tryouts/open" | jq .

echo -e "\n=== Tryouts: Get by ID ==="
curl -s -X GET "$V1/tryouts/TRYYOUT-UUID-DISINI" | jq .

echo -e "\n=== Tryouts: Start (butuh Bearer) ==="
curl -s -X POST "$V1/tryouts/TRYYOUT-UUID-DISINI/start" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Attempts: Get Questions (butuh Bearer) ==="
curl -s -X GET "$V1/attempts/ATTEMPT-UUID-DISINI/questions" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Attempts: Put Answer (butuh Bearer) ==="
curl -s -X PUT "$V1/attempts/ATTEMPT-UUID/answers/QUESTION-UUID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selected_option":"B","is_marked":true}' | jq .

echo -e "\n=== Attempts: Submit (butuh Bearer) ==="
curl -s -X POST "$V1/attempts/ATTEMPT-UUID-DISINI/submit" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Student: Dashboard (butuh Bearer) ==="
curl -s -X GET "$V1/student/dashboard" -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Student: My Attempts ==="
curl -s -X GET "$V1/student/attempts" -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Student: Attempt Detail ==="
curl -s -X GET "$V1/student/attempts/ATTEMPT-UUID-DISINI" -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Student: Certificates ==="
curl -s -X GET "$V1/student/certificates" -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Courses: List ==="
curl -s -X GET "$V1/courses/" | jq .

echo -e "\n=== Courses: Enroll (butuh Bearer) ==="
curl -s -X POST "$V1/courses/COURSE-UUID-DISINI/enroll" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Admin: Overview (Bearer admin) ==="
curl -s -X GET "$V1/admin/overview" -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Admin: Create Tryout (Bearer admin) ==="
curl -s -X POST "$V1/admin/tryouts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Simulasi OSN 2025",
    "short_title":"OSN-1",
    "description":"Latihan OSN",
    "duration_minutes":90,
    "questions_count":25,
    "level":"medium",
    "opens_at":"2025-06-01T00:00:00Z",
    "closes_at":"2025-06-02T23:59:59Z",
    "max_participants":200,
    "status":"open"
  }' | jq .

echo -e "\n=== Admin: Update Tryout (Bearer admin) ==="
curl -s -X PUT "$V1/admin/tryouts/TRYYOUT-UUID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title","duration_minutes":90,"questions_count":25,"level":"medium","opens_at":"2025-06-01T00:00:00Z","closes_at":"2025-06-02T23:59:59Z","status":"open"}' | jq .

echo -e "\n=== Admin: Delete Tryout (Bearer admin) ==="
curl -s -X DELETE "$V1/admin/tryouts/TRYYOUT-UUID" -H "Authorization: Bearer $TOKEN" -w "\nHTTP %{http_code}\n"

echo -e "\n=== Admin: Create Question (Bearer admin) ==="
curl -s -X POST "$V1/admin/tryouts/TRYYOUT-UUID/questions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sort_order":1,
    "type":"multiple_choice",
    "body":"Kompleksitas binary search?",
    "options":["O(1)","O(log n)","O(n)","O(n log n)"],
    "max_score":1
  }' | jq .

echo -e "\n=== Admin: Update Question (Bearer admin) ==="
curl -s -X PUT "$V1/admin/questions/QUESTION-UUID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"body":"Pertanyaan diupdate","max_score":2}' | jq .

echo -e "\n=== Admin: Delete Question (Bearer admin) ==="
curl -s -X DELETE "$V1/admin/questions/QUESTION-UUID" -H "Authorization: Bearer $TOKEN" -w "\nHTTP %{http_code}\n"

echo -e "\n=== Admin: Create Course (Bearer admin) ==="
curl -s -X POST "$V1/admin/courses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Pembinaan OSN Informatika","description":"Kelas intensif OSN"}' | jq .

echo -e "\n=== Admin: List Enrollments (Bearer admin) ==="
curl -s -X GET "$V1/admin/courses/COURSE-UUID/enrollments" -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n=== Admin: Issue Certificate (Bearer admin) ==="
curl -s -X POST "$V1/admin/certificates" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"USER-UUID","tryout_session_id":"TRYYOUT-UUID"}' | jq .
