FROM docker.io/golang:1.19

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum

COPY cmd /app/cmd
COPY entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

WORKDIR /app

ENTRYPOINT ["/app/entrypoint.sh"]

