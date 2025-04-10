package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	setEnvVars()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("No path provided")
		return
	}

	folderPath := args[0]

	fmt.Println("Initializing NeoFS client...")
	uploadDir(folderPath)
	fmt.Println("Finished uploading files to NeoFS.")
}
