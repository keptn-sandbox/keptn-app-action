FROM docker.io/golang:1.19 as builder

RUN mkdir -p /build

COPY go.mod /build/go.mod
COPY go.sum /build/go.sum

WORKDIR /build

RUN go mod download

COPY . /build

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o keptn-config-generator cmd/keptn-update-action/main.go

FROM docker.io/debian:bullseye-slim

COPY --from=builder /build/keptn-config-generator /keptn-config-generator
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
