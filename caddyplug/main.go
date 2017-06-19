package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	pluginFile          = "github.com/mholt/caddy/caddyhttp/httpserver/plugin.go"
	dnsProvidersPackage = "github.com/caddyserver/dnsproviders"

	pluginSrc = `package main

import _ "{package}"
`
)

var ()

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

func runCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
