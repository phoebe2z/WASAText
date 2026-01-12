package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "wasatext.db") // assuming default name
	if err != nil {
		fmt.Printf("Error opening: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='participants'").Scan(&name)
	if err != nil {
		fmt.Printf("Participants table existence: %v\n", err)
	} else {
		fmt.Printf("Participants table exists\n")
		rows, err := db.Query("PRAGMA table_info(participants)")
		if err != nil {
			fmt.Printf("Error getting table info: %v\n", err)
		} else {
			defer rows.Close()
			fmt.Println("Columns in participants:")
			for rows.Next() {
				var cid int
				var name string
				var dtype string
				var notnull int
				var dflt_value interface{}
				var pk int
				rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk)
				fmt.Printf("- %s (%s)\n", name, dtype)
			}
		}
	}
}
