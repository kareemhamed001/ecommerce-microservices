# API Gateway Service

Entry point for all client requests. Converts HTTP/REST calls to gRPC calls to backend services.

## Overview

- **Language**: Go
- **Protocol**: HTTP/REST (Gin framework)
- **Port**: 8080
- **gRPC Clients**: User, Product, Cart, Order services

## Features

✅ JWT Authentication & Token Validation
✅ Role-Based Access Control (RBAC)
✅ Rate Limiting
✅ Circuit Breaker Pattern
✅ Structured Logging
✅ Graceful Shutdown
✅ Health Checks
✅ Request ID Tracing

## Configuration

```env
APP_PORT=8080
APP_ENV=development
JWT_SECRET=your-secret-key
INTERNAL_AUTH_TOKEN=internal-token

# Service URLs (gRPC)
USER_SERVICE_URL=localhost:50051
PRODUCT_SERVICE_URL=localhost:50053
CART_SERVICE_URL=localhost:50055
ORDER_SERVICE_URL=localhost:50057

# Circuit Breaker
CIRCUIT_BREAKER_ENABLED=true
CIRCUIT_BREAKER_MAX_REQUESTS=5
CIRCUIT_BREAKER_TIMEOUT=60s
CIRCUIT_BREAKER_FAILURE_RATIO=0.5

# Timeouts
REQUEST_TIMEOUT=30s
IDLE_TIMEOUT=120s
READ_TIMEOUT=15s
WRITE_TIMEOUT=15s
```

## Key Endpoints

### Auth

- `POST /api/v1/users/register` - Register user
- `POST /api/v1/users/login` - Login (returns JWT)

### Protected Endpoints (require valid JWT)

- All `/api/v1/users/*` endpoints (except register/login)
- All `/api/v1/addresses/*` endpoints
- All `/api/v1/cart/*` endpoints
- All `/api/v1/orders/*` endpoints

### Admin-Only Endpoints

- `GET /api/v1/users/search` - Search users
- `DELETE /api/v1/users/delete` - Delete user
- `POST /api/v1/products/create` - Create product
- `POST /api/v1/categories/create` - Create category
- `PATCH /api/v1/orders/status` - Update order status

## Architecture

```
internal/
├── router/          # Route definitions
├── handlers/        # HTTP request handlers
├── middleware/      # Auth, CORS, logging
└── clients/         # gRPC client connections

cmd/
└── main.go         # Startup & shutdown logic
```

## Running

```bash
# Local development
cd services/ApiGateway
go run cmd/main.go

# Docker
docker-compose up api-gateway
```

## Request Flow

1. Client sends HTTP request
2. API Gateway validates JWT
3. Checks RBAC permissions
4. Applies rate limiting
5. Calls appropriate gRPC service
6. Returns JSON response

## Security

- Tokens are validated at every protected endpoint
- Role checks prevent unauthorized access
- Circuit breakers protect against cascading failures
- Rate limiting prevents abuse
- Internal auth tokens secure service-to-service communication
