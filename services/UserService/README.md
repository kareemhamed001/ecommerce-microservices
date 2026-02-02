# User Service

Manages user accounts, authentication, and address management. Exposes gRPC API.

## Overview

- **Language**: Go
- **Protocol**: gRPC
- **Port**: 50051
- **Database**: PostgreSQL
- **Auth**: JWT, Internal service token

## Features

✅ User registration & login
✅ JWT token generation
✅ Password hashing & verification
✅ Role management (admin, customer)
✅ Address management (create, update, delete, list)
✅ User search & filtering
✅ Distributed tracing
✅ Structured logging

## Configuration

```env
APP_PORT=50051
APP_ENV=development
JWT_SECRET=your-secret-key
INTERNAL_AUTH_TOKEN=internal-token

# Database
DB_DRIVER=postgres
DB_DSN=postgres://user:pass@host:5432/ecommerce
DB_MIGRATION_AUTO_RUN=true

# Tracing
JAEGER_ENDPOINT=localhost:4317
```

## gRPC API

### User Operations

- `CreateUser(CreateUserRequest)` - Register new user
- `Login(LoginRequest)` - Authenticate user
- `GetUserByID(GetUserByIDRequest)` - Fetch user details
- `UpdateUser(UpdateUserRequest)` - Update user info
- `DeleteUser(DeleteUserRequest)` - Delete user
- `SearchUsers(SearchUsersRequest)` - Search with pagination

### Address Operations

- `CreateAddress(CreateAddressRequest)` - Add address
- `GetAddressByID(GetAddressByIDRequest)` - Get address
- `ListAddressesByUserID(ListAddressesByUserIDRequest)` - List user addresses
- `UpdateAddress(UpdateAddressRequest)` - Update address
- `DeleteAddress(DeleteAddressRequest)` - Delete address

## Architecture

```
internal/
├── domain/           # User & address models
├── usecase/          # Business logic
├── repository/       # PostgreSQL access
│   └── postgresql/   # DB implementation
└── delivery/
    └── grpc/         # gRPC handlers

cmd/
└── main.go          # Startup & dependency injection
```

## Database Schema

```sql
-- Users
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  role VARCHAR(50) NOT NULL DEFAULT 'customer',
  created_at TIMESTAMP DEFAULT NOW()
);

-- Addresses
CREATE TABLE addresses (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  country VARCHAR(100),
  city VARCHAR(100),
  state VARCHAR(100),
  street VARCHAR(255),
  zip_code VARCHAR(20),
  created_at TIMESTAMP DEFAULT NOW()
);
```

## Running

```bash
# Local development
cd services/UserService
go run cmd/main.go

# Docker
docker-compose up user-service
```

## Error Handling

- Readable error messages instead of raw SQL errors
- Postgres errors mapped to business logic errors
- Proper error codes for each operation

## Security

- Passwords hashed with bcrypt
- JWT tokens for stateless authentication
- Internal service token for gRPC calls
- RBAC checks at service level
- Database query parameter binding prevents SQL injection
