FROM golang:alpine AS builder

LABEL image_author="Michael Prunchak"

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o mail-sender ./cmd/mail-sender

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/mail-sender .
COPY --from=builder /app/config ./config

CMD ["./mail-sender"]