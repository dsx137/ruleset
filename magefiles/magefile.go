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
	slog.Info("Initialing...")
	slog.Info("Initialed...")
}

func Compile() {
	compilers := []compiler.Compiler{
		compiler.NewMihomo(),
	}

	if err := pipeline.FreshDist(); err != nil {
		slog.Error("Failed to prepare dist directory", "error", err)
		return
	}

	sources, err := pipeline.GetSources()
	if err != nil {
		slog.Error("Failed to get ruleset sources", "error", err)
		return
	}

	if len(sources) == 0 {
		slog.Error("No .txt ruleset found", "rulesets_dir", filepath.ToSlash(common.PathRuleset))
		return
	}

	err = util.Parallel(4, func(feed func(func() error)) error {
		for _, src := range sources {
			feed(func() error {
				slog.Info("Compiling ruleset", "input", src)

				behaviorRules, warnings, err := compiler.ParseRules(src)
				if err != nil {
					return fmt.Errorf("parse %s: %w", src, err)
				}

				for _, warning := range warnings {
					slog.Warn("Skipped unsupported rule line", "input", src, "reason", warning)
				}

				rawOutputPath, err := util.GetRawOutputPath(util.GetRelPath(common.PathRuleset, src))
				if err != nil {
					return fmt.Errorf("resolve output path for %s: %w", src, err)
				}

				for _, cpl := range compilers {
					if err := cpl.Compile(behaviorRules, rawOutputPath); err != nil {
						return fmt.Errorf("compile %s: %w", src, err)
					}
				}
				return nil
			})
		}
		for prefix, srcs := range util.GetSrcPrefixMap(sources) {
			feed(func() error {
				slog.Info("Compiling rulesets with common prefix", "prefix", prefix, "count", len(srcs))
				behaviorRules := make(map[compiler.Behavior][]string)
				for _, src := range srcs {
					rules, warnings, err := compiler.ParseRules(src)
					if err != nil {
						return fmt.Errorf("parse %s: %w", src, err)
					}

					for _, warning := range warnings {
						slog.Warn("Skipped unsupported rule line", "input", src, "reason", warning)
					}

					for behavior, lines := range rules {
						behaviorRules[behavior] = append(behaviorRules[behavior], lines...)
					}
				}

				rawOutputPath, err := util.GetRawOutputPath(prefix)
				if err != nil {
					return fmt.Errorf("resolve output path for prefix %s: %w", prefix, err)
				}

				for _, cpl := range compilers {
					if err := cpl.Compile(behaviorRules, rawOutputPath); err != nil {
						return fmt.Errorf("compile prefix %s: %w", prefix, err)
					}
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		slog.Error("Failed to compile rulesets", "error", err)
		return
	}

	slog.Info("Build completed", "sources", len(sources), "dist_dir", filepath.ToSlash(common.PathDist))
	return
}

func Dump(path string) {
	if strings.TrimSpace(path) == "" {
		slog.Error("Path is required for dump")
		return
	}

	items, err := dump.Dump(path)
	if err != nil {
		slog.Error("Failed to dump ruleset", "error", err)
		return
	}

	for _, item := range items {
		if _, err := fmt.Fprintln(os.Stdout, item); err != nil {
			slog.Error("Failed to write dumped item to stdout", "error", err)
			return
		}
	}

	return
}

func Clean() {
	if err := pipeline.CleanDist(); err != nil {
		slog.Error("Failed to clean dist directory", "error", err)
		return
	}
	slog.Info("Cleaned dist directory", "dist_dir", filepath.ToSlash(common.PathDist))
	return
}
