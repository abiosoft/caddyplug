caddyplug
=========

caddyplug is [Caddy](https://caddyserver.com) plugin manager using Go plugins.

## Requirements
* Go 1.8
* Linux
* Caddy with hook.pluginloader plugin.

## Install
```
go get github.com/abiosoft/caddyplug/caddyplug
```

## Usage
```
  Usage:
    caddyplug <command> plugins...

  Commands:
    install    install plugins
    uninstall  uninstall plugins
    list       list plugins
```

Example
```sh
$ caddyplug install git linode
 ✓ git
 ✓ linode
```

## Note
* This is experimental and reliant on the stability of Go plugins.
* This is not an official Caddy product.
