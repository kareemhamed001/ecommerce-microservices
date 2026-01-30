# Services Flow (User, Product, Cart, Order)

## High-Level Architecture

- **UserService**: user accounts and addresses.
- **ProductService**: product catalog.
- **CartService**: cart state in Redis.
- **OrderService**: orders in Postgres.
- **API Gateway**: calls services over gRPC.

All inter-service calls are **gRPC** on the shared Docker network.

## Required Dependencies

- UserService: Postgres
- ProductService: Postgres
- OrderService: Postgres
- CartService: Redis

## Service-to-Service Dependencies

- CartService → UserService (validate user exists)
- CartService → ProductService (validate product exists)
- OrderService → UserService (validate user exists)
- OrderService → ProductService (validate product exists)

## Startup Order (Local/Docker)

1. UserService + DB
2. ProductService + DB
3. CartService + Redis
4. OrderService + DB
5. API Gateway

## Common gRPC Endpoints

- UserService: `GetUserByID`
- ProductService: `GetProductByID`
- CartService: `GetCart`, `AddItem`, `UpdateItem`, `RemoveItem`, `ClearCart`
- OrderService: `CreateOrder`, `GetOrderByID`, `ListOrders`

## Flow A — Add to Cart

1. API Gateway calls CartService `AddItem(user_id, product_id, quantity)`.
2. CartService validates user via UserService `GetUserByID`.
3. CartService validates product via ProductService `GetProductByID`.
4. CartService updates Redis hash: `cart:{user_id}`.

## Flow B — View Cart

1. API Gateway calls CartService `GetCart(user_id)`.
2. CartService validates user via UserService.
3. CartService returns items and total quantity from Redis.

## Flow C — Create Order

1. API Gateway calls OrderService `CreateOrder` with user_id, shipping data, discount, and items.
2. OrderService validates user via UserService `GetUserByID`.
3. For each item, OrderService validates product via ProductService `GetProductByID`.
4. OrderService calculates totals and persists order + items in Postgres.

> Optional integration: after successful order creation, OrderService can call CartService `ClearCart(user_id)`.

## Flow D — Add Order Item

1. API Gateway calls OrderService `AddOrderItem(order_id, product_id, quantity)`.
2. OrderService validates product via ProductService.
3. OrderService persists item and recalculates totals.

## Networking Notes

- Services discover each other via Docker network aliases:
  - `userservice_app:50051`
  - `productservice_app:50053`
  - `cartservice_app:50057` (if added)
  - `orderservice_app:50055` (if added)

## Known Gaps

- No distributed transaction across services.
- No inventory reservation in ProductService.
- No idempotency on order creation.
- No auth/authorization between services.
