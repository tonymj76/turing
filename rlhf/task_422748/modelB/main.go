package main

import (
	"context"
	"log"
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var appLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

func extractMetadata(ctx context.Context) map[string]string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	meta := make(map[string]string)
	for key, values := range md {
		meta[key] = strings.Join(values, ",")
	}
	return meta
}

func logError(method string, err error, meta map[string]string) {
	appLogger.Error().
		Str("method", method).
		Interface("metadata", meta).
		Err(err).
		Msg("Request failed")
}

func UnaryErrorLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		logError(info.FullMethod, err, extractMetadata(ctx))
	}
	return resp, err
}

func StreamErrorLoggingInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, stream)
	if err != nil {
		logError(info.FullMethod, err, nil)
	}
	return err
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryErrorLoggingInterceptor),
		grpc.StreamInterceptor(StreamErrorLoggingInterceptor),
	)

	// Register services here

	appLogger.Info().Msg("Server starting on port 50051")
	if err := s.Serve(lis); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to serve")
	}
}
