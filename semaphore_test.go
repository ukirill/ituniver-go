package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSemaphore(t *testing.T) {
	t.Run("Succesful semaphore creation", func(t *testing.T) {
		cap := 3
		s := newSemaphore(cap)
		assert.Equal(t, cap, len(s.tickets))
	})

	t.Run("New semaphore have all tickets", func(t *testing.T) {
		cap := 2
		s := newSemaphore(cap)
		ctx := context.Background()
		assert.NoError(t, s.waitOne(ctx))
		assert.NoError(t, s.waitOne(ctx))
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		assert.Error(t, s.waitOne(ctx))
	})

	t.Run("Semaphore with negative capacity fails", func(t *testing.T) {
		cap := -1
		assert.Panics(t, func() {
			newSemaphore(cap)
		})
	})
}

func TestWaitOne(t *testing.T) {
	t.Run("Non-empty semaphore returns unblockably on canceled context", func(t *testing.T) {
		cap := 2
		s := newSemaphore(cap)
		ctx := context.Background()
		assert.NoError(t, s.waitOne(ctx))
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		assert.NoError(t, s.waitOne(ctx))
	})
}

func TestRelease(t *testing.T) {
	t.Run("Successfull release of non-empty semaphore", func(t *testing.T) {
		cap := 1
		s := newSemaphore(cap)
		ctx := context.Background()
		assert.NoError(t, s.waitOne(ctx))
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		assert.Error(t, s.waitOne(ctx))
		s.release()
		assert.NoError(t, s.waitOne(ctx))
	})

	t.Run("Full semaphore panics on release attempt", func(t *testing.T) {
		cap := 1
		s := newSemaphore(cap)
		assert.Panics(t, func() {
			s.release()
		})
	})
}
