package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/JoeReid/jetbridge"
	"github.com/JoeReid/jetbridge/repositories"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
)

var _ repositories.MessageHandler = (*MessageHandler)(nil)

func NewMessageHandler(lambda lambdaiface.LambdaAPI) *MessageHandler {
	return &MessageHandler{
		lambda: lambda,
	}
}

type MessageHandler struct {
	lambda lambdaiface.LambdaAPI
}

func (m *MessageHandler) HandleJetstreamMessages(ctx context.Context, binding repositories.JetstreamBinding, messages []repositories.JetstreamMessage) error {
	if binding.Batching != nil {
		var batch jetbridge.JetstreamBatchedLambdaPayload
		for _, message := range messages {
			batch = append(batch, message.Payload())
		}

		payload, err := json.Marshal(batch)
		if err != nil {
			for _, message := range messages {
				if err := message.Nak(); err != nil {
					log.Printf("failed to NAK message: %v", err)
				}
			}

			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := m.run(ctx, binding, payload); err != nil {
			for _, message := range messages {
				if err := message.Nak(); err != nil {
					log.Printf("failed to NAK message: %v", err)
				}
			}

			return fmt.Errorf("failed to run lambda: %w", err)
		}

		return nil
	}

	var rtnErr error
	for _, message := range messages {
		if rtnErr != nil {
			if err := message.Nak(); err != nil {
				log.Printf("failed to NAK message: %v", err)
			}
			continue
		}

		payload, err := json.Marshal(message.Payload())
		if err != nil {
			rtnErr = fmt.Errorf("failed to marshal message: %w", err)

			if err := message.Nak(); err != nil {
				log.Printf("failed to NAK message: %v", err)
			}

			continue
		}

		if err := m.run(ctx, binding, payload); err != nil {
			rtnErr = fmt.Errorf("failed to run lambda: %w", err)

			if err := message.Nak(); err != nil {
				log.Printf("failed to NAK message: %v", err)
			}

			continue
		}
	}

	return rtnErr
}

func (m *MessageHandler) run(ctx context.Context, binding repositories.JetstreamBinding, payload []byte) error {
	out, err := m.lambda.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName: &binding.LambdaARN,
		Payload:      payload,
	})
	if err != nil {
		return err
	}

	if out.FunctionError != nil {
		return errors.New(*out.FunctionError)
	}

	log.Println("lambda invoked successfully", "function_name", binding.LambdaARN, "function_version", *out.ExecutedVersion)

	return nil
}
