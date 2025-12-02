package main

import (
	"fmt"
	"os"

	"browseql/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: browseql <database-file>")
		fmt.Println("\nExamples:")
		fmt.Println("  browseql mydb.db")
		fmt.Println("  browseql test.db")
		os.Exit(1)
	}

	dbPath := os.Args[1]

	// Create and run the application
	application := app.NewApp(dbPath)
	if err := application.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
