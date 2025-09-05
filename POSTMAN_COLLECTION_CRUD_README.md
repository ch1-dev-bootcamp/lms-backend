# LMS Backend API - Complete CRUD Operations

This document describes the updated Postman collection that includes all CRUD operations for the Learning Management System backend.

## 📁 Files

- **`LMS-Backend-API-Updated.postman_collection.json`** - Complete API collection with all CRUD operations
- **`LMS-Backend-Environment-Updated.postman_environment.json`** - Environment variables for testing

## 🚀 API Endpoints Overview

### **Health Checks**
- `GET /health` - Application health check
- `GET /health/database` - Database health check

### **Authentication**
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

### **Users** (Complete CRUD)
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `DELETE /api/v1/users/:id` - Delete user

### **Courses** (Complete CRUD)
- `GET /api/v1/courses/` - List courses (paginated)
- `POST /api/v1/courses/` - Create course
- `GET /api/v1/courses/:id` - Get course details
- `PUT /api/v1/courses/:id` - Update course
- `DELETE /api/v1/courses/:id` - Delete course

### **Lessons** (Complete CRUD)
- `GET /api/v1/lessons/` - List lessons (paginated)
- `POST /api/v1/lessons/` - Create lesson
- `GET /api/v1/lessons/:id` - Get lesson details
- `PUT /api/v1/lessons/:id` - Update lesson
- `DELETE /api/v1/lessons/:id` - Delete lesson

### **Enrollments** (Complete CRUD)
- `POST /api/v1/enrollments/` - Create enrollment
- `GET /api/v1/enrollments/` - List enrollments (paginated)
- `GET /api/v1/enrollments/:id` - Get enrollment details
- `PUT /api/v1/enrollments/:id` - Update enrollment
- `DELETE /api/v1/enrollments/:id` - Delete enrollment

### **Progress** (Partial CRUD)
- `POST /api/v1/progress/complete` - Complete lesson
- `GET /api/v1/progress/` - Get progress overview
- `GET /api/v1/progress/:user_id` - Get user progress

### **Certificates** (Read Operations)
- `GET /api/v1/certificates/` - List certificates (paginated)
- `GET /api/v1/certificates/:id` - Get certificate details
- `GET /api/v1/certificates/verify/:id` - Verify certificate

## 🔧 Environment Variables

### **Base Configuration**
- `base_url` - API base URL (default: http://localhost:8080)
- `jwt_token` - JWT authentication token

### **Test Data**
- `test_email` - Test user email
- `test_password` - Test user password
- `test_name` - Test user name
- `updated_name` - Updated user name
- `updated_email` - Updated user email

### **Entity IDs** (Auto-populated from responses)
- `user_id` - User ID from registration/login
- `instructor_id` - Instructor ID for course creation
- `course_id` - Course ID from course creation
- `lesson_id` - Lesson ID from lesson creation
- `enrollment_id` - Enrollment ID from enrollment creation
- `certificate_id` - Certificate ID from certificate creation

### **Test Content**
- `course_title` - Course title for testing
- `course_description` - Course description for testing
- `lesson_title` - Lesson title for testing
- `lesson_content` - Lesson content for testing

## 📋 Testing Workflow

### **1. Setup**
1. Import the collection and environment into Postman
2. Start the LMS backend server
3. Set the `base_url` environment variable if needed

### **2. Authentication Flow**
1. **Register User** - Create a new user account
2. **Login User** - Get JWT token (save to `jwt_token` variable)
3. **Get User Profile** - Verify user data

### **3. Course Management**
1. **Create Course** - Create a new course (save `course_id`)
2. **List Courses** - View all courses
3. **Get Course** - View specific course details
4. **Update Course** - Modify course information
5. **Delete Course** - Remove course

### **4. Lesson Management**
1. **Create Lesson** - Add lesson to course (save `lesson_id`)
2. **List Lessons** - View all lessons
3. **Get Lesson** - View specific lesson details
4. **Update Lesson** - Modify lesson content
5. **Delete Lesson** - Remove lesson

### **5. Enrollment Management**
1. **Create Enrollment** - Enroll user in course (save `enrollment_id`)
2. **List Enrollments** - View all enrollments
3. **Get Enrollment** - View specific enrollment
4. **Update Enrollment** - Modify enrollment status
5. **Delete Enrollment** - Remove enrollment

### **6. Progress Tracking**
1. **Complete Lesson** - Mark lesson as completed
2. **Get Progress** - View overall progress
3. **Get User Progress** - View user-specific progress

### **7. Certificate Management**
1. **List Certificates** - View all certificates
2. **Get Certificate** - View specific certificate
3. **Verify Certificate** - Verify certificate validity

## 🧪 Validation Tests

The collection includes validation tests to verify error handling:

- **Invalid Register Request** - Test validation errors
- **Invalid Course ID** - Test invalid UUID format
- **Non-existent Course** - Test 404 error handling

## 🔍 Error Handling

All endpoints include proper error handling with:
- **Validation Errors** (400) - Invalid input data
- **Not Found Errors** (404) - Resource not found
- **Conflict Errors** (409) - Duplicate entries
- **Database Errors** (500) - Server-side errors

## 📊 Response Examples

### **Success Response**
```json
{
  "message": "Operation successful",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### **Error Response**
```json
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": [
    {
      "field": "email",
      "message": "Must be a valid email address"
    }
  ],
  "request_id": "req-123456",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 🚀 Quick Start Commands

### **Start Server**
```bash
go run cmd/main.go
```

### **Test Health**
```bash
curl.exe "http://localhost:8080/health"
```

### **Register User**
```bash
curl.exe -X POST "http://localhost:8080/api/v1/auth/register" -H "Content-Type: application/json" -d "{\"email\":\"test@example.com\",\"password\":\"password123\",\"name\":\"Test User\",\"role\":\"student\"}"
```

### **Create Course**
```bash
curl.exe -X POST "http://localhost:8080/api/v1/courses/" -H "Content-Type: application/json" -d "{\"title\":\"Test Course\",\"description\":\"A test course\",\"instructor_id\":\"USER_ID_HERE\",\"status\":\"published\"}"
```

## 📝 Notes

- All endpoints require proper authentication (JWT token in Authorization header)
- Pagination is supported for list endpoints (page, page_size parameters)
- All UUIDs are automatically generated and returned in responses
- Error responses include request IDs for debugging
- Database operations are fully implemented with proper error handling

## 🔄 Updates Made

### **New CRUD Operations Added:**
- User Delete operation
- Enrollment Update/Delete operations
- All Lesson operations with database integration
- All Enrollment operations with database integration
- Progress operations with database integration

### **Enhanced Error Handling:**
- Centralized error handling middleware
- Structured error responses with request IDs
- Proper HTTP status codes
- Database error context preservation

### **Improved Testing:**
- Complete test coverage for all operations
- Validation test scenarios
- Environment variable management
- Response data extraction for chaining requests
