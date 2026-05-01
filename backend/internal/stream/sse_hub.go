package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"fraud-detection/internal/transaction"
)

type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan []byte]struct{}),
	}
}

func (h *SSEHub) Subscribe(ctx context.Context) <-chan []byte {
	ch := make(chan []byte, 16)

	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.mu.Lock()
		if _, exists := h.clients[ch]; exists {
			delete(h.clients, ch)
			close(ch)
		}
		h.mu.Unlock()
	}()

	return ch
}

func (h *SSEHub) PublishTransactionEvaluated(_ context.Context, evt transaction.TransactionEvaluatedEvent) error {
	if h == nil {
		return fmt.Errorf("sse hub is nil")
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal evaluated event: %w", err)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- payload:
		default:
		}
	}

	return nil
}
