package util

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dsx137/ruleset/magefiles/internal/common"
)

func GetRawOutputPath(sourcePath string) (string, error) {
	rel, err := filepath.Rel(common.PathRuleset, sourcePath)
	if err != nil {
		return "", fmt.Errorf("compute relative path for %s: %w", sourcePath, err)
	}
	name := strings.ReplaceAll(filepath.ToSlash(strings.TrimSuffix(rel, filepath.Ext(rel))), "/", "_")
	return filepath.Join(common.PathDist, name), nil
}

func GetOutputPath(rawOutputPath string, suffix string) string {
	return rawOutputPath + suffix
}
