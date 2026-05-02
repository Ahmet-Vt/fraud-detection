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

	fmt.Fprint(c.Writer, "retry: 2000\n\n")
	c.Writer.Flush()

	fmt.Fprint(c.Writer, "event: ready\n")
	fmt.Fprint(c.Writer, "data: {}\n\n")
	c.Writer.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg, ok := <-stream:
			if !ok {
				return
			}
			ev := msg.Event
			if ev == "" {
				ev = "message"
			}
			if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", ev, msg.Payload); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
