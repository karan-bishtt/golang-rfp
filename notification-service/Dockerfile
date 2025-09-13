# Dockerfile for Service1
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o notification-service ./cmd

EXPOSE 8082

CMD ["./notification-service"]
