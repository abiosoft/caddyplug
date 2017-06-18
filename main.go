package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/mholt/caddy"
)

const (
	pluginFile = "github.com/mholt/caddy/caddyhttp/httpserver/plugin.go"
	dnsDir     = "github.com/caddyserver/dnsproviders"

	pluginSrc = `package main

import _ "{package}"
`
)

var (
	fetchers = map[string]pluginFetcher{
		"http": fetcherFunc(fetchHTTPPlugins),
		"dns":  fetcherFunc(fetchDNSPlugins),
	}

	plugins = make(map[string]Plugin)

	// flags
	list      bool
	uninstall bool

	successMark = color.GreenString("✓")
	errorMark   = color.RedString("✗")
)

func init() {
	if runtime.GOOS != "linux" {
		fmt.Println("caddyplug is only supported on Linux")
		os.Exit(1)
	}

	flag.Usage = usage
	flag.BoolVar(&list, "list", false, "list plugins")
	flag.BoolVar(&uninstall, "uninstall", false, "uninstall plugins")
	flag.Parse()
}

func usage() {
	fmt.Println(`  Usage:
    caddyplug [flags] plugins...

  Flags:
    -list      list plugins
    -uninstall uninstall plugins
	`)
}

func main() {
	err := initPlugins()
	if err != nil {
		fmt.Println(err)
		return
	}
	pluginNames := flag.Args()
	if list {
		listPlugins()
		return
	}
	if len(pluginNames) == 0 {
		usage()
		return
	}
	var outputs []string
	for _, pluginName := range pluginNames {
		plugin, ok := plugins[pluginName]
		if !ok {
			fmt.Println("plugin not found", pluginName)
			return
		}
		var err error
		if uninstall {
			if err = plugin.Remove(); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = plugin.Build(); err != nil {
				fmt.Println(err)
			}
		}
		if err == nil {
			outputs = append(outputs, " "+successMark+" "+plugin.Name)
		} else {
			outputs = append(outputs, " "+errorMark+" "+plugin.Name)
		}
	}
	if len(outputs) > 0 {
		if uninstall {
			fmt.Println("Uninstalled:")
		}
		for _, p := range outputs {
			fmt.Println(p)
		}
	}
}

type pluginFetcher interface {
	FetchPlugins() ([]Plugin, error)
}

func listPlugins() {
	for pluginType, fetcher := range fetchers {
		plugins, err := fetcher.FetchPlugins()
		if err != nil {
			fmt.Println(err)
			return
		}
		sort.Slice(plugins, func(i, j int) bool {
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

func initPlugins() error {
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

func build(src string, output string) error {
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", output, src)
	return cmd.Run()
}

func install(packageName string) error {
	cmd := exec.Command("go", "get", "-v", packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generate(dir string, p Plugin) (string, error) {
	file := filepath.Join(dir, p.Name+".go")
	content := strings.Replace(pluginSrc, "{package}", p.Package, -1)
	return file, ioutil.WriteFile(file, []byte(content), 0666)
}

func goPath() string {
	if os.Getenv("GOPATH") == "" {
		return filepath.Join(os.Getenv("HOME"), "go")
	}
	return os.Getenv("GOPATH")
}

func pluginPath(pluginType string) string {
	p := filepath.Join(caddy.AssetsPath(), "plugins", pluginType)
	os.MkdirAll(p, 0755)
	return p
}

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
	file, err := generate(pluginPath(p.Type), p)
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
	return filepath.Join(pluginPath(p.Type), p.Name+".so")
}

type fetcherFunc func() ([]Plugin, error)

func (f fetcherFunc) FetchPlugins() ([]Plugin, error) { return f() }

func fetchHTTPPlugins() ([]Plugin, error) {
	var plugins []Plugin
	fset := token.NewFileSet()
	file := filepath.Join(goPath(), "src", pluginFile)
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return plugins, err
	}
	node, ok := f.Scope.Lookup("directives").Decl.(*ast.ValueSpec)
	if !ok {
		return plugins, fmt.Errorf("parsing error")
	}

	cmap := ast.NewCommentMap(fset, f, f.Comments)
	c := node.Values[0].(*ast.CompositeLit)
	for _, m := range c.Elts {
		if _, ok := cmap[m]; ok {
			pkg := strings.TrimSpace(cmap[m][0].Text())
			directive, err := strconv.Unquote(m.(*ast.BasicLit).Value)
			if err != nil {
				return plugins, err
			}
			// asserting that the comment word count is 1 may not be the best way
			// to confirm it is a repo path.
			if len(strings.Fields(pkg)) == 1 {
				plugin := Plugin{
					Name:    directive,
					Package: pkg,
					Type:    "http",
				}
				plugins = append(plugins, plugin)
			}
		}
	}
	return plugins, nil
}

func fetchDNSPlugins() ([]Plugin, error) {
	var plugins []Plugin
	srcDir := filepath.Join(goPath(), "src", dnsDir)
	d, err := os.Open(srcDir)
	if err != nil {
		return plugins, err
	}
	stats, err := d.Readdir(-1)
	if err != nil {
		return plugins, err
	}
	for _, stat := range stats {
		provider := stat.Name()
		// skip hidden files
		if strings.HasPrefix(provider, ".") || !stat.IsDir() {
			continue
		}
		plugin := Plugin{
			Name:    provider,
			Package: path.Join(dnsDir, provider),
			Type:    "dns",
		}
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}
