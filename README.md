caddyplug
=========

caddyplug is [Caddy](https://caddyserver.com) plugin manager using Go plugins.

## Requirements
* Go 1.8
* Linux
* Caddy with [Plugin Loader](http://) [Still WIP].

## Install
```
go get github.com/abiosoft/caddyplug
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
This is not an official Caddy product.
