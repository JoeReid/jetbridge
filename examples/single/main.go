package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/JoeReid/jetbridge"
	"github.com/aws/aws-lambda-go/lambda"
)

type Message struct {
	SentAt time.Time `json:"sent_at"`
	SentBy string    `json:"sent_by"`
	Text   string    `json:"text"`
}

func run(ctx context.Context, payload jetbridge.JetstreamLambdaPayload) error {
	var msg Message
	if err := json.Unmarshal(payload.Data, &msg); err != nil {
		return err
	}

	log.Printf("Message from %s at %s: %s", msg.SentBy, msg.SentAt, msg.Text)
	return nil
}

func main() {
	lambda.Start(run)
}
