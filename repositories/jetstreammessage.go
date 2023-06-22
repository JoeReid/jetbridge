package repositories

import "github.com/JoeReid/jetbridge"

//go:generate go run github.com/golang/mock/mockgen -destination=./mocks/mock_jetstreammessage.go -package=mocks . JetstreamMessage

type JetstreamMessage interface {
	Payload() jetbridge.JetstreamLambdaPayload
	Ack() error
	Nak() error
}
