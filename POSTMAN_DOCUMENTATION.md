# LMS Backend API - Complete Postman Documentation

## 📚 Overview

This document provides comprehensive documentation for the Learning Management System (LMS) Backend API Postman collection. The collection includes **50+ endpoints** covering all aspects of the LMS including authentication, user management, courses, lessons, enrollments, progress tracking, prerequisites, and certificates.

## 📁 Collection Files

- **`LMS-Backend-Complete.postman_collection.json`** - Complete API collection with all endpoints
- **`LMS-Backend-Complete.postman_environment.json`** - Environment variables for testing
- **`POSTMAN_DOCUMENTATION.md`** - This comprehensive documentation

## 🚀 Quick Start Guide

### 1. Import Collection and Environment

1. Open Postman
2. Click the **Import** button
3. Import both files:
   - `LMS-Backend-Complete.postman_collection.json`
   - `LMS-Backend-Complete.postman_environment.json`

### 2. Configure Environment

1. Select **LMS Backend Complete Environment** from the environment dropdown
2. Verify the `base_url` is set to `http://localhost:8080` (or your server URL)
3. Review and adjust test data variables as needed

### 3. Start the Backend

```bash
# Using Docker Compose (Recommended)
docker-compose up --build -d

# Or run locally
go run cmd/main.go
```

## 📋 API Endpoints Reference

### 🏥 Health Checks (2 endpoints)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Application health check | No |
| GET | `/health/database` | Database connectivity check | No |

### ℹ️ API Information (1 endpoint)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/` | API version and information | No |

### 🔐 Authentication (3 endpoints)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/auth/register` | Register new user | No |
| POST | `/api/v1/auth/login` | User login | No |
| POST | `/api/v1/auth/logout` | User logout | No |

### 👤 User Management (3 endpoints)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users/profile` | Get current user profile | Yes |
| PUT | `/api/v1/users/profile` | Update user profile | Yes |
| DELETE | `/api/v1/users/:id` | Delete user | Yes |

### 📚 Courses - Public Access (3 endpoints)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/courses` | List all courses | No |
| GET | `/api/v1/courses/search` | Search courses | No |
| GET | `/api/v1/courses/:id` | Get course details | No |

### 🎓 Courses - Management (4 endpoints)

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| POST | `/api/v1/courses` | Create course | Yes | Instructor |
| PUT | `/api/v1/courses/:id` | Update course | Yes | Instructor |
| DELETE | `/api/v1/courses/:id` | Delete course | Yes | Instructor |
| GET | `/api/v1/courses/my-courses` | List instructor's courses | Yes | Instructor |

### 📖 Lessons - Public Access (3 endpoints)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/lessons` | List all lessons | No |
| GET | `/api/v1/lessons/:id` | Get lesson details | No |
| GET | `/api/v1/courses/:course_id/lessons` | List course lessons | No |

### 📝 Lessons - Management (7 endpoints)

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| POST | `/api/v1/lessons` | Create lesson | Yes | Instructor |
| PUT | `/api/v1/lessons/:id` | Update lesson | Yes | Instructor |
| DELETE | `/api/v1/lessons/:id` | Delete lesson | Yes | Instructor |
| POST | `/api/v1/courses/:course_id/lessons` | Create course lesson | Yes | Instructor |
| PUT | `/api/v1/courses/:course_id/lessons/:id` | Update course lesson | Yes | Instructor |
| DELETE | `/api/v1/courses/:course_id/lessons/:id` | Delete course lesson | Yes | Instructor |
| PUT | `/api/v1/courses/:course_id/lessons/reorder` | Reorder lessons | Yes | Instructor |

### 📋 Enrollments (7 endpoints)

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| POST | `/api/v1/enrollments` | Enroll in course | Yes | - |
| GET | `/api/v1/enrollments/my-enrollments` | Get user enrollments | Yes | - |
| DELETE | `/api/v1/enrollments/courses/:course_id` | Unenroll from course | Yes | - |
| GET | `/api/v1/enrollments` | List all enrollments | Yes | Admin |
| GET | `/api/v1/enrollments/:id` | Get enrollment details | Yes | Admin |
| PUT | `/api/v1/enrollments/:id` | Update enrollment | Yes | Admin |
| DELETE | `/api/v1/enrollments/:id` | Delete enrollment | Yes | Admin |

### 📊 Progress Tracking (3 endpoints)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/progress/complete` | Mark lesson complete | Yes |
| GET | `/api/v1/progress/my-progress` | Get user progress | Yes |
| GET | `/api/v1/progress/courses/:course_id` | Get course progress | Yes |

### ✅ Course Completions (2 endpoints)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/completions` | List user completions | Yes |
| GET | `/api/v1/completions/courses/:course_id` | Get course completion | Yes |

