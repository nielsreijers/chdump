package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func query(db *sql.DB, query string) string {
	var result string

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Query '%s' failed: %v", query, err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&result); err != nil {
			log.Fatalf("Query '%s' result scan failed: %v", query, err)
		}
	} else {
		log.Fatalf("Query '%s' yielded no results.", query)
	}

	return result
}

func main() {
	// Replace these with your actual database credentials and connection details

	clickhouseDSN := os.Args[1]

	// Establish a connection to the database
	// connStr := fmt.Sprintf("%s/%s?username=%s&password=%s", address, database, username, password)
	db, err := sql.Open("clickhouse", clickhouseDSN)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	// Ping the database to ensure connection is established
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Query the current database name
	dbName := query(db, "SELECT currentDatabase()")
	dbCreate := query(db, fmt.Sprintf("SHOW CREATE DATABASE %s", dbName))
	fmt.Printf("%s\n;\n----------------------------------------\n", dbCreate)

	// Query to retrieve table names in the specified database
	tableNamesQuery := "SHOW TABLES"
	rows, err := db.Query(tableNamesQuery)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		// Query to retrieve the DDL of each table
		tableCreate := query(db, fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName))
		tableUuid := query(db, fmt.Sprintf("SELECT uuid FROM system.tables WHERE database='%s' AND name='%s'", dbName, tableName))
		idx := strings.Index(tableCreate, "(")
		fmt.Printf("%s\n;\n----------------------------------------\n", fmt.Sprintf("%s UUID '%s'\n%s", tableCreate[:idx], tableUuid, tableCreate[idx:]))
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error during row iteration: %v", err)
	}
}
