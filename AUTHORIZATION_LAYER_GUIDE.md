# Authorization Validation in Clean Architecture - Microservices

## Overview

Based on your project structure, authorization validation (checking if a user owns a resource) should follow the **Clean Architecture** principles across your microservices.

## Your Current Architecture Layers

```
┌─────────────────────────────────────┐
│  Delivery Layer (gRPC Handlers)     │  ← Entry point, basic validation
├─────────────────────────────────────┤
│  Usecase/Application Layer          │  ← Business logic, authorization
├─────────────────────────────────────┤
│  Domain Layer                       │  ← Business rules, entities
├─────────────────────────────────────┤
│  Repository/Infrastructure Layer    │  ← Data access
└─────────────────────────────────────┘
```

---

## Where to Validate Address Ownership

### **✅ CORRECT LOCATION: Usecase Layer (Application Layer)**

#### Why?

1. **It's Business Logic**: Ownership validation is a business rule, not infrastructure
2. **Requires Data Access**: Must check if the address belongs to the user (repository call)
3. **Precondition for Operations**: Must happen before performing the actual update/delete
4. **Consistent with DDD**: Encapsulates domain logic away from delivery layer
5. **Testable**: Easier to unit test without mocking gRPC context

#### How it works:

```
Handler (Delivery)
  ↓ validates input format
  ↓ extracts user ID from JWT context
  ↓ passes to Usecase
Usecase (Application)
  ↓ checks authorization (user owns address)
  ↓ calls repository for business logic
Repository (Infrastructure)
  ↓ accesses database
```

---

## Implementation Steps

### Step 1: Add Authorization Error to Domain

**File**: `services/UserService/internal/domain/errors.go`

```go
var (
    ErrUserNotFound         = errors.New("user not found")
    ErrInvalidCredentials   = errors.New("invalid email or password")
    ErrHashingPassword      = errors.New("error hashing password")
    ErrUnauthorized         = errors.New("unauthorized: user does not own this resource")
    ErrAddressNotFound      = errors.New("address not found")
)
```

### Step 2: Extract User ID from JWT Context

**File**: `services/UserService/internal/delivery/grpc/handler/user_handler.go`

Add a helper function to extract user ID from JWT token:

```go
func (h *UserGRPCHandler) extractUserIDFromContext(ctx context.Context) (int32, error) {
    // Extract JWT token from gRPC metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return 0, errors.New("missing metadata")
    }

    tokens := md.Get("authorization")
    if len(tokens) == 0 {
        return 0, errors.New("missing authorization token")
    }

    // Parse JWT and extract user ID
    claims, err := h.jwtManager.Verify(tokens[0])
    if err != nil {
        return 0, err
    }

    return int32(claims.UserID), nil
}
```

### Step 3: Update Address Usecase Interface

**File**: `services/UserService/internal/domain/use_cases.go`

Modify the interface to accept userID for authorization:

```go
type AddressUsecaseInterface interface {
    CreateAddress(ctx context.Context, userID int32, req *dto.CreateAddressRequest) (int32, error)
    GetAddressByID(ctx context.Context, userID int32, addressID int32) (Address, error)
    ListAddressesByUserID(ctx context.Context, userID int32) ([]Address, error)
    UpdateAddress(ctx context.Context, userID int32, addressID int32, req *dto.UpdateAddressRequest) error
    DeleteAddress(ctx context.Context, userID int32, addressID int32) error
}
```

### Step 4: Implement Authorization in Usecase

**File**: `services/UserService/internal/usecase/address_usecase.go`

Add authorization check before any operation:

```go
func (a *AddressUsecase) UpdateAddress(
    ctx context.Context,
    userID int32,
    addressID int32,
    req *dto.UpdateAddressRequest,
) error {
    ctx, span := a.tracer.Start(ctx, "AddressUsecase.UpdateAddress")
    defer span.End()

    // Authorization Check - IMPORTANT
    authCtx, authSpan := a.tracer.Start(ctx, "Authorize UpdateAddress")

    address, err := a.addressRepo.GetAddressByID(ctx, uint(addressID))
    if err != nil {
        authSpan.RecordError(err)
        authSpan.SetStatus(codes.Error, err.Error())
        authSpan.End()
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    // Check ownership
    if uint(userID) != address.UserID {
        err := domain.ErrUnauthorized
        authSpan.RecordError(err)
        authSpan.SetStatus(codes.Error, err.Error())
        authSpan.End()

        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    authSpan.End()

    // Continue with update logic
    addressToUpdate := domain.Address{
        ID:      address.ID,
        UserID:  address.UserID,
        Country: req.Country,
        City:    req.City,
        State:   req.State,
        Street:  req.Street,
        ZipCode: req.ZipCode,
    }

    updateAddressCtx, updateAddressSpan := a.tracer.Start(ctx, "addressRepo.UpdateAddress")

    _, err = a.addressRepo.UpdateAddress(updateAddressCtx, addressToUpdate)
    if err != nil {
        updateAddressSpan.RecordError(err)
        updateAddressSpan.SetStatus(codes.Error, err.Error())
        updateAddressSpan.End()

        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    updateAddressSpan.End()

    return nil
}

func (a *AddressUsecase) DeleteAddress(
    ctx context.Context,
    userID int32,
    addressID int32,
) error {
    ctx, span := a.tracer.Start(ctx, "AddressUsecase.DeleteAddress")
    defer span.End()

    // Authorization Check
    authCtx, authSpan := a.tracer.Start(ctx, "Authorize DeleteAddress")

    address, err := a.addressRepo.GetAddressByID(authCtx, uint(addressID))
    if err != nil {
        authSpan.RecordError(err)
        authSpan.SetStatus(codes.Error, err.Error())
        authSpan.End()

        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    if uint(userID) != address.UserID {
        err := domain.ErrUnauthorized
        authSpan.RecordError(err)
        authSpan.SetStatus(codes.Error, err.Error())
        authSpan.End()

        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    authSpan.End()

    err = a.addressRepo.DeleteAddress(ctx, uint(addressID))
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    return nil
}
```

