FROM golang:1.25.3-alpine AS builder
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o orderservice ./services/OrderService/cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

# Copy the binary
COPY --from=builder /app/orderservice .

# Copy migration files (embedded in binary via go:embed)
COPY --from=builder /app/services/OrderService/internal/migrations ./internal/migrations

USER nonroot:nonroot

EXPOSE 8080 50055

CMD ["./orderservice"]