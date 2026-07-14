// Package authctx carries the authenticated user's id on the gin request
// context. It exists so that transport packages for different domains
// (events/rest, users/rest) can share this without depending on each
// other — only on this tiny shared package, same as they already share
// utils.
package authctx

import "github.com/gin-gonic/gin"

const userIDKey = "userID"

// SetUserID stores the authenticated user's id on the request context.
// Called by auth middleware once a token has been validated.
func SetUserID(context *gin.Context, userID int64) {
	context.Set(userIDKey, userID)
}

// UserID returns the authenticated user's id. ok is false if no auth
// middleware ran on this request.
func UserID(context *gin.Context) (int64, bool) {
	v, exists := context.Get(userIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
