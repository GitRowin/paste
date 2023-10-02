# Paste

A simple service to anonymously share code snippets. Written in Go.

This project uses [SQLite](https://github.com/mattn/go-sqlite3) for storing pastes, [Chi](https://github.com/go-chi/chi) for routing, [highlight.js](https://highlightjs.org/) for (client-side) syntax highlighting, and a DIY template engine that I made for fun.

A live instance is available at https://paste.rowin.me/.
