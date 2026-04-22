package dump

import (
	"fmt"
	"path/filepath"
	"strings"
)

func Dump(path string) ([]string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mrs":
		return processMrsFile(path)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", filepath.ToSlash(path))
	}
}
