package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx137/ruleset/magefiles/internal/common"
	"github.com/dsx137/ruleset/magefiles/internal/compiler"
	"github.com/dsx137/ruleset/magefiles/internal/dump"
	"github.com/dsx137/ruleset/magefiles/internal/logging"
	"github.com/dsx137/ruleset/magefiles/internal/pipeline"
	"github.com/dsx137/ruleset/magefiles/internal/util"
)

func init() {
	slog.SetDefault(logging.New())
}

func Compile() error {
	compilers := []compiler.Compiler{
		compiler.NewMihomo(),
	}

	if err := pipeline.FreshDist(); err != nil {
		return err
	}

	sources, err := pipeline.GetSources()
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		slog.Warn("No .txt ruleset found", "rulesets_dir", filepath.ToSlash(common.PathRuleset))
		return nil
	}

	err = util.Parallel(4, func(feed func(func() error)) error {
		for _, src := range sources {
			feed(func() error {
				rel, relErr := filepath.Rel(common.PathRuleset, src)
				if relErr != nil {
					rel = src
				}

				slog.Info("Compiling ruleset", "input", filepath.ToSlash(filepath.Join(common.PathRuleset, rel)))

				behaviorRules, warnings, err := compiler.ParseRules(src)
				if err != nil {
					return fmt.Errorf("parse %s: %w", filepath.ToSlash(rel), err)
				}

				rawOutputPath, err := util.GetRawOutputPath(src)
				if err != nil {
					return fmt.Errorf("resolve output path for %s: %w", filepath.ToSlash(rel), err)
				}

				for _, warning := range warnings {
					slog.Warn("Skipped unsupported rule line", "input", filepath.ToSlash(filepath.Join(common.PathRuleset, rel)), "reason", warning)
				}

				for _, cpl := range compilers {
					if err := cpl.Compile(behaviorRules, rawOutputPath); err != nil {
						return fmt.Errorf("compile %s: %w", filepath.ToSlash(rel), err)
					}
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("Build completed", "sources", len(sources), "dist_dir", filepath.ToSlash(common.PathDist))
	return nil
}

func Dump(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("usage: mage print <path-to-mrs>")
	}

	items, err := dump.Dump(path)
	if err != nil {
		return err
	}

	for _, item := range items {
		if _, err := fmt.Fprintln(os.Stdout, item); err != nil {
			return err
		}
	}

	return nil
}

func Clean() error {
	if err := pipeline.CleanDist(); err != nil {
		return err
	}
	slog.Info("Cleaned dist directory", "dist_dir", filepath.ToSlash(common.PathDist))
	return nil
}
