# Order Service

Manages order processing with cross-service validation. Coordinates with UserService and ProductService to ensure data consistency.

## Overview

- **Language**: Go
- **Protocol**: gRPC
- **Port**: 50054
- **Database**: PostgreSQL
- **Auth**: Internal service token
- **Validation**: User & product verification

## Features

✅ Create orders with validation
✅ Track order status
✅ Manage order items
✅ Cross-service validation (user, products)
✅ Transaction support
✅ Order history
✅ Distributed tracing
✅ Readable error messages

## Configuration

```env
APP_PORT=50054
APP_ENV=development
INTERNAL_AUTH_TOKEN=internal-token

# Database
DB_DRIVER=postgres
DB_DSN=postgres://user:pass@host:5432/orders
DB_MIGRATION_AUTO_RUN=true

# gRPC Clients (for validation)
USER_SERVICE_URL=localhost:50051
PRODUCT_SERVICE_URL=localhost:50053

# Tracing
JAEGER_ENDPOINT=localhost:4317
```

## gRPC API

### Order Operations
- `CreateOrder(CreateOrderRequest)` - Place new order
- `GetOrderByID(GetOrderByIDRequest)` - Fetch order details
- `ListUserOrders(ListUserOrdersRequest)` - Get user's orders
- `UpdateOrderStatus(UpdateOrderStatusRequest)` - Change order status
- `CancelOrder(CancelOrderRequest)` - Cancel pending order

**Request Structure:**
```protobuf
message CreateOrderRequest {
  string user_id = 1;           // User placing order
  repeated OrderItem items = 2; // Cart items
  string shipping_address = 3;  // Delivery address
}

message OrderItem {
  int32 product_id = 1;
  int32 quantity = 2;
  float unit_price = 3;
}

message GetOrderByIDRequest {
  int32 order_id = 1;
}
```

## Architecture

```
internal/
├── domain/                  # Order & OrderItem models
├── usecase/                 # Business logic & validation
├── repository/              # PostgreSQL access
│   └── postgresql/          # DB implementation
├── delivery/
│   └── grpc/                # gRPC handlers
└── clients/
    ├── user/                # User service client
    └── product/             # Product service client

cmd/
└── main.go                  # Startup & dependency injection

migrations/
└── *.sql                    # Database migrations
```

## Database Schema

```sql
-- Orders
CREATE TABLE orders (
  id SERIAL PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  total_price DECIMAL(10, 2) NOT NULL,
  shipping_address TEXT NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

-- Order Items
CREATE TABLE order_items (
  id SERIAL PRIMARY KEY,
  order_id INTEGER NOT NULL,
  product_id INTEGER NOT NULL,
  quantity INTEGER NOT NULL,
  unit_price DECIMAL(10, 2) NOT NULL,
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
```

## Validation Flow

1. **User Validation** - Call UserService.GetUser(user_id)
2. **Product Validation** - Call ProductService.GetProducts(product_ids)
3. **Inventory Check** - Verify stock levels
4. **Address Validation** - Ensure valid shipping address
5. **Transaction** - Create order atomically with items

## Order Status Workflow

```
pending → processing → shipped → delivered
   ↓
   ↓
cancelled (can cancel from pending)
```

## Running

```bash
# Local development
cd services/OrderService
go run cmd/main.go

# Docker
docker-compose up order-service

# Requirements
# - PostgreSQL running
# - UserService accessible at configured URL
# - ProductService accessible at configured URL
```

## Error Handling

### Readable Errors (No Raw SQL)
- `ErrUserNotFound` - User doesn't exist
- `ErrInvalidOrder` - Missing required fields
- `ErrProductNotFound` - Product doesn't exist
- `ErrInsufficientInventory` - Not enough stock
- `ErrOrderNotFound` - Order ID invalid
- `ErrInvalidAddress` - Address validation failed
- `ErrDatabaseConnection` - DB connection issue

Postgres errors (23505, 23503, 23502, 08xxx) are mapped to above errors.

## Performance

- **Create Order**: ~50ms (includes gRPC validations)
- **Get Order**: ~5ms (cached queries)
- **List Orders**: ~20ms (pagination)
- Connection pooling for concurrent requests
- Prepared statements prevent SQL injection

## Transactions

- Order creation is atomic
- All items created or none
- Foreign key constraints enforced
- Rollback on validation failure

## Security

- Internal service token for gRPC
- User isolation (can only view own orders)
- Admin endpoints for status updates
- Rate limiting at API Gateway
- Address validation

## Integration

Calls:
- **UserService** (gRPC): Validate user exists
- **ProductService** (gRPC): Validate products & stock

Called by:
- **ApiGateway** (REST → gRPC): Order operations
- **CartService** (may query for cart-to-order)

## Observability

- OpenTelemetry tracing on gRPC calls
- Structured logging with correlation IDs
- Distributed transaction tracing
- Error tracking with Jaeger
