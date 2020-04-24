// +build tools

package main

import (
	_ "github.com/mattn/goveralls"
	_ "github.com/pierrre/gotestcover"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/cover"
)
