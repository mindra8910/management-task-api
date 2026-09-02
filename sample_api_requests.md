# Sample API Requests & Responses

## Base URL
```
http://localhost:8080
```

---

## 1. Register User

**Request:**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## 2. Login

**Request:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
}
```

---

## 3. Create Task (dengan Idempotency Key)

**Request:**
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Idempotency-Key: unique-key-001" \
  -d '{
    "title": "Setup project",
    "description": "Initialize Go project with dependencies",
    "status": "pending"
  }'
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Setup project",
    "description": "Initialize Go project with dependencies",
    "status": "pending",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## 4. List Tasks (dengan Filter)

**Request:**
```bash
curl -X GET "http://localhost:8080/tasks?status=pending&page=1&limit=10" \
  -H "Authorization: Bearer <TOKEN>"
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "user_id": "550e8400-e29b-41d4-a716-446655440001",
        "title": "Setup project",
        "description": "Initialize Go project with dependencies",
        "status": "pending",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10
  }
}
```

---

## 5. Get Task Detail

**Request:**
```bash
curl -X GET http://localhost:8080/tasks/660e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer <TOKEN>"
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Setup project",
    "description": "Initialize Go project with dependencies",
    "status": "pending",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## 6. Update Task

**Request:**
```bash
curl -X PUT http://localhost:8080/tasks/660e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "title": "Setup project (updated)",
    "status": "in_progress"
  }'
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "Setup project (updated)",
    "status": "in_progress",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

---

## 7. Delete Task

**Request:**
```bash
curl -X DELETE http://localhost:8080/tasks/660e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer <TOKEN>"
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Task deleted successfully"
}
```

---

## 8. Assign Task ke User Lain

**Request:**
```bash
curl -X POST http://localhost:8080/tasks/660e8400-e29b-41d4-a716-446655440001/assign \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "assignee_id": "550e8400-e29b-41d4-a716-446655440002"
  }'
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Task assigned successfully"
}
```

---

## Sample Users untuk Testing

| Name | Email | Password |
|------|-------|----------|
| John Doe | john@example.com | password123 |
| Jane Smith | jane@example.com | password123 |
| Bob Wilson | bob@example.com | password123 |

---

## Quick Test Script

```bash
# 1. Register
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'

# 2. Login
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}' | jq -r '.data.token')

# 3. Create Task
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: key-001" \
  -d '{"title":"My Task","description":"Task description","status":"pending"}'

# 4. List Tasks
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN"
```
