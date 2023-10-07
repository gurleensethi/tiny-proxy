package middleware

import (
	"errors"
	"log/slog"
	"net/http"
)

type RequestReceivedOptions struct {
	InfoLogger  *slog.Logger
	ErrorLogger *slog.Logger
	Request     *http.Request
	Writer      http.ResponseWriter
}

type PreResponseOptions struct {
	InfoLogger  *slog.Logger
	ErrorLogger *slog.Logger
	Request     *http.Request
	Writer      http.ResponseWriter
}

type PostResponseOptions struct {
	InfoLogger  *slog.Logger
	ErrorLogger *slog.Logger
	Request     *http.Request
	Writer      http.ResponseWriter
}

type Middleware interface {
	RequestReceived(opts RequestReceivedOptions) error
	PreResponse(opts PreResponseOptions) error
	PostResponse(opts PostResponseOptions) error
}

type Middlewares []Middleware

func (m Middlewares) ExecuteRequestReceived(opts RequestReceivedOptions) error {
	for _, middleware := range m {
		err := middleware.RequestReceived(opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m Middlewares) ExecutePreResponse(opts PreResponseOptions) error {
	for _, middleware := range m {
		err := middleware.PreResponse(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m Middlewares) ExecutePostResponse(opts PostResponseOptions) error {
	for _, middleware := range m {
		err := middleware.PostResponse(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadMiddleware(name string, opts map[string]any) (Middleware, error) {
	switch name {
	case "log":
		return NewLogMiddlewareFromOptions(opts)
	default:
		return nil, errors.New("unknown middleware: " + name)
	}
}
