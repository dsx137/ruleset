package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dsx137/ruleset/magefiles/internal/common"
)

func FreshDist() error {
	if err := os.RemoveAll(common.PathDist); err != nil {
		return fmt.Errorf("clean dist directory: %w", err)
	}

	if err := os.MkdirAll(common.PathDist, 0o755); err != nil {
		return fmt.Errorf("recreate dist directory: %w", err)
	}

	return nil
}

func GetSources() ([]string, error) {
	sources := make([]string, 0)

	err := filepath.Walk(common.PathRuleset, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			return nil
		}

		if !strings.EqualFold(filepath.Ext(info.Name()), ".txt") {
			return nil
		}

		sources = append(sources, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk ruleset directory: %w", err)
	}

	slices.Sort(sources)

	return sources, nil
}
