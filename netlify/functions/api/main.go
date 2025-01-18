package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	embedsql "github.com/mr-destructive/link-blog/embedsql"
	"github.com/mr-destructive/link-blog/models"
)

var (
	queries      *models.Queries
	listTemplate *template.Template
	linkTemplate *template.Template
	editTemplate *template.Template
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
	linkTemplate = template.Must(template.New("link").Parse(embedsql.LinkHTML))
	listTemplate = template.Must(template.New("list").Parse(embedsql.ListHTML))
	editTemplate = template.Must(template.New("edit").Parse(embedsql.EditHTML))
	switch req.HTTPMethod {
	case http.MethodGet:
		if req.QueryStringParameters["id"] != "" {

		}

		var links []models.Link
		links, err = queries.ListLinks(ctx)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}
		return respond(req, links)
	case http.MethodPost:

		var link models.CreateLinkParams
		formData, err := url.ParseQuery(req.Body)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
		}
		Url := formData.Get("url")
		content := formData.Get("commentary")
		if content == "" || Url == "" {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
		}
		link.Url = Url
		link.Commentary = content
		createdLinkId, err := queries.CreateLink(ctx, link)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
		}
		createdLink, err := queries.GetLink(ctx, createdLinkId)
		return respond(req, createdLink)
	case http.MethodPut:
		linkIdStr := req.QueryStringParameters["id"]
		formData, err := url.ParseQuery(req.Body)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
		}
		if len(formData) == 0 && linkIdStr != "" {
			linkId, err := strconv.Atoi(linkIdStr)
			linkObj, err := queries.GetLink(ctx, int64(linkId))
			var tpl bytes.Buffer
			err = editTemplate.Execute(&tpl, linkObj)
			if err != nil {
				return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
			}
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers:    map[string]string{"Content-Type": "text/html"},
				Body:       tpl.String(),
			}, nil
		}
		linkId, err := strconv.Atoi(linkIdStr)
		linkObj, err := queries.GetLink(ctx, int64(linkId))
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid link ID"}, nil
		}
		var link models.UpdateLinkParams
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
		}
		Url := formData.Get("url")
		content := formData.Get("commentary")
		if content == "" || Url == "" {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
		}
		if Url != "" && Url != linkObj.Url {
			link.Url = Url
		}
		link.ID = linkObj.ID
		link.Url = Url
		link.Commentary = content
		err = queries.UpdateLink(ctx, link)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
		}
		linkObj, err = queries.GetLink(ctx, int64(linkId))
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
		}
		return respond(req, linkObj)
	//case http.MethodDelete:
	default:
		return events.APIGatewayProxyResponse{StatusCode: 200}, err
	}
}

func respond(req events.APIGatewayProxyRequest, data any) (events.APIGatewayProxyResponse, error) {
	log.Printf("request headers: %v", req.Headers)

	if req.Headers["hx-request"] == "true" {
		var tpl bytes.Buffer

		switch v := data.(type) {
		case []models.Link:
			err := listTemplate.Execute(&tpl, v)
			if err != nil {
				return events.APIGatewayProxyResponse{StatusCode: 500}, err
			}
		case models.Link:
			err := linkTemplate.Execute(&tpl, v)
			if err != nil {
				return events.APIGatewayProxyResponse{StatusCode: 500}, err
			}
		default:
			return events.APIGatewayProxyResponse{StatusCode: 400}, fmt.Errorf("unsupported data type for HTML fragment generation: %T", data)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       tpl.String(),
		}, nil
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(dataBytes),
	}, nil
}
