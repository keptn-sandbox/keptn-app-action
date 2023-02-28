FROM docker.io/golang:1.19-alpine as builder
RUN apk --no-cache add ca-certificates


RUN mkdir -p /build

COPY go.mod /build/go.mod
COPY go.sum /build/go.sum

WORKDIR /build

RUN go mod download

COPY . /build

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o keptn-config-generator cmd/keptn-update-action/main.go

FROM alpine as final

# copy ca certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /build/keptn-config-generator /keptn-config-generator

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
