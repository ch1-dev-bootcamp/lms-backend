package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/lms-backend/internal/auth"
	apperrors "github.com/your-org/lms-backend/internal/errors"
	"github.com/your-org/lms-backend/pkg/logger"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			AbortWithError(c, apperrors.NewUnauthorizedError("Authorization header is required"))
			return
		}

		// Extract token from header
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			AbortWithError(c, apperrors.NewUnauthorizedError("Invalid authorization header format"))
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			logger.WithField("error", err.Error()).Warn("Invalid JWT token")
			AbortWithError(c, apperrors.NewUnauthorizedError("Invalid or expired token"))
			return
		}

		// Store user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			AbortWithError(c, apperrors.NewUnauthorizedError("User role not found"))
			return
		}

		role, ok := userRole.(string)
		if !ok {
			AbortWithError(c, apperrors.NewUnauthorizedError("Invalid user role"))
			return
		}

		// Check if user has required role
		if !hasRequiredRole(role, requiredRole) {
			AbortWithError(c, apperrors.NewForbiddenError("Insufficient permissions"))
			return
		}

		c.Next()
	}
}

// hasRequiredRole checks if user role meets the required role
func hasRequiredRole(userRole, requiredRole string) bool {
	// Define role hierarchy (admin > instructor > student)
	roleHierarchy := map[string]int{
		"student":    1,
		"instructor": 2,
		"admin":      3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	// User must have at least the required level
	return userLevel >= requiredLevel
}

// OptionalAuth middleware validates JWT token if present but doesn't require it
func OptionalAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from header
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Log warning but continue without authentication
			logger.WithField("error", err.Error()).Warn("Invalid authorization header format")
			c.Next()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			// Log warning but continue without authentication
			logger.WithField("error", err.Error()).Warn("Invalid JWT token")
			c.Next()
			return
		}

		// Store user information in context if token is valid
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", false
	}

	return userIDStr, true
}

// GetUserRoleFromContext extracts user role from context
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := userRole.(string)
	if !ok {
		return "", false
	}

	return role, true
}
