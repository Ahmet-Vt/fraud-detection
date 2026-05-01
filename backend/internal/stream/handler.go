package stream

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	hub *SSEHub
}

func NewHandler(hub *SSEHub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/stream/transactions", h.StreamTransactions)
}

func (h *Handler) StreamTransactions(c *gin.Context) {
	if h == nil || h.hub == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	stream := h.hub.Subscribe(c.Request.Context())

	// Send a lightweight ready event so the client can switch mode immediately.
	fmt.Fprint(c.Writer, "event: ready\n")
	fmt.Fprint(c.Writer, "data: {}\n\n")
	c.Writer.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case payload, ok := <-stream:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(c.Writer, "event: transaction_evaluated\ndata: %s\n\n", payload); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
