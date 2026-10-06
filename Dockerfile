FROM golang:1.27-alpine AS builder

WORKDIR /app

# download all dependencies for cache
COPY go.mod go.sum ./
RUN go mod download

# Copy all codes
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/server

# ---- Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
