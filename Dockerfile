FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o rsshub ./cmd/main.go