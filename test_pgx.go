package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	fmt.Println("testing sql.Open")
	db, err := sql.Open("pgx", "")
	fmt.Printf("db: %v, err: %v\n", db, err)
}
