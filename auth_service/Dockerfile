FROM golang:alpine AS builder

LABEL image_author="Michael Prunchak"

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN go build -o auth_service ./cmd

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/auth_service .
COPY --from=builder /app/config ./config

EXPOSE 8082

CMD ["./auth_service"]