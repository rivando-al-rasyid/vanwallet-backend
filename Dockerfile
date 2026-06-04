FROM golang:1.26.3-alpine3.23 AS builder

WORKDIR /app

# Install git for fetching dependencies (if needed)
RUN apk add --no-cache git

# Cache Go modules dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary within the /app directory
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o ./vanwallet \
    ./cmd/main.go

FROM alpine:3.23

WORKDIR /app

RUN apk add --no-cache tzdata

# Copy artifacts from the builder stage
COPY --from=builder /app/vanwallet ./vanwallet
COPY --from=builder /app/public ./public

EXPOSE 8080

CMD ["./vanwallet"]