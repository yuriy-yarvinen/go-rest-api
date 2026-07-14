package rest

import (
	"net/http"
	"strings"

	"go-rest-api/authctx"
	"go-rest-api/users"
	"go-rest-api/utils"

	"github.com/gin-gonic/gin"
)

// AuthRequired requires a valid "Authorization: Bearer <token>" header. It
// checks the token's signature and expiry, then confirms the user it names
// still exists — a token stays cryptographically valid for its full TTL
// even if the account was deleted after it was issued. On success the
// authenticated user's id is stored via authctx, for any handler (events or
// users) to read back with authctx.UserID(context).
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

		authctx.SetUserID(context, userID)
		context.Next()
	}
}
