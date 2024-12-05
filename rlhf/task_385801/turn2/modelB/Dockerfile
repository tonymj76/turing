# Use an official Go base image
FROM golang:1.21 as build

# Set the working directory in the container
WORKDIR /app

# Copy the go.mod and go.sum files to the container
COPY go.mod go.sum ./

# Install any required dependencies
RUN go mod download

# Copy the rest of the code to the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Use a smaller Alpine Linux image as the runtime environment
FROM alpine:latest

# Copy the built binary to the runtime image
COPY --from=build /app/main .

# Expose the port the app listens on
EXPOSE 8080

# Run the application when the container starts
CMD ["./main"]