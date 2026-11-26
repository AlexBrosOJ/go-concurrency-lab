package main

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

func TestWaitForSignal(t *testing.T) {
	ready := make(chan struct{})
	close(ready)
	if err := waitForSignal(context.Background(), ready); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Уже отменённый контекст отклоняем даже при готовом событии.
	if err := waitForSignal(ctx, ready); err != context.Canceled {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestParentCancelsWaitingChild(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, cancelParent := context.WithCancelCause(context.Background())
		defer cancelParent(nil)
		child, cancelChild := context.WithCancel(parent)
		defer cancelChild()
		result := make(chan error, 1)
		go func() { result <- waitForSignal(child, nil) }()
		synctest.Wait()

		cause := errors.New("test request ended")
		cancelParent(cause)
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case err := <-result:
			if err != context.Canceled || context.Cause(child) != cause {
				t.Fatalf("got err=%v cause=%v", err, context.Cause(child))
			}
		case <-timer.C:
			t.Fatal("waiter did not respond to parent cancellation")
		}
	})
}

func TestRequestIDFollowsDerivedContexts(t *testing.T) {
	root := context.Background()
	parent := withRequestID(root, "outer")
	child, cancel := context.WithCancel(parent)
	defer cancel()
	overridden := withRequestID(child, "inner")
	for _, tt := range []struct {
		ctx  context.Context
		want string
		ok   bool
	}{
		{root, "", false},
		{child, "outer", true},
		{overridden, "inner", true},
		{withRequestID(root, ""), "", true},
	} {
		if got, ok := requestID(tt.ctx); got != tt.want || ok != tt.ok {
			t.Fatalf("requestID = (%q, %t), want (%q, %t)", got, ok, tt.want, tt.ok)
		}
	}
	cancel()
	if got, ok := requestID(child); got != "outer" || !ok {
		t.Fatalf("metadata lost after cancellation: (%q, %t)", got, ok)
	}
}
