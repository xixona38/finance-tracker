package auth

import "context"

// UserIDFromContext reads the user ID placed in the context by the authentication middleware.
// The second result is false if there is no int64 user ID under the expected key.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	res := ctx.Value(userIDContextKey{})

	userID, ok := res.(int64)
	if !ok {
		return 0, false
	}

	return userID, true
}
