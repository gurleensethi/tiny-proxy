package middleware

import "log/slog"

func NewLogMiddlewareFromOptions(infoLogger, errLogger *slog.Logger, opts map[string]any) (*LogMiddleware, error) {
	m := &LogMiddleware{
		InfoLogger:  infoLogger,
		ErrorLogger: errLogger,
	}

	m.Options.Load(opts)

	err := m.Options.Validate()
	if err != nil {
		return nil, err
	}

	return m, nil
}

type LogMiddleware struct {
	InfoLogger  *slog.Logger
	ErrorLogger *slog.Logger
	Options     LogMiddlewareOptions `json:"options"`
}

func (m *LogMiddleware) PreProxyRequest(opts PreProxyRequestOptions) error {
	r := opts.ProxyRequest

	m.InfoLogger.Info("proxying request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("host", r.Host),
		slog.String("scheme", r.URL.Scheme),
		slog.String("remoteAddr", r.RemoteAddr),
	)

	return nil
}

func (*LogMiddleware) PostResponse(opts PostResponseOptions) error {
	return nil
}

func (*LogMiddleware) PreResponse(opts PreResponseOptions) error {
	return nil
}

func (m *LogMiddleware) RequestReceived(opts RequestReceivedOptions) error {
	r := opts.Request

	m.InfoLogger.Info("request received",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("host", r.Host),
		slog.String("scheme", r.URL.Scheme),
		slog.String("remoteAddr", r.RemoteAddr),
	)

	return nil
}

var _ Middleware = (*LogMiddleware)(nil)

type LogMiddlewareOptions struct {
	Format string `json:"format"`
}

func (opts *LogMiddlewareOptions) Load(o map[string]any) {
	if v, ok := o["format"].(string); ok {
		opts.Format = v
	}
}

func (opts LogMiddlewareOptions) Validate() error {
	return nil
}
