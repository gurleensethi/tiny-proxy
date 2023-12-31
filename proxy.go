package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gurleensethi/tiny-proxy/middleware"
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

		middlewares := middleware.Middlewares{}
		if server.Http != nil {
			for _, m := range server.Http.Middlewares {
				m, err := middleware.LoadMiddleware(m.Name, p.infoLog, p.errorLog, m.Options)
				if err != nil {
					return err
				}

				middlewares = append(middlewares, m)
			}
		}

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			err := middlewares.ExecuteRequestReceived(middleware.RequestReceivedOptions{
				Request: r,
				Writer:  w,
			})
			if err != nil {
				p.errorLog.Error("failed to execute request received middleware",
					slog.Any("error", err),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			matchResult, err := routeMatcher.Match(r)
			if err != nil {
				p.errorLog.Error("failed to match route", slog.Any("error", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if matchResult == nil {
				w.Header().Set("X-Response-From", "tiny-proxy")
				http.Error(w, "404 page not found", http.StatusNotFound)
				return
			}

			backendURL, _ := url.Parse(matchResult.Route.Backend.URL)

			backendURL.Path = r.URL.Path
			backendURL.RawQuery = r.URL.RawQuery

			// Replace all the placeholders in the rewrite path
			if matchResult.Route.Rewrite != nil {
				backendURL.Path = *matchResult.Route.Rewrite
				for k, v := range matchResult.Match {
					backendURL.Path = strings.ReplaceAll(backendURL.Path, k, v)
				}
			}

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

			middlewares.ExecutePreProxyRequest(middleware.PreProxyRequestOptions{
				Request:      r,
				Writer:       w,
				ProxyRequest: proxyRequest,
			})

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

			err = middlewares.ExecutePreResponse(middleware.PreResponseOptions{
				Request: r,
				Writer:  w,
			})
			if err != nil {
				p.errorLog.Error("failed to execute pre response middleware", slog.Any("error", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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

			err = middlewares.ExecutePostResponse(middleware.PostResponseOptions{
				Request: r,
				Writer:  w,
			})
			if err != nil {
				p.errorLog.Error("failed to execute pre response middleware", slog.Any("error", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		if server.Http != nil {
			var handler http.Handler = mux

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
