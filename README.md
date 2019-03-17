# caddyplug

caddyplug is an experimental [Caddy](https://caddyserver.com) plugin manager using Go plugins.

[Demonstration Video](https://youtu.be/uKSgHhh6-rA)

## Requirements

- Go 1.8
- Linux/macOS
- Caddy with hook.pluginloader plugin. Installable with `caddyplug install-caddy`.

## Install

```
go get github.com/abiosoft/caddyplug/caddyplug
```

## Usage

```
  Usage:
    caddyplug <command> [plugins...]

  Commands:
    install       install plugins
    uninstall     uninstall plugins
    list          list plugins
    install-caddy install caddy
    package       get plugin package
```

Example

```sh
$ caddyplug install git linode
 ✓ git
 ✓ linode
```

## Goal

### Building

#### Current:

- Edit source and add import line for plugin
- Rebuild Caddy
- Or select plugins and download on caddyserver.com/download
- Repeat

#### Desired:

- Install plugins

### Docker

#### Current:

Option 1

- Search for Docker image with desired plugins
- Give up and clone abiosoft/caddy (or similar) image
- Modify plugins arg in Dockerfile
- Worry about keeping track of upgrades to parent git/docker repo.

Option 2

- Use [abiosoft/caddy:builder](https://github.com/abiosoft/caddy-docker/blob/master/BUILDER.md)
- Requires `docker build` and/or pushing custom image to own registry

#### Desired:

Add plugins as required

```Dockerfile
FROM abiosoft/caddy:plugin # Hopefully this changes to 'FROM caddy'
RUN caddyplug install git hugo digitalocean
```

## Caveats

- Only works on Linux/macOS.
- Due to limitations of Go plugins, Caddy and plugins must be built with same Go version. Installing Caddy with caddyplug is recommended to ensure this.
- Not compatible with caddyserver.com/download yet. Requires [`CGO_ENABLED=1`](https://github.com/golang/go/issues/19569).
- Large Docker images. Multi-stage builds may help.
- Fetches `master` of plugin repositories.
- `go build --buildmode=plugin` is slow. Stop building on-demand, maybe.
- This is experimental and reliant on the [stability of Go plugins](https://github.com/golang/go/issues?utf8=%E2%9C%93&q=is%3Aissue%20is%3Aopen%20plugins).

## Note

- This is not an official Caddy product.
