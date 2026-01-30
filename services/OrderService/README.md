# OrderService

## Overview

OrderService manages orders and order items. It exposes a gRPC API consumed by the API Gateway. It validates user and product existence by calling UserService and ProductService over gRPC before persisting data.

## Responsibilities

- Create orders with multiple items.
- Validate user existence via UserService gRPC.
- Validate product existence via ProductService gRPC.
- Store shipping cost, shipping duration, discounts, and order totals.
- Provide order listing and status updates.

## Data Model

- **Order**
  - `user_id`
  - `shipping_cost`
  - `shipping_duration_days`
  - `discount`
  - `total`
  - `status`
  - `items` (one-to-many)
- **OrderItem**
  - `order_id`
  - `product_id` (numeric only; no FK)
  - `quantity`
  - `unit_price`
  - `total_price`

## gRPC API

Defined in [shared/proto/v1/order.proto](../../shared/proto/v1/order.proto):

- `CreateOrder`
- `GetOrderByID`
- `ListOrders`
- `AddOrderItem`
- `RemoveOrderItem`
- `UpdateOrderStatus`

## Configuration

Environment file: [services/OrderService/config/.env](config/.env)

Key variables:

- `DB_DSN`: Postgres connection string.
- `GRPC_PORT`: OrderService gRPC port.
- `PRODUCT_SERVICE_GRPC_ADDR`: ProductService gRPC address.
- `USER_SERVICE_GRPC_ADDR`: UserService gRPC address.

## Database

Migrations are located in [services/OrderService/internal/migrations](internal/migrations).
Tables:

- `orders`
- `order_items`

## How it Works (Flow)

1. `CreateOrder` validates user via UserService gRPC.
2. Each item validates product via ProductService gRPC.
3. Totals are computed (items + shipping - discount).
4. Order and items are saved in Postgres.

## Drawbacks / Not Best Practices (Current State)

- **No transactional consistency across services**: User/product checks are remote calls without distributed transactions or saga orchestration. A product can be deleted after validation, leading to stale orders.
- **No inventory reservation**: Product existence is verified, but stock is not reserved or decremented. Orders can be created even if inventory is insufficient.
- **No idempotency**: `CreateOrder` does not handle duplicate client retries and can create duplicate orders.
- **Minimal error modeling**: gRPC errors are raw; no standardized error codes or details for clients.
- **No pagination limits**: `ListOrders` defaults are basic and can be abused with large `per_page`.
- **No auth/authorization**: The service assumes trusted callers and does not verify user permissions.
- **No FK for product_id**: This is intentional due to microservices, but it also means referential integrity is not enforced at DB level.
- **No audit trail**: Status changes are not tracked over time (missing history table/events).
- **No outbox/events**: Changes are not emitted to a message broker for other services.
- **No rate limiting**: gRPC endpoints can be spammed without protection.

## Suggested Improvements

- Add idempotency keys for order creation.
- Introduce saga/outbox patterns for cross-service consistency.
- Validate inventory and/or reserve stock with ProductService.
- Add structured gRPC error codes and metadata.
- Add auth middleware and role-based checks.
- Add order status history table.
- Emit order events to RabbitMQ or similar.
