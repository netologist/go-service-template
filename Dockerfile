FROM golang:1.27-alpine AS builder

WORKDIR /app

# download all dependencies for cache
COPY go.mod go.sum ./
RUN go mod download

# Copy all codes
COPY . .

# Build application server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/server

# Build goose migration binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/goose github.com/pressly/goose/v3/cmd/goose

# ---- Runtime ----
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

# Copy server binary and goose migration tool
COPY --from=builder /bin/server /server
COPY --from=builder /bin/goose /bin/goose

# Copy database migrations directory
COPY database/migrations /database/migrations

EXPOSE 8080

ENTRYPOINT ["/server"]
