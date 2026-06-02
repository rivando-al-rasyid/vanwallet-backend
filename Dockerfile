FROM golang:1.26.3-alpine3.23 AS builder

WORKDIR /app

RUN apk add --no-cache git

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN swag init -g ./cmd/main.go

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /vanwallet ./cmd/


FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /vanwallet ./vanwallet
COPY --from=builder /app/public ./public
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./vanwallet"]