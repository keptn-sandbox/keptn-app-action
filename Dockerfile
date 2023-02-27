FROM docker.io/golang:1.19 as builder

COPY cmd /app/cmd
COPY entrypoint.sh /app/entrypoint.sh

WORKDIR /app

RUN ["./entrypoint.sh"]

