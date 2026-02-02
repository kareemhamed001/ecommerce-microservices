# Product Service

Manages product catalog and categories. Includes Redis caching for performance. Exposes gRPC API.

## Overview

- **Language**: Go
- **Protocol**: gRPC
- **Port**: 50053
- **Database**: PostgreSQL
- **Cache**: Redis (30min TTL)
- **Auth**: Internal service token

## Features

✅ Product CRUD operations
✅ Category management
✅ Redis caching with TTL
✅ Inventory tracking
✅ Discount system (percentage, fixed)
✅ Full-text search
✅ Pagination
✅ Distributed tracing

## Configuration

```env
APP_PORT=50053
APP_ENV=development
INTERNAL_AUTH_TOKEN=internal-token

# Database
DB_DRIVER=postgres
DB_DSN=postgres://user:pass@host:5432/products
DB_MIGRATION_AUTO_RUN=true

# Redis Cache
REDIS_ENABLED=true
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Tracing
JAEGER_ENDPOINT=localhost:4317
```

## gRPC API

### Product Operations

- `CreateProduct(CreateProductRequest)` - Add product
- `GetProductByID(GetProductByIDRequest)` - Fetch product (with caching)
- `GetProductsByIDs(GetProductsByIDsRequest)` - Bulk fetch
- `ListProducts(ListProductsRequest)` - List with pagination
- `UpdateProduct(UpdateProductRequest)` - Update product info
- `DeleteProduct(DeleteProductRequest)` - Delete product

### Category Operations

- `CreateCategory(CreateCategoryRequest)` - Add category
- `GetCategoryByID(GetCategoryByIDRequest)` - Fetch category
- `ListCategories(ListCategoriesRequest)` - List with pagination
- `UpdateCategory(UpdateCategoryRequest)` - Update category
- `DeleteCategory(DeleteCategoryRequest)` - Delete category

## Architecture

```
internal/
├── domain/           # Product & category models
├── usecase/          # Business logic with caching
├── repository/       # PostgreSQL access
│   └── postgresql/   # DB implementation
├── cache/            # Redis caching
└── delivery/
    └── grpc/         # gRPC handlers

cmd/
└── main.go          # Startup & dependency injection
```

## Database Schema

```sql
-- Products
CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  short_description VARCHAR(500),
  price DECIMAL(10, 2) NOT NULL,
  discount_type VARCHAR(50),
  discount_value DECIMAL(10, 2),
  discount_start_date TIMESTAMP,
  discount_end_date TIMESTAMP,
  image_url VARCHAR(500),
  quantity INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Categories
CREATE TABLE categories (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);
```

## Caching Strategy

- **Product Cache**: 30-minute TTL
- **Category Cache**: 1-hour TTL
- **Cache Key Format**: `product:{id}`, `category:{id}`
- **Cache Invalidation**: On update/delete operations

## Running

```bash
# Local development
cd services/ProductService
go run cmd/main.go

# Docker
docker-compose up product-service
```

## Performance

- Redis caching reduces database queries
- Bulk product fetch for order processing
- Connection pooling for PostgreSQL
- Efficient pagination with limits

## Error Handling

- Readable error messages instead of raw SQL
- Postgres errors mapped to business errors
- Proper HTTP status codes via gRPC

## Security

- Internal service token for gRPC calls
- Query parameter binding prevents SQL injection
- Admin-only endpoints for modifications
- Rate limiting at API Gateway level
