package compiler

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx137/ruleset/magefiles/internal/util"
	P "github.com/metacubex/mihomo/constant/provider"
	"github.com/metacubex/mihomo/rules/provider"
)

type Mihomo struct {
}

func NewMihomo() *Mihomo {
	return &Mihomo{}
}

var behaviorMap = map[Behavior]P.RuleBehavior{
	BehaviorDomain:       P.Domain,
	BehaviorDomainSuffix: P.Domain,
	BehaviorIP:           P.IPCIDR,
	BehaviorCIDR:         P.IPCIDR,
}

var behaviorSuffix = map[P.RuleBehavior]string{
	P.Domain: "_domain.mrs",
	P.IPCIDR: "_ipcidr.mrs",
}

func (m *Mihomo) Compile(src string, rawOutputPath string, rules map[Behavior][]string) error {
	totalRules := 0
	mihomoRules := map[P.RuleBehavior][]string{
		P.Domain: make([]string, 0),
		P.IPCIDR: make([]string, 0),
	}
	unknownBehaviors := make([]Behavior, 0)

	for behavior, lines := range rules {
		if len(lines) == 0 {
			continue
		}

		totalRules += len(lines)

		mihomoBehavior, ok := behaviorMap[behavior]
		if !ok {
			unknownBehaviors = append(unknownBehaviors, behavior)
			continue
		}

		mihomoRules[mihomoBehavior] = append(mihomoRules[mihomoBehavior], lines...)
	}

	if len(unknownBehaviors) > 0 {
		for _, b := range unknownBehaviors {
			slog.Warn("Unknown behavior type, skipping rules", "input", filepath.ToSlash(src), "behavior", b)
		}
	}

	if totalRules == 0 {
		slog.Warn("No valid rules found in source", "input", filepath.ToSlash(src))
		return nil
	}

	for _, mihomoBehavior := range []P.RuleBehavior{P.Domain, P.IPCIDR} {
		lines := mihomoRules[mihomoBehavior]
		if len(lines) == 0 {
			continue
		}

		suffix := behaviorSuffix[mihomoBehavior]
		dst := util.GetOutputPath(rawOutputPath, suffix)

		if err := m.convertLines(lines, dst, mihomoBehavior); err != nil {
			return err
		}

		slog.Info("Generated ruleset", "input", filepath.ToSlash(src), "output", filepath.ToSlash(dst), "behavior", mihomoBehavior, "rules", len(lines))
	}

	return nil
}

func (m *Mihomo) convertLines(lines []string, outputPath string, behavior P.RuleBehavior) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", filepath.ToSlash(outputPath), err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file %s: %w", filepath.ToSlash(outputPath), err)
	}

	buf := []byte(strings.Join(lines, "\n"))

	if err := provider.ConvertToMrs(buf, behavior, P.TextRule, outFile); err != nil {
		_ = outFile.Close()
		return fmt.Errorf("convert to mrs %s: %w", filepath.ToSlash(outputPath), err)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("flush output file %s: %w", filepath.ToSlash(outputPath), err)
	}

	return nil
}