### Step 5: Update Handler to Extract User ID and Pass It

**File**: `services/UserService/internal/delivery/grpc/handler/user_handler.go`

```go
func (h *UserGRPCHandler) UpdateAddress(ctx context.Context, in *pb.UpdateAddressRequest) (*pb.UpdateAddressResponse, error) {
    ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.UpdateAddress")
    defer span.End()

    // Extract user ID from JWT token
    userID, err := h.extractUserIDFromContext(ctx)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    _, validateAddressSpan := h.tracer.Start(ctx, "Validate UpdateAddressRequest")

    updateAddressRequest := dto.UpdateAddressRequest{
        Country: in.GetCountry(),
        City:    in.GetCity(),
        State:   in.GetState(),
        Street:  in.GetStreet(),
        ZipCode: in.GetZipCode(),
    }

    err = h.validate.Struct(updateAddressRequest)
    validationErrors := err.(validator.ValidationErrors)
    if validationErrors != nil {
        validateAddressSpan.RecordError(validationErrors)
        validateAddressSpan.SetStatus(codes.Error, validationErrors.Error())
        validateAddressSpan.End()
        return nil, validationErrors
    }
    validateAddressSpan.End()

    updateAddressCtx, updateAddressSpan := h.tracer.Start(ctx, "Usecase UpdateAddress")

    // Pass userID for authorization check
    err = h.addressUsecase.UpdateAddress(
        updateAddressCtx,
        userID,
        in.GetId(),
        &updateAddressRequest,
    )
    if err != nil {
        updateAddressSpan.RecordError(err)
        updateAddressSpan.SetStatus(codes.Error, err.Error())
        updateAddressSpan.End()
        return nil, err
    }
    updateAddressSpan.End()

    return &pb.UpdateAddressResponse{}, nil
}

func (h *UserGRPCHandler) DeleteAddress(ctx context.Context, in *pb.DeleteAddressRequest) (*pb.DeleteAddressResponse, error) {
    ctx, span := h.tracer.Start(ctx, "UserGRPCHandler.DeleteAddress")
    defer span.End()

    // Extract user ID from JWT token
    userID, err := h.extractUserIDFromContext(ctx)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    deleteAddressCtx, deleteAddressSpan := h.tracer.Start(ctx, "Usecase DeleteAddress")

    // Pass userID for authorization check
    err = h.addressUsecase.DeleteAddress(deleteAddressCtx, userID, in.GetId())
    if err != nil {
        deleteAddressSpan.RecordError(err)
        deleteAddressSpan.SetStatus(codes.Error, err.Error())
        deleteAddressSpan.End()
        return nil, err
    }
    deleteAddressSpan.End()

    return &pb.DeleteAddressResponse{}, nil
}
```

---

## Authorization Validation Architecture Summary

| Layer               | Responsibility            | Example                                                            |
| ------------------- | ------------------------- | ------------------------------------------------------------------ |
| **Delivery (gRPC)** | Extract user from context | `extractUserIDFromContext()`                                       |
| **Usecase**         | Authorize ownership       | `if address.UserID != requestingUserID { return ErrUnauthorized }` |
| **Domain**          | Define error types        | `ErrUnauthorized`                                                  |
| **Repository**      | Just fetch/update data    | No authorization logic here                                        |

---

## Advanced: Middleware Approach (Optional)

For larger projects, create a gRPC interceptor middleware:

**File**: `services/UserService/internal/delivery/grpc/middleware/auth.go`

```go
package middleware

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/metadata"
)

func AuthInterceptor(jwtManager *jwt.JWTManager) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Validate JWT and attach user info to context
        md, _ := metadata.FromIncomingContext(ctx)
        tokens := md.Get("authorization")

        if len(tokens) == 0 {
            return nil, status.Errorf(codes.Unauthenticated, "missing token")
        }

        claims, err := jwtManager.Verify(tokens[0])
        if err != nil {
            return nil, status.Errorf(codes.Unauthenticated, "invalid token")
        }

        // Attach user info to context
        newCtx := context.WithValue(ctx, "userID", claims.UserID)
        return handler(newCtx, req)
    }
}
```

Then in handler:

```go
userID := ctx.Value("userID").(int32)
```

---

## Best Practices

✅ **DO**:

- Validate authorization in the **Usecase layer**
- Check ownership before any resource modification
- Include authorization checks in tracing/monitoring
- Return specific error types (ErrUnauthorized)
- Use gRPC interceptors for JWT validation
- Follow principle of least privilege

❌ **DON'T**:

- Skip authorization checks
- Put authorization logic in handlers (delivery layer)
- Put authorization in repositories (data access)
- Rely only on client-side validation
- Mix authentication (who are you) with authorization (what can you do)

---

## Testing Authorization

```go
func TestUpdateAddressUnauthorized(t *testing.T) {
    // Test when user tries to update address they don't own
    userID := int32(1)
    addressID := int32(99) // belongs to user 2

    err := usecase.UpdateAddress(ctx, userID, addressID, req)

    assert.Equal(t, domain.ErrUnauthorized, err)
}
```
