package main

import (
	"context"
	"fmt"
)

type semaphore struct {
	tickets chan struct{}
}

func newSemaphore(cap int) *semaphore {
	tickets := make(chan struct{}, cap)
	for i := 0; i < cap; i++ {
		tickets <- struct{}{}
	}
	return &semaphore{tickets: tickets}
}

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

func (s *semaphore) release() {
	select {
	case s.tickets <- struct{}{}:
		return
	default:
		panic("all semaphore tickets are already returned")
	}
}
