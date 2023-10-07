package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

func New(c *ProxyConfig, infoLog *slog.Logger, errorLog *slog.Logger) *Proxy {
	return &Proxy{
		config:   c,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

type Proxy struct {
	config   *ProxyConfig
	infoLog  *slog.Logger
	errorLog *slog.Logger
}

func (p *Proxy) Start(ctx context.Context) error {
	for _, s := range p.config.Servers {
		server := s

		mux := http.NewServeMux()
		routeMatcher := NewRegexRouteMatcher()
		err := routeMatcher.LoadRoutes(server.Http.Routes...)
		if err != nil {
			return err
		}

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			route, err := routeMatcher.Match(r)
			if err != nil {
				p.errorLog.Error("failed to match route", slog.Any("error", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if route == nil {
				w.Header().Set("X-Response-From", "tiny-proxy")
				http.Error(w, "404 page not found", http.StatusNotFound)
				return
			}

			backendURL, _ := url.Parse(route.Backend.URL)

			if server.Http.Log {
				p.infoLog.Info("proxying request to backend",
					slog.String("path", backendURL.Path),
					slog.String("method", r.Method),
					slog.String("host", backendURL.Host),
					slog.String("scheme", backendURL.Scheme),
					slog.String("to", backendURL.String()),
				)
			}

			backendURL.Path = r.URL.Path
			backendURL.RawQuery = r.URL.RawQuery

			proxyRequest := &http.Request{
				Method:        r.Method,
				URL:           backendURL,
				Body:          r.Body,
				Header:        r.Header,
				ContentLength: r.ContentLength,
				Form:          r.Form,
				PostForm:      r.PostForm,
				MultipartForm: r.MultipartForm,
				Trailer:       r.Trailer,
				RemoteAddr:    r.RemoteAddr,
				TLS:           r.TLS,
			}

			resp, err := http.DefaultClient.Do(proxyRequest)
			if err != nil {
				p.errorLog.Error("failed to proxy request",
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
					slog.String("host", r.Host),
					slog.String("scheme", r.URL.Scheme),
					slog.Any("error", err),
				)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			for key, values := range resp.Header {
				for _, v := range values {
					w.Header().Set(key, v)
				}
			}

			w.WriteHeader(resp.StatusCode)

			buffer := make([]byte, 4*1024)
			for {
				n, err := resp.Body.Read(buffer)

				w.Write(buffer[:n])

				if err == io.EOF {
					break
				} else if err != nil {
					p.errorLog.Error("failed to read response body",
						slog.Any("error", err),
					)

					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			err = resp.Body.Close()
			if err != nil {
				p.errorLog.Error("failed to close response body",
					slog.Any("error", err),
				)
			}
		})

		if server.Http != nil {
			var handler http.Handler = mux

			if server.Http.Log {
				handler = requestLogger(p.infoLog, mux)
			}

			p.infoLog.Info("starting http server...",
				slog.String("host", server.Http.Host),
				slog.Int("port", server.Http.Port))

			return http.ListenAndServe(
				// TODO: start http server only if http config is not nil
				fmt.Sprintf("%s:%d", server.Http.Host, server.Http.Port),
				handler,
			)
		}

	}

	return errors.New("no server config found")
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request received",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("host", r.Host),
			slog.String("scheme", r.URL.Scheme),
			slog.String("remoteAddr", r.RemoteAddr),
		)

		next.ServeHTTP(w, r)
	})
}
