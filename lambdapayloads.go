package jetbridge

import "github.com/nats-io/nats.go"

// TODO: refactor to be a root level package

type JetstreamBatchedLambdaPayload []JetstreamLambdaPayload

type JetstreamLambdaPayload struct {
	Subject  string
	Header   nats.Header
	Data     []byte
	Metadata nats.MsgMetadata
}