### 🔗 Prerequisites (4 endpoints)

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| POST | `/api/v1/prerequisites` | Create prerequisite | Yes | Instructor |
| GET | `/api/v1/prerequisites/courses/:course_id` | List prerequisites | Yes | Instructor |
| DELETE | `/api/v1/prerequisites/courses/:course_id/required/:required_course_id` | Delete prerequisite | Yes | Instructor |
| GET | `/api/v1/prerequisites/check/courses/:course_id` | Check prerequisites | Yes | - |

### 🏆 Certificates (4 endpoints)

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| GET | `/api/v1/certificates` | List certificates | No | - |
| GET | `/api/v1/certificates/:id` | Get certificate | No | - |
| GET | `/api/v1/certificates/verify/:code` | Verify certificate | No | - |
| POST | `/api/v1/certificates` | Create certificate | Yes | Admin |

### 🧪 Validation Tests (5 endpoints)

Test endpoints for validation and error handling scenarios.

## 🔧 Environment Variables

### Core Configuration

| Variable | Default Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | API base URL |
| `auth_token` | *(auto-populated)* | JWT token for authenticated requests |
| `student_token` | *(auto-populated)* | JWT token for student role |
| `instructor_token` | *(auto-populated)* | JWT token for instructor role |
| `admin_token` | *(auto-populated)* | JWT token for admin role |

### Test User Accounts

| Variable | Default Value | Description |
|----------|--------------|-------------|
| `test_email` | `john.doe@example.com` | Default test user email |
| `test_password` | `SecurePass123!` | Default test user password |
| `test_name` | `John Doe` | Default test user name |
| `test_role` | `student` | Default test user role |
| `instructor_email` | `instructor@example.com` | Instructor account email |
| `instructor_password` | `InstructorPass123!` | Instructor account password |
| `admin_email` | `admin@example.com` | Admin account email |
| `admin_password` | `AdminPass123!` | Admin account password |

### Entity IDs

| Variable | Description |
|----------|-------------|
| `user_id` | Sample user UUID |
| `instructor_id` | Sample instructor UUID |
| `course_id` | Sample course UUID |
| `required_course_id` | Prerequisite course UUID |
| `lesson_id` | Sample lesson UUID |
| `lesson_id_1` | First lesson for reordering |
| `lesson_id_2` | Second lesson for reordering |
| `enrollment_id` | Sample enrollment UUID |
| `certificate_id` | Sample certificate UUID |
| `certificate_code` | Certificate verification code |

### Content Variables

| Variable | Description |
|----------|-------------|
| `course_title` | Sample course title |
| `course_description` | Sample course description |
| `lesson_title` | Sample lesson title |
| `lesson_content` | Sample lesson content |
| `search_query` | Search query string |
| `page` | Pagination page number (1-indexed) |
| `page_size` | Items per page (1-100) |

## 🧪 Testing Workflows

### 1. Basic Setup Flow

```
1. Health Check → Verify API is running
2. Database Health → Verify database connection
3. API Root → Confirm API version
```

### 2. User Registration & Authentication Flow

```
1. Register User → Create new account
2. Login → Get JWT token (auto-saved)
3. Get Profile → Verify authentication
4. Update Profile → Test profile updates
```

### 3. Course Management Flow (Instructor)

```
1. Login as Instructor
2. Create Course → Save course_id
3. Update Course → Modify course details
4. List My Courses → View instructor's courses
5. Delete Course → Clean up
```

### 4. Lesson Management Flow

```
1. Create Lesson → Add to course
2. List Course Lessons → Verify creation
3. Update Lesson → Modify content
4. Reorder Lessons → Change sequence
5. Delete Lesson → Remove from course
```

### 5. Student Enrollment Flow

```
1. Login as Student
2. List Courses → Browse available
3. Check Prerequisites → Verify eligibility
4. Enroll in Course → Join course
5. List My Enrollments → View enrolled courses
```

### 6. Progress Tracking Flow

```
1. List Course Lessons
2. Complete Lesson → Mark as done
3. Get Course Progress → Check completion %
4. Get My Progress → Overall progress
5. Get Course Completion → Verify completion
```

### 7. Certificate Flow

```
1. Complete all course lessons
2. Admin creates certificate
3. List Certificates → View all
4. Get Certificate → View details
5. Verify Certificate → Validate by code
```

## 📝 Request/Response Examples

### Successful Registration Response

```json
{
    "message": "User registered successfully",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "success"
}
```

### Login Response with Token

```json
{
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "success"
}
```

### Paginated Response

```json
{
    "data": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "title": "Introduction to Go",
            "description": "Learn Go programming",
            "status": "published"
        }
    ],
    "pagination": {
        "page": 1,
        "page_size": 10,
        "total": 50,
        "total_pages": 5
    }
}
```

### Validation Error Response

```json
{
    "error": "validation_failed",
    "message": "Request validation failed",
    "details": [
        {
            "field": "email",
            "message": "Must be a valid email address",
            "value": "invalid-email"
        },
        {
            "field": "password",
            "message": "Must be at least 8 characters",
            "value": "short"
        }
    ]
}
```

