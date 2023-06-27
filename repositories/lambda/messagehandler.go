package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/JoeReid/jetbridge"
	"github.com/JoeReid/jetbridge/repositories"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"go.uber.org/zap"
)

var _ repositories.MessageHandler = (*MessageHandler)(nil)

type MessageHandler struct {
	logger *zap.Logger
	lambda lambdaiface.LambdaAPI
}

func (m *MessageHandler) HandleJetstreamMessages(ctx context.Context, binding repositories.JetstreamBinding, messages []repositories.JetstreamMessage) error {
	if binding.MaxMessages > 0 && binding.MaxLatency > 0 {
		var batch jetbridge.JetstreamBatchedLambdaPayload
		for _, message := range messages {
			batch = append(batch, message.Payload())
		}

		payload, err := json.Marshal(batch)
		if err != nil {
			for _, message := range messages {
				if err := message.Nak(); err != nil {
					m.logger.Error("failed to NAK message", zap.Error(err))
				}
			}

			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := m.run(ctx, binding, payload); err != nil {
			for _, message := range messages {
				if err := message.Nak(); err != nil {
					m.logger.Error("failed to NAK message", zap.Error(err))
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
				m.logger.Error("failed to NAK message", zap.Error(err))
			}
			continue
		}

		payload, err := json.Marshal(message.Payload())
		if err != nil {
			rtnErr = fmt.Errorf("failed to marshal message: %w", err)

			if err := message.Nak(); err != nil {
				m.logger.Error("failed to NAK message", zap.Error(err))
			}

			continue
		}

		if err := m.run(ctx, binding, payload); err != nil {
			rtnErr = fmt.Errorf("failed to run lambda: %w", err)

			if err := message.Nak(); err != nil {
				m.logger.Error("failed to NAK message", zap.Error(err))
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
		var logs string
		if out.LogResult != nil {
			logs = *out.LogResult
		}

		m.logger.Debug(
			"lambda returned error",
			zap.String("function_name", binding.LambdaARN),
			zap.String("function_version", *out.ExecutedVersion),
			zap.String("error", *out.FunctionError),
			zap.String("logs", logs),
		)
		return errors.New(*out.FunctionError)
	}

	m.logger.Info("lambda invoked successfully", zap.String("function_name", binding.LambdaARN), zap.String("function_version", *out.ExecutedVersion))

	return nil
}

func NewMessageHandler(lambda lambdaiface.LambdaAPI) (*MessageHandler, error) {
	zl, err := zap.NewDevelopment() // TODO: this needs to be managed better
	if err != nil {
		return nil, err
	}
	zl.With(zap.String("component", "lambda"))

	return &MessageHandler{
		logger: zl,
		lambda: lambda,
	}, nil
}
