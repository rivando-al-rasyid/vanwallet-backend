# --- Build Stage ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o ./vanwallet \
    ./cmd/main.go


# --- Final Runtime Stage ---
FROM alpine:3.20

WORKDIR /app

# ca-certificates is required for outbound HTTPS requests
RUN apk add --no-cache tzdata 

# Corrected paths: Copying from /app to the current WORKDIR (.)
COPY --from=builder /app/vanwallet ./vanwallet
COPY --from=builder /app/public ./public
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./vanwallet"]