## 🔒 Authentication & Authorization

### JWT Token Usage

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

### Role-Based Access Control

| Role | Access Level |
|------|-------------|
| **Admin** | Full system access, user management, certificate creation |
| **Instructor** | Course/lesson management, view enrollments, prerequisites |
| **Student** | Enroll in courses, track progress, view certificates |

### Role Hierarchy

```
Admin > Instructor > Student
```

Higher roles inherit permissions from lower roles.

## ✅ Validation Rules

### User Registration

- **Email**: Valid email format, required
- **Password**: Minimum 8 characters, required
- **Name**: 2-100 characters, required
- **Role**: Must be `admin`, `instructor`, or `student`

### Course Creation

- **Title**: 3-200 characters, required
- **Description**: Maximum 1000 characters, optional
- **Status**: `draft`, `published`, or `archived`

### Lesson Creation

- **Course ID**: Valid UUID, required
- **Title**: 3-200 characters, required
- **Content**: Maximum 5000 characters, optional
- **Order Number**: Minimum 1, optional
- **Duration**: 1-480 minutes, optional

### Pagination

- **Page**: Minimum 1, required
- **Page Size**: 1-100, required

## 🐛 Troubleshooting

### Common Issues & Solutions

#### 1. Connection Refused

**Problem**: Cannot connect to API
**Solution**:
- Verify backend is running: `docker-compose ps`
- Check port 8080 is not in use
- Ensure firewall allows connection

#### 2. Authentication Failed

**Problem**: 401 Unauthorized error
**Solution**:
- Login first to get JWT token
- Check token hasn't expired
- Verify Authorization header format: `Bearer <token>`

#### 3. Insufficient Permissions

**Problem**: 403 Forbidden error
**Solution**:
- Verify user role meets requirements
- Use appropriate account (student/instructor/admin)
- Check role-based access requirements

#### 4. Validation Errors

**Problem**: 400 Bad Request with validation errors
**Solution**:
- Review error details in response
- Check required fields are provided
- Verify data types and formats
- Ensure values meet validation constraints

#### 5. Not Found Errors

**Problem**: 404 Not Found
**Solution**:
- Verify entity IDs are correct
- Check if resource exists
- Ensure proper URL path

### Debug Tips

1. **Enable Postman Console** (View → Show Postman Console)
2. **Check Response Headers** for additional error information
3. **Review Request Body** in Postman's request builder
4. **Monitor Backend Logs**: `docker-compose logs -f lms-backend`
5. **Use Pre-request Scripts** for dynamic data
6. **Set Up Tests** for automated validation

## 🔄 Best Practices

### 1. Testing Order

- Start with health checks
- Register and login users first
- Create entities before testing operations
- Clean up test data after testing

### 2. Environment Management

- Use different environments for dev/staging/prod
- Keep sensitive data in environment variables
- Regularly update test data
- Document custom variables

### 3. Error Handling

- Always check response status codes
- Read error messages and details
- Handle both success and failure cases
- Implement retry logic for transient errors

### 4. Performance Testing

- Use realistic data volumes
- Test pagination with large datasets
- Monitor response times
- Test concurrent requests

## 📊 Response Status Codes

| Code | Status | Description |
|------|--------|-------------|
| 200 | OK | Successful request |
| 201 | Created | Resource created successfully |
| 204 | No Content | Successful deletion |
| 400 | Bad Request | Validation error or invalid request |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

## 🚀 Advanced Features

### Automated Testing

The collection includes:
- Pre-request scripts for data setup
- Test scripts for response validation
- Environment variable auto-population
- Chain requests with data passing

### Collection Runner

Use Postman's Collection Runner for:
- Bulk testing all endpoints
- Performance testing
- Data-driven testing with CSV/JSON
- Generating test reports

### Newman CLI

Run tests from command line:

```bash
# Install Newman
npm install -g newman

# Run collection
newman run LMS-Backend-Complete.postman_collection.json \
  -e LMS-Backend-Complete.postman_environment.json

# Generate HTML report
newman run LMS-Backend-Complete.postman_collection.json \
  -e LMS-Backend-Complete.postman_environment.json \
  -r html --reporter-html-export report.html
```

## 📚 Additional Resources

- [Postman Documentation](https://learning.postman.com/docs/)
- [JWT.io](https://jwt.io/) - JWT token decoder
- [UUID Generator](https://www.uuidgenerator.net/) - Generate test UUIDs
- [Go Documentation](https://golang.org/doc/) - Go language reference

## 🤝 Contributing

When adding new endpoints:

1. Add endpoint to appropriate folder in collection
2. Include request description
3. Add example request/response
4. Update environment variables if needed
5. Document in this README
6. Test thoroughly before committing

## 📄 License

This Postman collection is part of the LMS Backend project and follows the same license terms.

---

**Last Updated**: January 2024
**Version**: 1.0.0
**Maintainer**: LMS Backend Team

For issues or questions, please refer to the main project repository.
