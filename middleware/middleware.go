package middleware

import (
	"errors"
	"log/slog"
	"net/http"
)

type RequestReceivedOptions struct {
	Request *http.Request
	Writer  http.ResponseWriter
}

type PreProxyRequestOptions struct {
	Request      *http.Request
	ProxyRequest *http.Request
	Writer       http.ResponseWriter
}

type PreResponseOptions struct {
	Request *http.Request
	Writer  http.ResponseWriter
}

type PostResponseOptions struct {
	Request *http.Request
	Writer  http.ResponseWriter
}

type Middleware interface {
	RequestReceived(opts RequestReceivedOptions) error

	// PreProxyRequest is called after a match for the request is found and before the request is proxied.
	PreProxyRequest(opts PreProxyRequestOptions) error

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

func (m Middlewares) ExecutePreProxyRequest(opts PreProxyRequestOptions) error {
	for _, middleware := range m {
		err := middleware.PreProxyRequest(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadMiddleware(name string, infoLogger, errorLogger *slog.Logger, opts map[string]any) (Middleware, error) {
	switch name {
	case "log":
		return NewLogMiddlewareFromOptions(
			infoLogger.With(slog.String("middleware", name)),
			errorLogger.With(slog.String("middleware", name)),
			opts,
		)
	case "cors":
		return NewCORSMiddlewareFromOptions(
			infoLogger.With(slog.String("middleware", name)),
			errorLogger.With(slog.String("middleware", name)),
			opts,
		)
	default:
		return nil, errors.New("unknown middleware: " + name)
	}
}
