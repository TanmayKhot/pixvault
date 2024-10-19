package main

import (
	stdctx "context"
	"fmt"

	"github.com/TanmayKhot/pixvault/context"
	"github.com/TanmayKhot/pixvault/models"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	ctx := stdctx.Background()
	user := models.User{
		Email: "a@a.com",
	}

	ctx = context.WithUser(ctx, &user)
	retrieveduser := context.User(ctx)
	fmt.Println(retrieveduser.Email)
}
