# 🛒 E-Commerce Microservices Platform

A production-ready, scalable e-commerce platform built with Go, microservices architecture, and gRPC. This project demonstrates modern distributed systems patterns and best practices.

## 📋 Quick Start

### Prerequisites

- Go 1.25.3+
- Docker & Docker Compose
- PostgreSQL 16+
- Redis 7+

### Development Setup (Docker Compose)

```bash
# Start all services
make up

# View logs
docker compose logs -f

# Stop services
make down
```

The API Gateway will be available at `http://localhost:8080`

### Project Structure

```
.
├── services/                    # Microservices
│   ├── ApiGateway/            # HTTP → gRPC gateway (Port 8080)
│   ├── UserService/           # User & address (gRPC:50051)
│   ├── ProductService/        # Products & categories (gRPC:50053)
│   ├── CartService/           # Shopping cart (gRPC:50055)
│   └── OrderService/          # Orders (gRPC:50057)
├── pkg/                        # Shared packages
│   ├── db/                    # Database initialization
│   ├── jwt/                   # JWT authentication
│   ├── logger/                # Structured logging
│   ├── redis/                 # Redis client
│   ├── tracer/                # OpenTelemetry
│   └── grpcmiddleware/        # gRPC interceptors
├── shared/                     # Protocol Buffers definitions
└── docker-compose.yaml         # Local development
```

---

## 🏗️ System Architecture

```
Client (Web/Mobile)
         │ HTTP (REST)
         ↓
    API Gateway (Gin, JWT, RBAC)
         │
    ┌────┴────┬────┬──────┬────────────┐
    ↓         ↓    ↓      ↓            ↓
  User      Product Cart  Order      Jaeger
  Service   Service  Svc   Service    (Tracing)
  (gRPC)    (gRPC)   (gRPC)(gRPC)
    │         │      │      │
    ↓         ↓      ↓      ↓
  Postgres  Postgres Redis Postgres
```

---

## 🚀 Core Services

| Service             | Port  | Protocol | Database   | Purpose                        |
| ------------------- | ----- | -------- | ---------- | ------------------------------ |
| **API Gateway**     | 8080  | HTTP     | -          | REST entry point, JWT, routing |
| **User Service**    | 50051 | gRPC     | PostgreSQL | Users, auth, addresses         |
| **Product Service** | 50053 | gRPC     | PostgreSQL | Products, categories           |
| **Cart Service**    | 50055 | gRPC     | Redis      | Shopping cart                  |
| **Order Service**   | 50057 | gRPC     | PostgreSQL | Order processing               |
| **Jaeger**          | 16686 | HTTP     | -          | Distributed tracing            |

---

## 📚 API Endpoints

### Authentication

```bash
POST   /api/v1/users/register        # Register
POST   /api/v1/users/login           # Login
```

### Users (Authenticated)

```bash
GET    /api/v1/users/profile         # Get profile
PUT    /api/v1/users/update          # Update profile
GET    /api/v1/users/search          # Search (admin)
DELETE /api/v1/users/delete          # Delete (admin)
```

### Addresses

```bash
POST   /api/v1/addresses/create      # Create
GET    /api/v1/addresses/list        # List
PUT    /api/v1/addresses/update      # Update
DELETE /api/v1/addresses/delete      # Delete
```

### Products

```bash
GET    /api/v1/products              # List
GET    /api/v1/products/by-id        # Get
POST   /api/v1/products/create       # Create (admin)
PUT    /api/v1/products/update       # Update (admin)
DELETE /api/v1/products/delete       # Delete (admin)
```

### Categories

```bash
GET    /api/v1/categories            # List
GET    /api/v1/categories/by-id      # Get
POST   /api/v1/categories/create     # Create (admin)
PUT    /api/v1/categories/update     # Update (admin)
DELETE /api/v1/categories/delete     # Delete (admin)
```

### Cart

```bash
GET    /api/v1/cart                  # Get
POST   /api/v1/cart/items/add        # Add item
PUT    /api/v1/cart/items/update     # Update qty
DELETE /api/v1/cart/items/remove     # Remove
DELETE /api/v1/cart/clear            # Clear
```

### Orders

```bash
POST   /api/v1/orders/create         # Create
GET    /api/v1/orders                # List
GET    /api/v1/orders/by-id          # Get
PATCH  /api/v1/orders/status         # Update (admin)
```

---

## 🔒 Security

- ✅ **JWT Authentication**: Stateless, token-based
- ✅ **RBAC**: Admin & Customer roles
- ✅ **Internal Service Auth**: Secure gRPC
- ✅ **Circuit Breakers**: Fault tolerance
- ✅ **Error Abstraction**: No SQL leaks
- ✅ **Graceful Shutdown**: Proper cleanup

---

## 🔍 Observability

- **Tracing**: Jaeger UI at `http://localhost:16686`
- **Logging**: Structured JSON with correlation IDs
- **Health Checks**: `GET /health` and `/api/v1/health`

---

## 🛠️ Development Commands

```bash
# Start all services
make up

# Stop services
make down

# Generate gRPC code
make proto

# View logs
docker compose logs -f

# Health check
curl http://localhost:8080/health
```

---

## 📖 Service Documentation

- [API Gateway](services/ApiGateway/README.md)
- [User Service](services/UserService/README.md)
- [Product Service](services/ProductService/README.md)
- [Cart Service](services/CartService/README.md)
- [Order Service](services/OrderService/README.md)

---

## 🚢 Deployment

### Docker Compose

```bash
docker compose up --build
```

---

## 📊 Key Technologies

| Component     | Tech                   |
| ------------- | ---------------------- |
| Language      | Go 1.25.3              |
| HTTP          | Gin                    |
| gRPC          | Protocol Buffers v3    |
| ORM           | GORM                   |
| Databases     | PostgreSQL 16, Redis 7 |
| Logging       | Zap                    |
| Tracing       | OpenTelemetry + Jaeger |
| Container     | Docker                 |

---

**For detailed information, see service-specific README files.**
