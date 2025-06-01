# Build stage
FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o ./server ./cmd/server/

# Final stage
FROM ubuntu:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY ./assets ./assets
COPY .env .env
COPY ./migrations ./migrations
CMD ["./server"]