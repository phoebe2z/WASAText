package api

import (
	"net/http"
	"strconv"
	"strings"

	"git.phoebe2z/WASAText/service/api/reqcontext"
)

// authenticate extracts the user ID from the Authorization header.
// WASAText uses a simple "Bearer <userId>" scheme, where userId is an integer.
//
// If the route does not require authentication, handlers simply ignore ctx.AuthenticatedUser = 0.
// If authentication fails for a protected route, the handler must return 401 manually.
func (rt *_router) authenticate(w http.ResponseWriter, r *http.Request, ctx *reqcontext.RequestContext) error {

	// Read Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// No token provided → treat as unauthenticated (ctx.AuthenticatedUser = 0)
		return nil
	}

	// Expect format: "Bearer <id>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		// Invalid header format → ignore silently, handler will enforce 401 if needed
		return nil
	}

	// Parse user ID
	userIdStr := parts[1]
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil || userId <= 0 {
		// Invalid ID → treat as unauthenticated
		return nil
	}

	// Authentication success
	ctx.AuthenticatedUser = userId
	return nil
}
