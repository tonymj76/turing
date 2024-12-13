package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

const (
	errorTag = "error"
)

func UnaryErrorLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Extract correlation ID from context
	correlationID := ctx.Value("correlation-id")
	if correlationID == nil {
		correlationID = "missing-correlation-id"
	}

	var err error
	resp, err := handler(ctx, req)

	if err != nil {
		// Categorize errors based on code
		var errorCategory string
		switch status.Code(err) {
		case codes.Internal:
			errorCategory = "internal"
		case codes.NotFound:
			errorCategory = "not-found"
		default:
			errorCategory = "other"
		}

		// Log error with context and categorization
		logger.Error().
			Str("method", info.FullMethod).
			Str("correlation-id", "").
			Str("error-category", errorCategory).
			Err(err).
			Msg(fmt.Sprintf("Unary request failed: %v", err))
	}

	return resp, err
}

func StreamErrorLoggingInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	correlationID := stream.Context().Value("correlation-id")
	if correlationID == nil {
		correlationID = "missing-correlation-id"
	}

	err := handler(srv, stream)

	if err != nil {
		errorCategory := "other"
		switch status.Code(err) {
		case codes.Internal:
			errorCategory = "internal"
		case codes.DeadlineExceeded:
			errorCategory = "deadline-exceeded"
		}

		logger.Error().
			Str("method", info.FullMethod).
			Str("correlation-id", "").
			Str("error-category", errorCategory).
			Err(err).
			Msg(fmt.Sprintf("Stream request failed: %v", err))
	}

	return err
}

func main() {
	rand.Seed(time.Now().UnixNano())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryErrorLoggingInterceptor),
		grpc.StreamInterceptor(StreamErrorLoggingInterceptor),
	)

	// Register your service handlers here

	fmt.Println("server starting on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
