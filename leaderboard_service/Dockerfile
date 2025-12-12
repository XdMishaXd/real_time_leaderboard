FROM golang:alpine AS builder

LABEL image_author="Michael Prunchak"

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN go build -o leaderboard_service ./leaderboard_service

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/leaderboard_service .
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./leaderboard_service"]