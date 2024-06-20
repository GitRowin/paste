# Paste

A lightweight service to anonymously share code snippets. Written in Go.

A public instance is available at https://paste.rowin.dev/.

This project is meant to be used with Cloudflare as it relies on `CF-Connecting-IP`, `CF-IPCountry`,
and `Cloudflare-CDN-Cache-Control`.

## Features

- Light and dark mode (uses your operating system's setting)
- Quickly save using CTRL + S
- Syntax highlighted and raw pastes

## Compiling

Compiling this project requires CGO to be enabled because of the `go-sqlite3` dependency.

`CGO_ENABLED=1 go build`

## This project uses

- [SQLite](https://github.com/mattn/go-sqlite3) for storing pastes
- [Chi](https://github.com/go-chi/chi) for routing
- [highlight.js](https://highlightjs.org/) for syntax highlighting
- [Toastify](https://github.com/apvarun/toastify-js) for toasts
- A DIY template engine that I made for fun
