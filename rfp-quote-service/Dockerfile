# Dockerfile for Service1
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o rfp-quote-service ./cmd

EXPOSE 8084

CMD ["./rfp-quote-service"]
