FROM golang:1.26.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

FROM alpine:latest

WORKDIR /root/

ADD https://github.com/pressly/goose/releases/download/v3.22.1/goose_linux_x86_64 /usr/local/bin/goose
RUN chmod +x /usr/local/bin/goose

COPY --from=builder /app/server .
COPY --from=builder /app/db/migrations ./db/migrations
COPY entrypoint.sh /entrypoint.sh
COPY .env /root/.env
RUN chmod +x /entrypoint.sh

EXPOSE 8090

ENTRYPOINT ["/entrypoint.sh"]