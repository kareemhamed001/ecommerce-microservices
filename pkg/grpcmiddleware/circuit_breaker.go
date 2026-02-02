package grpcmiddleware

import (
	"context"
	"time"

	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CircuitBreakerConfig struct {
	Enabled      bool
	MaxRequests  uint32
	Interval     time.Duration
	Timeout      time.Duration
	FailureRatio float64
	MinRequests  uint32
}

func CircuitBreakerUnaryClientInterceptor(name string, cfg CircuitBreakerConfig) grpc.UnaryClientInterceptor {
	if !cfg.Enabled {
		return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if cfg.MinRequests > 0 && counts.Requests < cfg.MinRequests {
				return false
			}
			if counts.Requests == 0 {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= cfg.FailureRatio
		},
		IsSuccessful: func(err error) bool {
			return !isBreakerFailure(err)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warnf("event=circuit_breaker_state_change name=%s from=%s to=%s", name, from.String(), to.String())
		},
	}

	cb := gobreaker.NewCircuitBreaker(settings)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})
		return err
	}
}

func isBreakerFailure(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return true
	}

	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Internal:
		return true
	default:
		return false
	}
}
