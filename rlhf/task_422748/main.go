package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

type loggerKey struct{}

// GetLogger retrieves the logger from the context.
func GetLogger(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value(loggerKey{}).(*log.Logger)
	if !ok {
		return nil
	}
	return logger
}

// StreamServerInterceptor logs stream errors.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logger := GetLogger(ss.Context())
	if logger == nil {
		logger = log.New(os.Stderr, "MyApp: ", log.LstdFlags)
	}

	err := handler(srv, ss)
	if err != nil {
		logger.Printf("Stream error: %v", err)
	}
	return err
}

// UnaryServerInterceptor logs unary errors.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger := GetLogger(ctx)
	if logger == nil {
		logger = log.New(os.Stderr, "MyApp: ", log.LstdFlags)
		ctx = context.WithValue(ctx, loggerKey{}, logger)
	}

	resp, err := handler(ctx, req)
	if err != nil {
		logger.Printf("Unary error: %v", err)
	}
	return resp, err
}

func main() {
	// Create a custom logger
	customLogger := log.New(os.Stderr, "MyApp: ", log.LstdFlags)

	// Create the server with interceptors
	s := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor),
		grpc.StreamInterceptor(StreamServerInterceptor),
	)

	// Set up the listener
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Add logger to the base context
	baseCtx := context.WithValue(context.Background(), loggerKey{}, customLogger)

	log.Println("Server is starting...")
	if err := s.Serve(listener); err != nil {
		GetLogger(baseCtx).Fatalf("Failed to serve: %v", err)
	}
}
