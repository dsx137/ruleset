package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	P "github.com/metacubex/mihomo/constant/provider"
	"github.com/metacubex/mihomo/rules/provider"
)

func main() {
	log.SetFlags(0)

	var (
		behaviorFlag string
		formatFlag   string
		inputPath    string
		outputPath   string
	)

	flag.StringVar(&behaviorFlag, "behavior", "", "rule behavior: domain, ipcidr, or classical")
	flag.StringVar(&formatFlag, "format", "text", "input format")
	flag.StringVar(&inputPath, "input", "", "path to source file")
	flag.StringVar(&outputPath, "output", "", "path to output .mrs file")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s -behavior domain -format text -input rulesets/CN/Mirrors.txt -output dist/CN_Mirrors.mrs\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if strings.TrimSpace(behaviorFlag) == "" || strings.TrimSpace(inputPath) == "" || strings.TrimSpace(outputPath) == "" {
		flag.Usage()
		log.Fatal("-behavior, -input, and -output are required")
	}

	behavior, err := P.ParseBehavior(strings.ToLower(strings.TrimSpace(behaviorFlag)))
	if err != nil {
		log.Fatal(err)
	}

	format, err := P.ParseRuleFormat(strings.ToLower(strings.TrimSpace(formatFlag)))
	if err != nil {
		log.Fatal(err)
	}

	buf, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := provider.ConvertToMrs(buf, behavior, format, outFile); err != nil {
		_ = outFile.Close()
		log.Fatal(err)
	}

	if err := outFile.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Converted %s -> %s\n", inputPath, outputPath)
}
