package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var plugins = make(map[string]Plugin)

// Plugin is a caddy plugin.
type Plugin struct {
	Name    string
	Package string
	Type    string
}

// Build builds the plugin.
func (p Plugin) Build() error {
	if err := install(p.Package); err != nil {
		return err
	}
	file, err := generate(libraryPath(p.Type), p)
	if err != nil {
		return err
	}
	return build(file, filepath.Join(filepath.Dir(file), p.Name+".so"))
}

// Remove uninstalls the plugin.
func (p Plugin) Remove() error {
	if !p.Installed() {
		return nil
	}
	return os.Remove(p.PluginFile())
}

// Installed checks if the plugin is installed.
func (p Plugin) Installed() bool {
	stat, err := os.Stat(p.PluginFile())
	// TODO not all stat errors indicate file not present.
	return err == nil && !stat.IsDir()
}

// PluginFile returns the file path to the plugin .so file.
func (p Plugin) PluginFile() string {
	return filepath.Join(libraryPath(p.Type), p.Name+".so")
}

func initPlugins() error {
	if err := fetchDependencies(); err != nil {
		return err
	}
	for _, fetcher := range fetchers {
		p, err := fetcher.FetchPlugins()
		if err != nil {
			return err
		}
		for _, plugin := range p {
			plugins[plugin.Name] = plugin
		}
	}
	return nil
}

func generate(dir string, p Plugin) (string, error) {
	file := filepath.Join(dir, p.Name+".go")
	content := strings.Replace(pluginSrc, "{package}", p.Package, -1)
	return file, ioutil.WriteFile(file, []byte(content), 0666)
}

func build(src string, output string) error {
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", output, src)
	return cmd.Run()
}

func install(packageName string) error {
	return runCmd("go", "get", "-v", packageName)
}
