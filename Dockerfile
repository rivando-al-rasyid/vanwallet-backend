FROM golang:1.26.3-alpine3.23 AS builder

WORKDIR /app

RUN apk add --no-cache git


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /vanwallet \
    ./cmd/main.go

FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache \
    tzdata

COPY --from=builder /vanwallet ./vanwallet
COPY --from=builder /app/public ./public
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./vanwallet"]