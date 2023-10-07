package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()

	infoLog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	errorLog := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))

	infoLog.Info("welcome to tiny proxy...")

	// TODO: accept the path from command line flags
	config, err := LoadConfig("proxy-config.yaml")
	if err != nil {
		errorLog.Error("failed to load config, please check your configuration file")
		panic(err)
	}

	proxy := New(config, infoLog, errorLog)

	err = proxy.Start(ctx)
	if err != nil {
		panic(err)
	}
}
