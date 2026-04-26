package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

type separatorHandler struct {
	handler slog.Handler
	writer  io.Writer
	mu      *sync.Mutex
}

func (h *separatorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *separatorHandler) Handle(ctx context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err := h.writer.Write([]byte("\n==========================================================\n")); err != nil {
		return err
	}

	return h.handler.Handle(ctx, record)
}

func (h *separatorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &separatorHandler{
		handler: h.handler.WithAttrs(attrs),
		writer:  h.writer,
		mu:      h.mu,
	}
}

func (h *separatorHandler) WithGroup(name string) slog.Handler {
	return &separatorHandler{
		handler: h.handler.WithGroup(name),
		writer:  h.writer,
		mu:      h.mu,
	}
}

func SetupLogging(path string) (func() error, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				attr.Value = slog.StringValue(attr.Value.Time().Format("15:04"))
			}

			if attr.Key == slog.SourceKey {
				if source, ok := attr.Value.Any().(*slog.Source); ok {
					source.Function = ""
				}
			}

			return attr
		},
	}

	textHandler := slog.NewTextHandler(file, opts)
	logger := slog.New(&separatorHandler{
		handler: textHandler,
		writer:  file,
		mu:      &sync.Mutex{},
	})

	slog.SetDefault(logger)

	return file.Close, nil
}
