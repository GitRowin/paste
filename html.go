package main

import (
	"strings"
	"text/template"
)

type Attr struct {
	Key   string
	Value string
}

// Doc is short for document
func Doc(children ...string) string {
	builder := strings.Builder{}

	builder.WriteString("<!doctype html>")

	for _, child := range children {
		builder.WriteString(child)
	}

	return builder.String()
}

// El is short for element
func El(name string, children ...string) string {
	return createElement(name, nil, false, children...)
}

// Ela is short for element with attributes
func Ela(name string, attributes []Attr, children ...string) string {
	return createElement(name, attributes, false, children...)
}

// Sela is short for self-closing element with attributes
func Sela(name string, attributes []Attr) string {
	return createElement(name, attributes, true)
}

func createElement(name string, attributes []Attr, selfClosing bool, children ...string) string {
	builder := strings.Builder{}

	builder.WriteString("<")
	builder.WriteString(name)

	for _, attr := range attributes {
		builder.WriteString(" ")
		builder.WriteString(attr.Key)
		builder.WriteString(`="`)
		builder.WriteString(attr.Value)
		builder.WriteString(`"`)
	}

	builder.WriteString(">")

	if !selfClosing {
		for _, child := range children {
			builder.WriteString(child)
		}

		builder.WriteString("</")
		builder.WriteString(name)
		builder.WriteString(">")
	}

	return builder.String()
}

func If(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}

func IfComputed(cond bool, f func() string) string {
	if cond {
		return f()
	}
	return ""
}

func IfElse(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func IfElseComputed(cond bool, a, b func() string) string {
	if cond {
		return a()
	}
	return b()
}

func ForEach[T any](arr []T, f func(index int, value T) string) string {
	builder := strings.Builder{}

	for i, v := range arr {
		builder.WriteString(f(i, v))
	}

	return builder.String()
}

func Esc(s string) string {
	return template.HTMLEscapeString(s)
}
