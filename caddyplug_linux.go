package caddyplug

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func init() {
	loadPlugins("http")
	loadPlugins("dns")
}

func loadPlugins(pluginType string) {
	dir, err := os.Open(filepath.Join(pluginsDir(), pluginType))
	if err != nil {
		fmt.Println(err)
		return
	}
	plugins, err := dir.Readdirnames(-1)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, pluginName := range plugins {
		if !strings.HasSuffix(pluginName, ".so") {
			continue
		}
		pluginFile := filepath.Join(dir.Name(), pluginName)
		_, err := plugin.Open(pluginFile)
		if err != nil {
			fmt.Println(err)
		}
	}
}
