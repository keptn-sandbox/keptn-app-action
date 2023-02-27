FROM docker.io/golang:1.19

COPY go.mod /app/
COPY go.sum /app/

COPY cmd /app/cmd
COPY entrypoint.sh /app/entrypoint.sh

WORKDIR /app

RUN ["./entrypoint.sh"]

