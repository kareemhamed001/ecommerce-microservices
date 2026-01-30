FROM golang:1.25.3-alpine AS builder
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o cartservice ./services/CartService/cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

# Copy the binary
COPY --from=builder /app/cartservice .

USER nonroot:nonroot

EXPOSE 8080 50057

CMD ["./cartservice"]