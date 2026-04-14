//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	defaultBehavior = "domain"
	inputFormat     = "text"
)

// Build compiles all text rulesets into dist/ as .mrs files.
func Build() error {
	root, err := projectRoot()
	if err != nil {
		return err
	}

	rulesetsDir := filepath.Join(root, "rulesets")
	distDir := filepath.Join(root, "dist")

	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return fmt.Errorf("create dist directory: %w", err)
	}

	var sources []string
	err = filepath.Walk(rulesetsDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(info.Name()), ".txt") {
			sources = append(sources, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk rulesets directory: %w", err)
	}

	if len(sources) == 0 {
		fmt.Printf("No .txt rulesets found in %s\n", filepath.ToSlash(rulesetsDir))
		return nil
	}

	restore, err := chdir(root)
	if err != nil {
		return err
	}
	defer restore()

	goBin := goBinary()

	for _, src := range sources {
		rel, err := filepath.Rel(rulesetsDir, src)
		if err != nil {
			return fmt.Errorf("compute relative source path for %s: %w", src, err)
		}

		behavior := detectBehavior(rel)
		dst := filepath.Join(distDir, outputName(rel))

		displaySrc := filepath.ToSlash(filepath.Join("rulesets", rel))
		displayDst := filepath.ToSlash(filepath.Join("dist", outputName(rel)))
		fmt.Printf("Compiling %s -> %s (behavior=%s)\n", displaySrc, displayDst, behavior)

		if err := sh.Run(goBin,
			"run",
			"./cmd/converter/",
			"-behavior", behavior,
			"-format", inputFormat,
			"-input", src,
			"-output", dst,
		); err != nil {
			return fmt.Errorf("compile %s: %w", displaySrc, err)
		}
	}

	return nil
}

// Clean removes all generated rulesets from dist/.
func Clean() error {
	root, err := projectRoot()
	if err != nil {
		return err
	}

	distDir := filepath.Join(root, "dist")
	fmt.Printf("Removing %s\n", filepath.ToSlash(distDir))
	if err := os.RemoveAll(distDir); err != nil {
		return fmt.Errorf("remove dist directory: %w", err)
	}

	return nil
}

// All runs Clean and then Build.
func All() error {
	mg.SerialDeps(Clean, Build)
	return nil
}

func projectRoot() (string, error) {
	var candidates []string

	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Dir(filepath.Dir(file)))
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd, filepath.Dir(wd))
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}

		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		info, err := os.Stat(filepath.Join(candidate, "rulesets"))
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("unable to locate project root containing rulesets/")
}

func chdir(dir string) (func(), error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("determine current working directory: %w", err)
	}

	if err := os.Chdir(dir); err != nil {
		return nil, fmt.Errorf("change working directory to %s: %w", filepath.ToSlash(dir), err)
	}

	return func() {
		_ = os.Chdir(wd)
	}, nil
}

func goBinary() string {
	if envGo := os.Getenv("GO"); envGo != "" {
		return envGo
	}

	const homebrewGo = "/opt/homebrew/bin/go"
	if _, err := os.Stat(homebrewGo); err == nil {
		return homebrewGo
	}

	return "go"
}

func detectBehavior(rel string) string {
	for _, part := range strings.Split(filepath.ToSlash(filepath.Dir(rel)), "/") {
		switch strings.ToLower(part) {
		case "domain", "classical", "ipcidr":
			return strings.ToLower(part)
		}
	}

	return defaultBehavior
}

func outputName(rel string) string {
	name := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel)) + ".mrs"
	return strings.ReplaceAll(name, "/", "_")
}
