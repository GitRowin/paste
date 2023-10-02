# Paste

A lightweight service to anonymously share code snippets. Written in Go.

A public instance is available at https://paste.rowin.me/.

## Features

- Light and dark mode (uses your operating system's setting)
- Quickly save using CTRL + S
- Syntax highlighted and raw pastes

## This project uses

- [SQLite](https://github.com/mattn/go-sqlite3) for storing pastes
- [Chi](https://github.com/go-chi/chi) for routing
- [highlight.js](https://highlightjs.org/) for syntax highlighting
- [Toastify](https://github.com/apvarun/toastify-js) for toasts
- A DIY template engine that I made for fun
