package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	embedsql "github.com/mr-destructive/link-blog/embedsql"
	"github.com/mr-destructive/link-blog/models"
)

var (
	queries *models.Queries
)

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	ctx := context.Background()
	dbName := os.Getenv("DB_NAME")
	dbToken := os.Getenv("DB_TOKEN")

	var err error
	dbString := fmt.Sprintf("libsql://%s?authToken=%s", dbName, dbToken)
	db, err := sql.Open("libsql", dbString)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	defer db.Close()

	queries = models.New(db)
	if _, err := db.ExecContext(ctx, embedsql.DDL); err != nil {
		log.Printf("error creating tables: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil

}
