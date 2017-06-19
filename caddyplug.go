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

const noPlugin = "no plugins found, use caddyplug to add plugins"

func init() {
	caddy.RegisterEventHook("pluginloader", hook)
}

func pluginsDir() string {
	return filepath.Join(os.Getenv("HOME"), "lib", "caddy")
}

func hook(event caddy.EventName, info interface{}) error {
	switch event {
	case caddy.StartupEvent:
		if runtime.GOOS != "linux" {
			log.Println("pluginloader is only supported on Linux")
			return nil
		}
		if stat, err := os.Stat(pluginsDir()); err != nil || !stat.IsDir() {
			fmt.Println(noPlugin)
			return nil
		}
		count := 0
		if httpPlugins := listPlugins("http"); len(httpPlugins) > 0 {
			fmt.Println("http plugins loaded:", strings.Join(httpPlugins, ", "))
			count += len(httpPlugins)
		}
		if dnsPlugins := listPlugins("dns"); len(dnsPlugins) > 0 {
			fmt.Println("dns plugins loaded:", strings.Join(dnsPlugins, ", "))
			count += len(dnsPlugins)
		}
		if count == 0 {
			fmt.Println(noPlugin)
		}
	}
	return nil
}

func listPlugins(pluginType string) []string {
	var plugins []string
	dir, err := os.Open(filepath.Join(pluginsDir(), pluginType))
	defer dir.Close()

	if err != nil {
		return plugins
	}
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return plugins
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".so") {
			continue
		}
		plugins = append(plugins, strings.TrimSuffix(name, ".so"))
	}
	return plugins
}
