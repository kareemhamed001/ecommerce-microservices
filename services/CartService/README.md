# Cart Service

High-performance shopping cart backed by Redis. Manages user cart items with fast operations.

## Overview

- **Language**: Go
- **Protocol**: gRPC
- **Port**: 50055
- **Database**: Redis
- **Auth**: Internal service token
- **Performance**: Sub-millisecond operations

## Features

✅ Add/remove items to cart
✅ Update item quantities
✅ Get user cart
✅ Clear cart
✅ Atomic operations (thread-safe)
✅ Session-based cart storage
✅ Cart expiration support
✅ Distributed tracing

## Configuration

```env
APP_PORT=50055
APP_ENV=development
INTERNAL_AUTH_TOKEN=internal-token

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Tracing
JAEGER_ENDPOINT=localhost:4317
```

## gRPC API

### Cart Operations

- `AddItem(AddItemRequest)` - Add/update item quantity
- `RemoveItem(RemoveItemRequest)` - Remove item from cart
- `GetCart(GetCartRequest)` - Fetch user's cart
- `ClearCart(ClearCartRequest)` - Empty cart
- `UpdateItem(UpdateItemRequest)` - Modify item quantity

**Request Structure:**

```protobuf
message AddItemRequest {
  string user_id = 1;      // User identifier
  int32 product_id = 2;    // Product to add
  int32 quantity = 3;      // Quantity to add
}

message GetCartRequest {
  string user_id = 1;      // User identifier
}

message GetCartResponse {
  map<string, int32> items = 1;  // {product_id: quantity}
}
```

## Architecture

```
internal/
├── domain/           # Cart models
├── usecase/          # Business logic
├── repository/       # Redis implementation
│   └── redis/        # Redis-specific code
└── delivery/
    └── grpc/         # gRPC handlers

cmd/
└── main.go          # Startup & dependency injection
```

## Redis Schema

**Key Pattern:** `cart:{user_id}`  
**Data Type:** Hash

```
HSET cart:user123 {
  "product_1": 2,
  "product_5": 1,
  "product_12": 3
}
```

## Operations

### Add Item

- Uses `HIncrBy` for atomic quantity increment
- Creates hash if doesn't exist
- O(1) operation

### Get Cart

- Uses `HGetAll` to fetch all items
- Returns product_id → quantity map
- O(N) where N = items in cart

### Remove Item

- Uses `HDel` for atomic deletion
- Returns error if item not found
- O(1) operation

### Clear Cart

- Uses `Del` to remove entire hash
- Fast cleanup
- O(1) operation

## Running

```bash
# Local development
cd services/CartService
go run cmd/main.go

# Docker
docker-compose up cart-service

# Requirements
# - Redis running (localhost:6379 by default)
```

## Performance

- **Add to cart**: ~0.1ms
- **Get cart**: ~0.5ms (with 10 items)
- **Remove item**: ~0.1ms
- **No database hits**: All operations in-memory

## Scalability

- Redis handles millions of carts
- Connection pooling for concurrent requests
- Horizontal scaling via Redis Cluster

## Error Handling

- User not found validation
- Invalid quantity checks
- Redis connection errors
- Proper gRPC error codes

## Security

- Internal service token required
- User isolation (no cross-user access)
- Rate limiting at API Gateway
- Cart data expires with session

## Integration

Called by:

- **ApiGateway**: REST endpoints for cart operations
- **OrderService**: Fetch cart before order creation

Calls:

- **Redis**: All persistence layer
