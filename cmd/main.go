package main

import (
    "fmt"
    "os"
    "browseql/internal/app"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: browseql <database-file>")
        fmt.Println("Example: browseql test.db")
        os.Exit(1)
    }

    dbPath := os.Args[1]
    application := app.NewApp(dbPath)
    if err := application.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}