package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiosoft/errs"
)

type dependencies []struct {
	name       string
	importPath string
	installed  bool
	updateFunc func() error
}

var packageDependecies = dependencies{
	{name: "caddy", importPath: "github.com/caddyserver/caddy", updateFunc: fetchCaddy},
	{name: "dnsproviders", importPath: "github.com/caddyserver/dnsproviders", updateFunc: fetchDNSProviders},
	{name: "hook.pluginloader", importPath: "github.com/abiosoft/caddyplug", updateFunc: fetchCaddyPlug},
}

func (d dependencies) installed() bool {
	for _, dep := range d {
		if !dep.installed {
			return false
		}
	}
	return true
}

func (d dependencies) missing() string {
	var s []string
	for i := range d {
		if !d[i].installed {
			s = append(s, d[i].name)
		}
	}
	return strings.Join(s, ", ")
}

func (d dependencies) check() bool {
	var buf bytes.Buffer
	err := shellCmd{Stdout: &buf, Dir: goPath(), Silent: true}.
		run("go", "list", "./...")
	if err != nil {
		return false
	}
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		for i := range d {
			if strings.HasPrefix(line, d[i].importPath) {
				d[i].installed = true
			}
		}
	}
	return d.installed()
}

func (d dependencies) update() error {
	if d.check() {
		return nil
	}
	log("fetching missing dependencies:", d.missing())
	var e errs.Group
	for _, dep := range d {
		if !dep.installed {
			e.Add(dep.updateFunc)
		}
	}
	e.Add(func() error {
		log("done fetching depedencies.")
		log()
		return nil
	})
	return e.Exec()
}

func fetchCaddy() error {
	var e errs.Group
	e.Add(func() error {
		return shellCmd{}.run("go", "get", "-d", "github.com/caddyserver/caddy")
	})
	caddyPath := filepath.Join(goPath(), "src", "github.com/caddyserver/caddy")
	e.Add(func() error {
		return shellCmd{Dir: caddyPath, Silent: true}.
			run("git", "checkout", caddyVersion)
	})
	e.Add(func() error {
		return install("github.com/caddyserver/caddy")
	})
	return e.Exec()
}

func fetchCaddyPlug() error {
	return install("github.com/abiosoft/caddyplug")
}

func fetchDNSProviders() error {
	var e errs.Group
	dnsDir := filepath.Join(goPath(), "src", dnsProvidersPackage)
	if _, err := os.Stat(dnsDir); err != nil {
		e.Add(func() error {
			return shellCmd{}.run("git", "clone", "https://"+dnsProvidersPackage, dnsDir)
		})
	}
	dnsProvidersPath := filepath.Join(goPath(), "src", dnsProvidersPackage)
	e.Add(func() error {
		return shellCmd{Dir: dnsProvidersPath, Silent: true}.
			run("git", "checkout", dnsProvidersVersion)
	})
	return e.Exec()
}

func fetchDependencies() error {
	return packageDependecies.update()
}

func log(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}
