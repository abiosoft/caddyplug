package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/abiosoft/caddyplug/internal"
)

const (
	// TODO Go plugins require plugins and loaders to be built with same library versions.
	// this is not scalable, introduce flags maybe.
	caddyVersion        = "master"
	dnsProvidersVersion = "master"

	directivesFile      = "github.com/caddyserver/caddy/caddyhttp/httpserver/plugin.go"
	dnsProvidersPackage = "github.com/caddyserver/dnsproviders"
	caddyPackage        = "github.com/caddyserver/caddy"

	pluginSrc = `package main

import _ "{package}"
`
)

func usage(err ...interface{}) {
	exitWithError(append(err, `  Usage:
    caddyplug <command> [plugins...]

  Commands:
    install       install plugins
    uninstall     uninstall plugins
    list          list plugins
    install-caddy install caddy
    package       get plugin package
`)...)
}

func init() {
	switch runtime.GOOS {
	case "linux", "darwin":
	default:
		exitWithError("caddyplug is only supported on Linux and macOS")
	}

	// init once
	once.pluginPath = map[string]*sync.Once{}
	for _, pluginType := range []string{"http", "dns", "server", "caddyfile", "hook"} {
		once.pluginPath[pluginType] = &sync.Once{}
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
		usage(fmt.Sprintf("unkown command %s", os.Args[1]))
		return
	}

	if err := initPlugins(); err != nil {
		exitWithError(err)
	}
	cmd(pluginNames)
}

type shellCmd struct {
	Silent bool
	Stdin  bool
	Stdout io.Writer
	Dir    string
}

func (s shellCmd) run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	if !s.Silent {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	if s.Stdout != nil {
		cmd.Stdout = s.Stdout
	}
	if s.Stdin {
		cmd.Stdin = os.Stdin
	}
	if s.Dir != "" {
		cmd.Dir = s.Dir
	}
	cmd.Env = env()
	return cmd.Run()
}

var once struct {
	goPath     sync.Once
	pluginPath map[string]*sync.Once
}

func goPath() string {
	p := filepath.Join(internal.LibDir(), "gopath")
	once.goPath.Do(func() {
		os.MkdirAll(p, 0755)
	})
	return p
}

func systemGoPath() string {
	if os.Getenv("GOPATH") == "" {
		return filepath.Join(os.Getenv("HOME"), "go")
	}
	return os.Getenv("GOPATH")
}

func pluginPath(pluginType string) string {
	p := filepath.Join(internal.PluginsDir(), pluginType)
	once.pluginPath[pluginType].Do(func() {
		os.MkdirAll(p, 0755)
	})
	return p
}

// env replaces the GOPATH in env vars and returns
// resulting env vars.
func env() []string {
	env := []string{
		"GOPATH=" + goPath(),
	}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPATH=") {
			continue
		}
		env = append(env, e)
	}
	return env
}
