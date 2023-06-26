package server

import (
	"context"
	"log"

	"github.com/bufbuild/connect-go"
)

func LoggingInterceptor(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		log.Printf("Request: %v", req.Spec().Procedure)
		return next(ctx, req)
	}
}
