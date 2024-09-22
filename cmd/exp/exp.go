package main

import (
	"fmt"
	"log"
)

func Connect() error {
	panic("Connection failed")
}

func CreateUser() error {
	err := Connect()
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func CreateOrg() error {
	err := CreateUser()
	if err != nil {
		return fmt.Errorf("create org: %w", err)
	}
	return nil
}

func main() {
	err := CreateUser()
	if err != nil {
		log.Println(err)
	}

	err = CreateOrg()
	if err != nil {
		log.Println(err)
	}
}
