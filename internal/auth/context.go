package auth

import "context"

// UserIDFromContext retrieves the authenticated user's ID from the context.
// It returns zero and false if the value is absent or is not an int64.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	res := ctx.Value(userIDContextKey{})

	userID, ok := res.(int64)
	if !ok {
		return 0, false
	}

	return userID, true
}
