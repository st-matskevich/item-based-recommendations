//go:build tools
// +build tools

package tools

// See this https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md

import (
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
)
