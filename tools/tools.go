//go:build tools

package tools

// нужен для генерирующих библиотек
import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
