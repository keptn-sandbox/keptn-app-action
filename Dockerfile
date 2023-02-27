FROM docker.io/golang:1.19

COPY go.mod .
COPY go.sum .

COPY cmd .
COPY entrypoint.sh .

RUN ["./entrypoint.sh"]

