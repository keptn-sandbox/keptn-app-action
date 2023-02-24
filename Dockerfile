FROM docker.io/golang:1.19 as builder

COPY . /build
RUN mkdir -p /build
WORKDIR /build
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o keptn-config-generator cmd/keptn-update-action/main.go

FROM docker.io/debian:bullseye-slim

COPY --from=builder /build/keptn-config-generator /keptn-config-generator

ENTRYPOINT ["/keptn-config-generator"]