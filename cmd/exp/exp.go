package main

import (
	"fmt"

	"github.com/TanmayKhot/pixvault/models"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	cfg := models.DefaultPostgresConfig()

	db, err := models.OpenDBConnection(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected!")

	us := models.UserService{
		DB: db,
	}
	user, err := us.Create("bob@bobross.com", "bob123")
	if err != nil {
		panic(err)
	}
	fmt.Println(user)
}
