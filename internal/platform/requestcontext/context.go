package requestcontext

import "context"

type userIDContextKey struct{}

// UserIDFromContext reads the user ID added by the authentication middleware.
// It returns false if the value is missing or is not an int64.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	res := ctx.Value(userIDContextKey{})

	userID, ok := res.(int64)
	if !ok {
		return 0, false
	}

	return userID, true
}

// WithUserID returns a new context containing the user ID, leaving the original context unchanged.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}
