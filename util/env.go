package util

import (
	"path"
	"path/filepath"
	"runtime"
)

func GetEnvFile() string {
	_, b, _, _ := runtime.Caller(0)
	return path.Join(filepath.Dir(b), "../.env")
}
