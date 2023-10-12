package main

import (
	"errors"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type BackendConfig struct {
	URL string `yaml:"url"`
}

func (c BackendConfig) Validate() error {
	if c.URL == "" {
		return errors.New("http.routes[].backend.url is required")
	}

	_, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	return nil
}

type RouteConfig struct {
	Path    string        `yaml:"path"`
	Rewrite *string       `yaml:"rewrite"`
	Backend BackendConfig `yaml:"backend"`
}

func (c RouteConfig) Validate() error {
	if c.Path == "" {
		return errors.New("http.routes[].path is required")
	}

	return nil
}

type MiddlewareConfig struct {
	Name    string                 `yaml:"name"`
	Options map[string]interface{} `yaml:"options"`
}

func (c MiddlewareConfig) Validate() error {
	if c.Name == "" {
		return errors.New("http.middlewares[].name is required")
	}

	return nil
}

type HttpConfig struct {
	Host        string             `yaml:"host"`
	Port        int                `yaml:"port"`
	Middlewares []MiddlewareConfig `yaml:"middlewares"`
	Routes      []RouteConfig      `yaml:"routes"`
}

func (c HttpConfig) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}

	if c.Port == 0 {
		return errors.New("port is required")
	}

	if c.Port < 0 || c.Port > 65535 {
		return errors.New("port must be between 0 and 65535")
	}

	for _, middleware := range c.Middlewares {
		err := middleware.Validate()
		if err != nil {
			return err
		}
	}

	for _, route := range c.Routes {
		err := route.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

type ServerConfig struct {
	Http *HttpConfig `yaml:"http"`
}

func (c ServerConfig) Validate() error {
	err := c.Http.Validate()
	if err != nil {
		return err
	}

	return nil
}

type ProxyConfig struct {
	Servers []ServerConfig `yaml:"servers"`
}

func (c ProxyConfig) Validate() error {
	for _, server := range c.Servers {
		err := server.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadConfig(filePath string) (*ProxyConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ProxyConfig

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
