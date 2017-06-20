package caddyplug

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mholt/caddy"
)

const (
	noPlugin      = "no plugins found, use caddyplug to add plugins"
	rebuildPlugin = "error occured while loading some plugins, try reinstalling them"
)

func init() {
	caddy.RegisterEventHook("pluginloader", hook)
}

var hook caddy.EventHook = func(event caddy.EventName, info interface{}) error {
	switch event {
	case caddy.StartupEvent:
		if runtime.GOOS != "linux" {
			log.Println("pluginloader is only supported on Linux")
			return nil
		}
		count := 0
		if httpPlugins := loadedPlugins["http"]; len(httpPlugins) > 0 {
			fmt.Println("http plugins loaded:", strings.Join(httpPlugins, ", "))
			count += len(httpPlugins)
		}
		if dnsPlugins := loadedPlugins["dns"]; len(dnsPlugins) > 0 {
			fmt.Println("dns plugins loaded:", strings.Join(dnsPlugins, ", "))
			count += len(dnsPlugins)
		}
		if loadError {
			fmt.Println(rebuildPlugin)
		} else if count == 0 {
			fmt.Println(noPlugin)
		}
	}
	return nil
}

var (
	loadedPlugins = map[string][]string{
		"http": []string{},
		"dns":  []string{},
	}
	loadError bool
)

// PluginsDir is the directory for built plugins.
func PluginsDir() string {
	return filepath.Join(LibDir(), "plugins")
}

// LibDir is the directory for caddy plugin loader resources.
func LibDir() string {
	return filepath.Join(os.Getenv("HOME"), "lib", "caddy")
}
