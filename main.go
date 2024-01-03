package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sod-auctions/auctions-db"
	"log"
	"net/http"
	"os"
	"strconv"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

var database *auctions_db.Database

func init() {
	log.SetFlags(0)
	var err error
	database, err = auctions_db.NewDatabase(os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
}

func getIntervalAndLimit(rangeParam string) (int, int) {
	var interval, limit int

	switch rangeParam {
	case "1d":
		interval = 1
		limit = 24
	case "1w":
		interval = 1
		limit = 168
	case "1m":
		interval = 6
		limit = 124
	case "3m":
		interval = 12
		limit = 186
	case "max":
		interval = 24
		limit = 1825
	default:
		interval = 0
		limit = 0
	}

	return interval, limit
}

type Auction struct {
	Timestamp int32 `json:"timestamp"`
	Quantity  int32 `json:"quantity"`
	Min       int32 `json:"min"`
	P05       int32 `json:"p05"`
	P10       int32 `json:"p10"`
	P25       int32 `json:"p25"`
	P50       int32 `json:"p50"`
	P75       int32 `json:"p75"`
	P90       int32 `json:"p90"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	realmId, _ := strconv.Atoi(event.QueryStringParameters["realmId"])
	auctionHouseId, _ := strconv.Atoi(event.QueryStringParameters["auctionHouseId"])
	itemId, _ := strconv.Atoi(event.QueryStringParameters["itemId"])
	rangeParam, _ := event.QueryStringParameters["range"]

	interval, limit := getIntervalAndLimit(rangeParam)

	auctions, err := database.GetAuctions(
		int16(interval), int16(realmId), int16(auctionHouseId), int32(itemId), int16(limit))
	if err != nil {
		log.Printf("An error occurred: %v\n", err)

		errorMessage := ErrorMessage{Error: "An internal error occurred"}
		body, _ := json.Marshal(errorMessage)

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "http://localhost:3000",
				"Access-Control-Allow-Methods": "GET, OPTIONS",
				"Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept, Authorization",
			},
			Body: string(body),
		}, nil
	}

	var mAuctions []*Auction
	for _, auction := range auctions {
		mAuctions = append(mAuctions, &Auction{
			Timestamp: auction.Timestamp,
			Quantity:  auction.Quantity,
			Min:       auction.Min,
			P05:       auction.P05,
			P10:       auction.P10,
			P25:       auction.P25,
			P50:       auction.P50,
			P75:       auction.P75,
			P90:       auction.P90,
		})
	}

	body, _ := json.Marshal(mAuctions)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "http://localhost:3000",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept, Authorization",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
