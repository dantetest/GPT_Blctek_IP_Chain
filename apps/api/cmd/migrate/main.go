package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

func main() {
	if len(os.Args) != 2 { fmt.Fprintln(os.Stderr, "usage: migrate <up|down|status>"); os.Exit(2) }
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" { dsn = "blctekip:blctekip@tcp(localhost:3306)/blctekip?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci" }
	db, err := sql.Open("mysql", dsn); if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }; defer db.Close()
	if err := goose.SetDialect("mysql"); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
	if err := goose.Run(os.Args[1], db, "migrations"); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}
