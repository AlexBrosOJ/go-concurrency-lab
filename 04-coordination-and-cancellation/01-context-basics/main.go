package main

import (
	"context"
	"errors"
	"fmt"
)

type requestIDKey struct{}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func requestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok
}

// waitForSignal ждёт событие или завершение контекста вызывающего кода.
func waitForSignal(ctx context.Context, signal <-chan struct{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-signal:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func propagationExample() {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	first, cancelFirst := context.WithCancel(parent)
	defer cancelFirst()
	second, cancelSecond := context.WithCancel(parent)
	defer cancelSecond()

	cancelFirst()
	fmt.Printf("Child cancelled: parent=%v first=%v second=%v\n", parent.Err(), first.Err(), second.Err())
	cancelParent()
	fmt.Printf("Parent cancelled: parent=%v second=%v\n", parent.Err(), second.Err())
}

func eventExample() {
	event := make(chan struct{})
	close(event)
	fmt.Println("Event ready:", waitForSignal(context.Background(), event))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fmt.Println("Wait cancelled:", waitForSignal(ctx, nil))
}

func causeExample() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	child, cancelChild := context.WithCancel(ctx)
	defer cancelChild()
	cancel(errors.New("request withdrawn"))
	fmt.Printf("Child status: %v; cause: %v\n", child.Err(), context.Cause(child))
}

func metadataExample() {
	root := context.Background()
	parent := withRequestID(root, "request-42")
	child, cancel := context.WithCancel(parent)
	defer cancel()
	id, ok := requestID(child)
	fmt.Printf("Request ID: %s; present=%t\n", id, ok)
	_, ok = requestID(root)
	fmt.Println("Root has request ID:", ok)
	cancel()
	id, _ = requestID(child)
	fmt.Println("Request ID after cancellation:", id)
}

func main() {
	fmt.Println("=== 1. Parent and child cancellation ===")
	propagationExample()
	fmt.Println("\n=== 2. Wait for an event ===")
	eventExample()
	fmt.Println("\n=== 3. Cancellation cause ===")
	causeExample()
	fmt.Println("\n=== 4. Request metadata ===")
	metadataExample()
}
