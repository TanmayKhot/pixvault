package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/TanmayKhot/pixvault/cmd/exp/models"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("SMTP_HOST")
	portstr := os.Getenv("SMTP_PORT")
	port, err := strconv.Atoi(portstr)
	if err != nil {
		panic(err)
	}
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	es := models.NewEmailService(models.SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	})

	err = es.ForgotPassword("t@t.com", "https://pixvault.com/reset-pw?token=abc123")
	if err != nil {
		panic(err)
	}
	fmt.Println("Email sent")
}
