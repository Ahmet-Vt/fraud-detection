package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"fraud-detection/internal/transaction"
)

const sseClientBuffer = 256

type SseEnvelope struct {
	Event   string
	Payload []byte
}

type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan SseEnvelope]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan SseEnvelope]struct{}),
	}
}

func (h *SSEHub) Subscribe(ctx context.Context) <-chan SseEnvelope {
	ch := make(chan SseEnvelope, sseClientBuffer)

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

func (h *SSEHub) broadcast(env SseEnvelope) {
	if h == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- env:
		default:
		}
	}
}

func (h *SSEHub) PublishTransactionEvaluated(_ context.Context, evt transaction.TransactionEvaluatedEvent) error {
	if h == nil {
		return fmt.Errorf("sse hub is nil")
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal evaluated event: %w", err)
	}

	h.broadcast(SseEnvelope{Event: "transaction_evaluated", Payload: payload})
	return nil
}

func (h *SSEHub) PublishSse(_ context.Context, event string, payload []byte) error {
	if h == nil {
		return fmt.Errorf("sse hub is nil")
	}
	if event == "" {
		return fmt.Errorf("sse event name is empty")
	}
	h.broadcast(SseEnvelope{Event: event, Payload: payload})
	return nil
}
