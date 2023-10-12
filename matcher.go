package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type RouteMatcher interface {
	Match(r *http.Request) (*MatchResult, error)
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

type MatchResult struct {
	Route *RouteConfig
	Match map[string]string
}

func (m *RegexRouteMatcher) Match(r *http.Request) (*MatchResult, error) {
	for _, route := range m.routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			match := make(map[string]string)

			for i, v := range matches[1:] {
				match[fmt.Sprintf("{%d}", i+1)] = v
			}

			return &MatchResult{
				Route: &route.config,
				Match: match,
			}, nil
		}

	}

	return nil, nil
}
