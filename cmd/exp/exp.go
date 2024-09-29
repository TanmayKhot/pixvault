package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func query(name, email string) string {
	return fmt.Sprintf(`INSERT INTO users(name,email) 
	VALUES(%s, %s);`, name, email)
}

func main() {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "password",
		Database: "pixvault",
		SSLMode:  "disable",
	}
	db, err := sql.Open("pgx", cfg.String())
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to db!")
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		amount INT,
		description TEXT
	);`)

	if err != nil {
		panic(err)
	}

	// 	name := "',''); DROP TABLE users; --"
	// 	email := "'tom@smith.com'"

	// 	row := db.QueryRow(`
	//   				INSERT INTO users (name, email)
	//   				VALUES ($1, $2)
	// 				RETURNING id;`, name, email)

	// 	var id int
	// 	err = row.Scan(&id)

	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("User created. Id = ", id)

	userID := 3
	row := db.QueryRow(`
			SELECT name FROM users
			WHERE id = $1;`, userID)

	var name string
	err = row.Scan(&name)
	if err != nil {
		panic(err)
	}
	fmt.Println("Name of id 3 is ", name)

	// for i := 1; i <= 5; i++ {
	// 	amount := i * 100
	// 	desc := fmt.Sprintf("Fake order #%d", i)
	// 	_, err := db.Exec(`
	// 	INSERT INTO orders(user_id, amount, description)
	// 	VALUES($1,$2,$3);`, userID, amount, desc)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
	// fmt.Println("Created fake orders")
	type Order struct {
		ID          int
		UserID      int
		Amount      int
		Description string
	}

	var orders []Order

	rows, err := db.Query(`
	SELECT id, amount, description
	FROM orders
	WHERE user_id=$1`, userID)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var order Order
		order.UserID = userID
		err := rows.Scan(&order.ID, &order.Amount, &order.Description)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("Here are the orders: ", orders)
}
