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

	if len(os.Args) < 2 {
		errorLog.Error("please provide a path to the configuration file")
		os.Exit(1)
	}

	configPath := os.Args[1]

	infoLog.Info("welcome to tiny proxy...")
	infoLog.Info("loading configuration file", slog.String("path", configPath))

	// TODO: accept the path from command line flags
	config, err := LoadConfig(configPath)
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
