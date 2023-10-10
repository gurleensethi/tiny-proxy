package middleware

import (
	"encoding/json"
	"log/slog"
	"strings"
)

func NewCORSMiddlewareFromOptions(infoLogger, errLogger *slog.Logger, opts map[string]any) (*CORSMiddleware, error) {
	m := &CORSMiddleware{
		InfoLogger:  infoLogger,
		ErrorLogger: errLogger,
	}

	marshaledOpts, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(marshaledOpts, &m.Options)
	if err != nil {
		return nil, err
	}

	err = m.Options.Validate()
	if err != nil {
		return nil, err
	}

	return m, nil
}

var _ Middleware = (*CORSMiddleware)(nil)

type CORSMiddlewareOptions struct {
	All          bool     `json:"all"`
	AllowOrigins []string `json:"allowOrigins"`
	AllowMethods []string `json:"allowMethods"`
	AllowHeaders []string `json:"allowHeaders"`
}

func (opts *CORSMiddlewareOptions) Validate() error {
	return nil
}

type CORSMiddleware struct {
	InfoLogger  *slog.Logger
	ErrorLogger *slog.Logger
	Options     CORSMiddlewareOptions `json:"options"`
}

func (m *CORSMiddleware) PostResponse(opts PostResponseOptions) error {
	return nil
}

func (m *CORSMiddleware) PreResponse(opts PreResponseOptions) error {
	if m.Options.All {
		opts.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		opts.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		opts.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	} else {
		if len(m.Options.AllowOrigins) > 0 {
			opts.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(m.Options.AllowOrigins, ","))
		}

		if len(m.Options.AllowMethods) > 0 {
			opts.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(m.Options.AllowMethods, ","))
		}

		if len(m.Options.AllowHeaders) > 0 {
			opts.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(m.Options.AllowHeaders, ","))
		}
	}

	return nil
}

func (m *CORSMiddleware) RequestReceived(opts RequestReceivedOptions) error {
	return nil
}
