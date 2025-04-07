# Stage 1: Build the Go application.
FROM golang:1.23-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o rlaas .

# Stage 2: Create a minimal container for the application.
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/rlaas .
EXPOSE 8080
CMD ["./rlaas"]
