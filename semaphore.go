package main

import (
	"context"
	"fmt"
)

type semaphore struct {
	tickets chan struct{}
}

// newSemaphore creates new semaphore with capacity cap (must be positive).
func newSemaphore(cap int) *semaphore {
	tickets := make(chan struct{}, cap)
	for i := 0; i < cap; i++ {
		tickets <- struct{}{}
	}
	return &semaphore{tickets: tickets}
}

// waitOne takes one ticket from semaphore.
// Blocks while no tickets, could take ticket in unblocking way on done context.
func (s *semaphore) waitOne(ctx context.Context) error {
	select {
	case <-s.tickets:
		return nil
	case <-ctx.Done():
		select {
		case <-s.tickets:
			return nil
		default:
			return fmt.Errorf("Semaphore is not acquired, operation canceled : %w", ctx.Err())
		}
	}
}

// release returns one ticket to semaphore. Panics if all tickets returned already.
func (s *semaphore) release() {
	select {
	case s.tickets <- struct{}{}:
		return
	default:
		panic("all semaphore tickets are already returned")
	}
}
