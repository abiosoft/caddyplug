package caddyplug

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/abiosoft/caddyplug/internal"
)

func init() {
	loadPlugins("http")
	loadPlugins("dns")
}

func loadPlugins(pluginType string) {
	dir, err := os.Open(filepath.Join(internal.PluginsDir(), pluginType))
	if err != nil {
		return
	}
	plugins, err := dir.Readdirnames(-1)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, pluginLib := range plugins {
		if !strings.HasSuffix(pluginLib, ".so") {
			continue
		}
		pluginName := strings.TrimSuffix(pluginLib, ".so")
		pluginFile := filepath.Join(dir.Name(), pluginLib)
		_, err := plugin.Open(pluginFile)
		if err != nil {
			fmt.Println("error loading "+pluginName+": ", err)
			loadError = true
			continue
		}
		loadedPlugins[pluginType] = append(loadedPlugins[pluginType], pluginName)
	}
}
