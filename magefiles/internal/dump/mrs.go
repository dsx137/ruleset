package dump

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	P "github.com/metacubex/mihomo/constant/provider"
	"github.com/metacubex/mihomo/rules/provider"
)

func processMrsFile(path string) ([]string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read mrs file %s: %w", filepath.ToSlash(path), err)
	}

	behavior, err := detectMrsBehavior(buf)
	if err != nil {
		return nil, fmt.Errorf("detect mrs behavior %s: %w", filepath.ToSlash(path), err)
	}

	var out bytes.Buffer
	if err := provider.ConvertToMrs(buf, behavior, P.MrsRule, &out); err != nil {
		return nil, fmt.Errorf("dump mrs as text %s: %w", filepath.ToSlash(path), err)
	}

	rules := make([]string, 0)
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		rules = append(rules, line)
	}

	return rules, nil
}

func detectMrsBehavior(buf []byte) (P.RuleBehavior, error) {
	r, err := zstd.NewReader(bytes.NewReader(buf))
	if err != nil {
		return 0, err
	}
	defer r.Close()

	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return 0, err
	}

	if header != provider.MrsMagicBytes {
		return 0, fmt.Errorf("invalid mrs magic bytes")
	}

	var behaviorByte [1]byte
	if _, err := io.ReadFull(r, behaviorByte[:]); err != nil {
		return 0, err
	}

	switch behaviorByte[0] {
	case P.Domain.Byte():
		return P.Domain, nil
	case P.IPCIDR.Byte():
		return P.IPCIDR, nil
	case P.Classical.Byte():
		return 0, fmt.Errorf("classical mrs dump is not supported")
	default:
		return 0, fmt.Errorf("unknown behavior byte: %d", behaviorByte[0])
	}
}
