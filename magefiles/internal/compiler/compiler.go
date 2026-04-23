package compiler

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx137/ruleset/magefiles/internal/util"
)

type Behavior int

const (
	BehaviorDomain Behavior = iota
	BehaviorDomainSuffix
	BehaviorIP
	BehaviorCIDR
)

type Parser func(line string) (Behavior, string, bool)

type Compiler interface {
	Compile(rules map[Behavior][]string, rawOutputPath string) error
}

func ParseRules(inputPath string) (map[Behavior][]string, []string, error) {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read source file %s: %w", filepath.ToSlash(inputPath), err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(buf)))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	rules := map[Behavior][]string{
		BehaviorDomain:       make([]string, 0),
		BehaviorDomainSuffix: make([]string, 0),
		BehaviorIP:           make([]string, 0),
		BehaviorCIDR:         make([]string, 0),
	}
	warnings := make([]string, 0)
	parsers := []Parser{ParseIP, ParseCIDR, ParseDomainSuffix, ParseDomain}

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSpace(stripInlineComment(line))
		if line == "" {
			continue
		}

		parsed := false
		for _, parser := range parsers {
			behavior, parsedLine, matched := parser(line)
			if !matched {
				continue
			}

			rules[behavior] = append(rules[behavior], parsedLine)
			parsed = true
			break
		}

		if !parsed {
			warnings = append(warnings, fmt.Sprintf("line %d: unsupported rule %q", lineNo, line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return rules, warnings, nil
}

func stripInlineComment(line string) string {
	hashIndex := strings.Index(line, "#")
	if hashIndex >= 0 {
		return line[:hashIndex]
	}
	return line
}

func ParseCIDR(line string) (Behavior, string, bool) {
	if _, _, err := net.ParseCIDR(line); err == nil {
		return BehaviorCIDR, line, true
	}
	return 0, "", false
}

func ParseIP(line string) (Behavior, string, bool) {
	ip := net.ParseIP(line)
	if ip == nil {
		return 0, "", false
	}

	if ip.To4() != nil {
		return BehaviorIP, line + "/32", true
	}

	return BehaviorIP, line + "/128", true
}

func ParseDomainSuffix(line string) (Behavior, string, bool) {
	if !strings.HasPrefix(line, "+.") || len(line) <= 2 {
		return 0, "", false
	}

	suffix := strings.TrimPrefix(line, "+.")
	if util.IsValidDomain(suffix) {
		return BehaviorDomainSuffix, line, true
	}
	return 0, "", false
}

func ParseDomain(line string) (Behavior, string, bool) {
	if !util.IsValidDomain(line) {
		return 0, "", false
	}
	return BehaviorDomain, line, true
}
