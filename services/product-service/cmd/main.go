package main

import (
	"fmt"
	"os"
	"product-service/internal/handler"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env file for local development only
	if _, err := os.Stat("../../.env"); err == nil {
		err := godotenv.Load("../../.env")
		if err != nil {
			fmt.Printf("Error loading .env: %v\n", err)
		}
	}
	
	// These will work both locally (after loading .env) and in Lambda (via environment variables)
	fmt.Printf("DB_HOST: %s\n", os.Getenv("DB_HOST"))
	fmt.Printf("DB_PASSWORD: %s\n", os.Getenv("DB_PASSWORD"))
	fmt.Printf("DB_USER: %s\n", os.Getenv("DB_USER"))
	fmt.Printf("DB_NAME: %s\n", os.Getenv("DB_NAME"))
	fmt.Printf("DB_PORT: %s\n", os.Getenv("DB_PORT"))

	h := handler.NewLambdaHandler()
	lambda.Start(h.HandleRequest)
}
