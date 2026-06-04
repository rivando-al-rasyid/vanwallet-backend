FROM golang:1.26.4-alpine3.22 AS builder
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

FROM alpine:3.22
WORKDIR /app

RUN apk add --no-cache \
    tzdata \
    ca-certificates \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

COPY --from=builder /vanwallet ./vanwallet
COPY --from=builder /app/public ./public
COPY --from=builder /app/docs ./docs

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080
CMD ["./vanwallet"]