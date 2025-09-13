# Dockerfile for Service1
FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o category-service ./cmd

EXPOSE 8083

CMD ["./category-service"]
