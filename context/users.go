package context

import (
	"context"

	"github.com/TanmayKhot/pixvault/models"
)

type key string

const (
	userKey key = "user"
)

// Using a function to add user to context without exporting the key
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	val := ctx.Value(userKey)      // Retrieve the user
	user, ok := val.(*models.User) // Convert the val to "user" data type (context returns "any" datatype)
	// ok will hold True if the type conversion was successful, else False
	if !ok {
		return nil
	}
	return user
}
