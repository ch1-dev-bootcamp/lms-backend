# LMS Backend API - Postman Collection

This repository contains a comprehensive Postman collection for testing the Learning Management System (LMS) Backend API built with Go.

## 📁 Files Included

- `LMS-Backend-API.postman_collection.json` - Complete API collection with 15 endpoints
- `LMS-Backend-Environment.postman_environment.json` - Environment variables for testing
- `POSTMAN_COLLECTION_README.md` - This documentation file

## 🚀 Quick Start

### 1. Import Collection and Environment

1. Open Postman
2. Click **Import** button
3. Import both files:
   - `LMS-Backend-API.postman_collection.json`
   - `LMS-Backend-Environment.postman_environment.json`

### 2. Set Environment

1. Select the **LMS Backend Environment** from the environment dropdown
2. Ensure the `base_url` is set to `http://localhost:8080` (or your server URL)

### 3. Start the LMS Backend

```bash
docker-compose up --build -d
```

## 📋 API Endpoints Overview

### Health Checks (2 endpoints)
- `GET /health` - Application health check
- `GET /health/database` - Database health check

### API Root (1 endpoint)
- `GET /api/v1/` - API information

### Authentication (3 endpoints)
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

### Users (2 endpoints)
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile

### Courses (5 endpoints)
- `GET /api/v1/courses/` - List courses (with pagination)
- `POST /api/v1/courses/` - Create course
- `GET /api/v1/courses/:id` - Get course details
- `PUT /api/v1/courses/:id` - Update course
- `DELETE /api/v1/courses/:id` - Delete course

### Lessons (5 endpoints)
- `GET /api/v1/lessons/` - List lessons (with pagination)
- `POST /api/v1/lessons/` - Create lesson
- `GET /api/v1/lessons/:id` - Get lesson details
- `PUT /api/v1/lessons/:id` - Update lesson
- `DELETE /api/v1/lessons/:id` - Delete lesson

### Enrollments (3 endpoints)
- `POST /api/v1/enrollments/` - Enroll in course
- `GET /api/v1/enrollments/` - List enrollments (with pagination)
- `GET /api/v1/enrollments/:id` - Get enrollment details

### Progress (3 endpoints)
- `POST /api/v1/progress/complete` - Complete lesson
- `GET /api/v1/progress/` - Get progress
- `GET /api/v1/progress/:user_id` - Get user progress

### Certificates (3 endpoints)
- `GET /api/v1/certificates/` - List certificates (with pagination)
- `GET /api/v1/certificates/:id` - Get certificate details
- `GET /api/v1/certificates/verify/:id` - Verify certificate

### Validation Tests (3 endpoints)
- Invalid registration test
- Invalid course creation test
- Invalid pagination test

## 🔧 Environment Variables

### Base Configuration
- `base_url` - API base URL (default: http://localhost:8080)
- `auth_token` - JWT authentication token (auto-populated after login)

### Test Data
- `test_email` - Test user email
- `test_password` - Test user password
- `test_name` - Test user name
- `test_role` - Test user role

### Entity IDs
- `user_id` - Sample user UUID
- `course_id` - Sample course UUID
- `lesson_id` - Sample lesson UUID
- `enrollment_id` - Sample enrollment UUID
- `certificate_id` - Sample certificate UUID
- `instructor_id` - Sample instructor UUID

### Content Variables
- `course_title` - Sample course title
- `course_description` - Sample course description
- `lesson_title` - Sample lesson title
- `lesson_content` - Sample lesson content
- `page` - Pagination page number
- `page_size` - Pagination page size

## 🧪 Testing Workflow

### 1. Health Check
Start by testing the health endpoints to ensure the API is running.

### 2. Authentication Flow
1. Register a new user
2. Login with the registered credentials
3. Copy the `auth_token` from the login response
4. Set it in the environment variables

### 3. CRUD Operations
Test the full CRUD operations for courses and lessons:
1. Create entities
2. List entities with pagination
3. Get specific entities
4. Update entities
5. Delete entities

### 4. Business Logic
Test enrollment and progress tracking:
1. Enroll in a course
2. Complete lessons
3. Track progress
4. View certificates

### 5. Validation Testing
Test the validation system with invalid data to ensure proper error handling.

## 🔍 Request Validation

All POST and PUT endpoints include comprehensive validation:

### Email Validation
- Must be a valid email format
- Required for registration and login

### Password Validation
- Minimum 8 characters
- Required for registration and login

### UUID Validation
- Must be valid UUID format
- Required for entity IDs

### String Length Validation
- Minimum and maximum length constraints
- Applied to titles, descriptions, and content

### Enum Validation
- Role must be: admin, instructor, or student
- Status must be: draft, published, or archived

### Pagination Validation
- Page must be >= 1
- Page size must be between 1 and 100

## 📊 Response Formats

### Success Responses
```json
{
  "message": "Operation successful",
  "data": { ... },
  "status": "success"
}
```

### Error Responses
```json
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": [
    {
      "field": "email",
      "message": "Must be a valid email address",
      "value": "invalid-email"
    }
  ]
}
```

### Paginated Responses
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

## 🐛 Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure the LMS backend is running: `docker-compose up -d`
   - Check if port 8080 is available

2. **Authentication Errors**
   - Make sure to login first and copy the `auth_token`
   - Set the token in the environment variables

3. **Validation Errors**
   - Check the request body format
   - Ensure all required fields are provided
   - Verify data types and formats

4. **Environment Variables Not Working**
   - Ensure the correct environment is selected
   - Check that variable names match exactly

### Debug Tips

1. Check the **Console** tab in Postman for detailed error messages
2. Use the **Pre-request Script** tab to log variables
3. Enable **Request/Response logging** in Postman settings
4. Check the LMS backend logs: `docker-compose logs lms-backend`

## 📝 Notes

- All timestamps are in RFC3339 format
- UUIDs are generated automatically for new entities
- Pagination is 1-indexed (page 1, not 0)
- All endpoints return JSON responses
- CORS is enabled for cross-origin requests

## 🔄 Updates

This collection is maintained alongside the LMS Backend API. When new endpoints are added or existing ones are modified, this collection will be updated accordingly.

---

**Happy Testing! 🚀**
