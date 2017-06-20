package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/abiosoft/errs"
	"github.com/fatih/color"
)

var (
	commands = map[string]func([]string){
		"install":       installPlugins,
		"uninstall":     uninstallPlugins,
		"list":          listPlugins,
		"install-caddy": installCaddy,
		"help":          func([]string) { usage() },
	}

	successMark = color.GreenString("✓")
	errorMark   = color.RedString("✗")
)

func installPlugins(pluginNames []string) {
	var outputs []string
	var errored bool
	for _, pluginName := range pluginNames {
		plugin, ok := plugins[pluginName]
		if !ok {
			outputs = append(outputs, " "+errorMark+" "+pluginName+" - plugin not found")
			errored = true
			continue
		}
		fmt.Println("installing", plugin.Name+"...")
		if err := plugin.Build(); err != nil {
			fmt.Println(err)
			errored = true
			outputs = append(outputs, " "+errorMark+" "+plugin.Name)
		} else {
			outputs = append(outputs, " "+successMark+" "+plugin.Name)
		}
	}
	if len(outputs) > 0 {
		for _, p := range outputs {
			fmt.Println(p)
		}
	}
	if errored {
		os.Exit(1)
	}
}

func uninstallPlugins(pluginNames []string) {
	var success []string
	var failures []string
	for _, pluginName := range pluginNames {
		plugin, ok := plugins[pluginName]
		if !ok {
			fmt.Println("plugin not found", pluginName)
			return
		}
		if err := plugin.Remove(); err != nil {
			fmt.Println(err)
			failures = append(failures, plugin.Name)
		} else {
			success = append(success, plugin.Name)
		}
	}
	if len(success) > 0 {
		fmt.Println("Uninstalled:")
		fmt.Println(" ", strings.Join(success, ", "))
	}
	if len(failures) > 0 {
		fmt.Println("Failed to uninstall:")
		fmt.Println(" ", strings.Join(failures, ", "))
	}
}

func listPlugins([]string) {
	for pluginType, fetcher := range fetchers {
		plugins, err := fetcher.FetchPlugins()
		if err != nil {
			fmt.Println(err)
			return
		}
		sort.Slice(plugins, func(i, j int) bool {
			if plugins[i].Installed() != plugins[j].Installed() {
				return plugins[i].Installed()
			}
			return plugins[i].Name < plugins[j].Name
		})
		if len(plugins) > 1 {
			fmt.Println(pluginType + ":")
		}
		for _, plugin := range plugins {
			check := "  "
			if plugin.Installed() {
				check = " ✓"
			}
			fmt.Println(check, plugin.Name)
		}
	}
}

const (
	pluginLoaderFile = "caddy/caddymain/pluginloader.go"
	pluginLoaderSrc  = `package caddymain
	import _ "github.com/abiosoft/caddyplug"`
)

func installCaddy([]string) {
	fmt.Println("installing Caddy...")
	outputFile := "/usr/bin/caddy"

	// if not writable, fall back to local paths
	if !writable("/usr/bin") {
		outputFile = "/usr/local/bin/caddy"
		// check if GOBIN is in PATH and use it instead
		for _, binPath := range strings.Split(os.Getenv("PATH"),
			string([]byte{filepath.ListSeparator})) {
			if filepath.Clean(binPath) == filepath.Join(systemGoPath(), "bin") {
				outputFile = filepath.Join(systemGoPath(), "bin", "caddy")
				break
			}
		}
	}

	var e errs.Group
	pluginFile := filepath.Join(goPath(), "src", caddyPackage, pluginLoaderFile)
	e.Add(func() error {
		return ioutil.WriteFile(pluginFile, []byte(pluginLoaderSrc), 0644)
	})
	e.Add(func() error {
		return shellCmd{}.
			run("go", "build", "-o", outputFile, caddyPackage+"/caddy")
	})
	e.Add(func() error {
		fmt.Println(" ", successMark, "installed Caddy in", outputFile)
		return nil
	})
	e.Add(func() error {
		return os.Remove(pluginFile)
	})

	if err := e.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}
