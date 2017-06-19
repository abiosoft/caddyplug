package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	// TODO Go plugins require plugins and loaders to be built with same library versions.
	// this is not scalable, introduce flags maybe.
	caddyVersion        = "master"
	dnsProvidersVersion = "master"

	pluginFile          = "github.com/mholt/caddy/caddyhttp/httpserver/plugin.go"
	dnsProvidersPackage = "github.com/caddyserver/dnsproviders"

	pluginSrc = `package main

import _ "{package}"
`
)

func usage() {
	fmt.Println(`  Usage:
    caddyplug <command> plugins...

  Commands:
    install    install plugins
    uninstall  uninstall plugins
    list       list plugins
`)
}

func init() {
	if runtime.GOOS != "linux" {
		fmt.Println("caddyplug is only supported on Linux")
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	var pluginNames []string
	if len(os.Args) > 2 {
		pluginNames = os.Args[2:]
	}
	cmd, ok := commands[os.Args[1]]
	if !ok {
		fmt.Println("unkown command", os.Args[1])
		usage()
		return
	}

	if err := initPlugins(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cmd(pluginNames)
}

type shellCmd struct {
	Silent bool
	Dir    string
}

func (s shellCmd) run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	if !s.Silent {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if s.Dir != "" {
		cmd.Dir = s.Dir
	}
	return cmd.Run()
}

func goPath() string {
	if os.Getenv("GOPATH") == "" {
		return filepath.Join(os.Getenv("HOME"), "go")
	}
	return os.Getenv("GOPATH")
}

func libraryPath(pluginType string) string {
	p := filepath.Join(os.Getenv("HOME"), "lib", "caddy", pluginType)
	os.MkdirAll(p, 0755)
	return p
}
