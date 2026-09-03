package mock

import (
	"context"
	"log/slog"
	"sync"
)

type MockHandler struct {
	mu      sync.Mutex
	Records []slog.Record
}

func NewMockLogger() *slog.Logger {
	return slog.New(NewMockHandler())
}

func NewMockHandler() *MockHandler {
	return &MockHandler{}
}

func (h *MockHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *MockHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Records = append(h.Records, r.Clone())
	return nil
}

func (h *MockHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *MockHandler) WithGroup(string) slog.Handler { return h }
