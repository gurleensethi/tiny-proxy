package main

import (
	"net/http"
	"regexp"
	"strings"
)

type RouteMatcher interface {
	Match(r *http.Request) (*RouteConfig, error)
}

type RegexRoute struct {
	regex  *regexp.Regexp
	config RouteConfig
}

type RegexRouteMatcher struct {
	routes []RegexRoute
}

var _ RouteMatcher = (*RegexRouteMatcher)(nil)

func NewRegexRouteMatcher() *RegexRouteMatcher {
	return &RegexRouteMatcher{}
}

func (m *RegexRouteMatcher) LoadRoutes(routes ...RouteConfig) error {
	for _, r := range routes {
		pathRegex := r.Path
		if pathRegex == "" || pathRegex[0] != '^' {
			pathRegex = "^" + pathRegex
		}
		pathRegex = strings.ReplaceAll(pathRegex, "/", "\\/")

		regex, err := regexp.Compile(pathRegex)
		if err != nil {
			return err
		}

		m.routes = append(m.routes, RegexRoute{
			regex:  regex,
			config: r,
		})
	}

	return nil
}

func (m *RegexRouteMatcher) Match(r *http.Request) (*RouteConfig, error) {
	for _, route := range m.routes {
		if route.regex.MatchString(r.URL.Path) {
			return &route.config, nil
		}
	}

	return nil, nil
}
