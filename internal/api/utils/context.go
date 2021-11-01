package utils

import "context"

type ContextKey int

const (
	USER_ID_CTX_KEY ContextKey = 0
)

func SetUserID(ctx context.Context, uid UID) context.Context {
	return context.WithValue(ctx, USER_ID_CTX_KEY, uid)
}

func GetUserID(ctx context.Context) UID {
	return ctx.Value(USER_ID_CTX_KEY).(UID)
}
