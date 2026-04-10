package authctx

import "context"

type contextKey string

const currentUserKey contextKey = "current_user"

type CurrentUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func WithCurrentUser(ctx context.Context, user CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}

func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	user, ok := ctx.Value(currentUserKey).(CurrentUser)
	return user, ok
}
