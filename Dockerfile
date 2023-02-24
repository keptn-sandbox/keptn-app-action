FROM docker.io/golang:1.19

COPY src /app
WORKDIR /app

CMD go run main.go
