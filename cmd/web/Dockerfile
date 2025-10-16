# Dockerfile for the web service
ARG GO_VERSION=1.24.4-alpine
ARG ALPINE_VERSION=3.21

# Builder stage
FROM golang:${GO_VERSION} AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

# Generate Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN cd /app && $GOPATH/bin/swag init -g cmd/web/main.go -o cmd/web/docs
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o web ./cmd/web/main.go


# Runtime stage
FROM alpine:${ALPINE_VERSION}
RUN apk --no-cache add busybox-suid curl rsync tzdata tcpdump tree ca-certificates s-nail \
    && ln -s /usr/bin/mail /usr/bin/s-nail

WORKDIR /app

COPY --from=builder /app/web .

COPY --from=builder /app/cmd/web/docs ./docs

EXPOSE 80

CMD ["./web"]