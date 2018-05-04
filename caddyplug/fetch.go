package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var fetchers = map[string]pluginFetcher{
	"http":   fetcherFunc(fetchHTTPPlugins),
	"dns":    fetcherFunc(fetchDNSPlugins),
	"others": fetcherFunc(fetchOtherPlugins),
}

type pluginFetcher interface {
	FetchPlugins() ([]Plugin, error)
}

type fetcherFunc func() ([]Plugin, error)

func (f fetcherFunc) FetchPlugins() ([]Plugin, error) { return f() }

func fetchHTTPPlugins() ([]Plugin, error) {
	var plugins []Plugin
	fset := token.NewFileSet()
	file := filepath.Join(goPath(), "src", directivesFile)
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return plugins, err
	}
	node, ok := f.Scope.Lookup("directives").Decl.(*ast.ValueSpec)
	if !ok {
		return plugins, fmt.Errorf("parsing error")
	}

	cmap := ast.NewCommentMap(fset, f, f.Comments)
	c := node.Values[0].(*ast.CompositeLit)
	for _, m := range c.Elts {
		if cm, ok := cmap[m]; ok {
			pkg := strings.TrimSpace(cm[len(cm)-1].Text())
			directive, err := strconv.Unquote(m.(*ast.BasicLit).Value)
			if err != nil {
				return plugins, err
			}
			// asserting that the comment word count is 1 may not be the best way
			// to confirm it is a repo path.
			if len(strings.Fields(pkg)) == 1 {
				plugin := Plugin{
					Name:    directive,
					Package: pkg,
					Type:    "http",
				}
				plugins = append(plugins, plugin)
			}
		}
	}
	return plugins, nil
}

func fetchDNSPlugins() ([]Plugin, error) {
	var plugins []Plugin
	srcDir := filepath.Join(goPath(), "src", dnsProvidersPackage)
	d, err := os.Open(srcDir)
	if err != nil {
		return plugins, err
	}
	stats, err := d.Readdir(-1)
	if err != nil {
		return plugins, err
	}
	for _, stat := range stats {
		provider := stat.Name()
		// skip hidden files
		if strings.HasPrefix(provider, ".") || !stat.IsDir() {
			continue
		}
		plugin := Plugin{
			Name:    provider,
			Package: path.Join(dnsProvidersPackage, provider),
			Type:    "dns",
		}
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}

// TODO: this needs to be dynamic.
func fetchOtherPlugins() ([]Plugin, error) {
	return []Plugin{
		{
			Type:    "server",
			Name:    "net",
			Package: "github.com/pieterlouw/caddy-net/caddynet",
		},
		{
			Type:    "server",
			Name:    "dns",
			Package: "github.com/coredns/coredns/core/dnsserver",
		},
		{
			Type:    "caddyfile",
			Name:    "docker",
			Package: "github.com/lucaslorentz/caddy-docker-proxy/plugin",
		},
		{
			Type:    "hook",
			Name:    "service",
			Package: "github.com/hacdias/caddy-service",
		},
	}, nil
}
