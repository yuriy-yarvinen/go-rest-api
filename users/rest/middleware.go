package rest

import (
	"net/http"
	"strings"

	"go-rest-api/users"
	"go-rest-api/utils"

	"github.com/gin-gonic/gin"
)

const contextUserIDKey = "userID"

// AuthRequired requires a valid "Authorization: Bearer <token>" header. It
// checks the token's signature and expiry, then confirms the user it names
// still exists — a token stays cryptographically valid for its full TTL
// even if the account was deleted after it was issued. On success the
// authenticated user's id is stored in the context; handlers read it back
// with UserID(context).
func AuthRequired(service *users.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString, ok := strings.CutPrefix(context.GetHeader("Authorization"), "Bearer ")
		if !ok || tokenString == "" {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or malformed Authorization header"})
			return
		}

		token, err := utils.ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		userID, ok := utils.UserIDFromToken(token)
		if !ok {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		if _, err := service.GetByID(userID); err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User no longer exists"})
			return
		}

		context.Set(contextUserIDKey, userID)
		context.Next()
	}
}

// UserID returns the authenticated user's id set by AuthRequired. ok is
// false when called outside a route protected by AuthRequired.
func UserID(context *gin.Context) (int64, bool) {
	v, exists := context.Get(contextUserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
