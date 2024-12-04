# Use an official Go runtime as the base image
FROM golang:1.21 AS build

# Set the working directory in the container to /app
WORKDIR /app

# Copy the Go source code into the container at /app
COPY . .

# Install any needed packages specified in go.mod
RUN go mod download

# Build your application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Build the final image with the application
FROM scratch

COPY --from=build /app/main .

CMD ["./main"]