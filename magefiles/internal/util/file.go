package util

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/dsx137/ruleset/magefiles/internal/common"
)

func GetRawOutputPath(src string) (string, error) {
	name := strings.ReplaceAll(filepath.ToSlash(strings.TrimSuffix(src, filepath.Ext(src))), "/", "_")
	return filepath.Join(common.PathDist, name), nil
}

func GetRelPath(basePath string, targetPath string) string {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return targetPath
	}
	return filepath.ToSlash(rel)
}

func GetRelPaths(basePath string, targetPaths []string) []string {
	ret := make([]string, len(targetPaths))
	for i, p := range targetPaths {
		ret[i] = GetRelPath(basePath, p)
	}

	return ret
}

func GetSrcPrefixMap(sources []string) map[string][]string {
	srcMap := make(map[string][]string, len(sources))
	uniq := make(map[string]struct{}, len(sources))

	for _, src := range sources {
		rel := GetRelPath(common.PathRuleset, src)
		s := strings.TrimSpace(rel)
		s = strings.ReplaceAll(s, "\\", "/")
		s = strings.TrimPrefix(s, "./")
		s = path.Clean(s)
		if s == "" {
			continue
		}
		if _, ok := uniq[s]; ok {
			continue
		}
		uniq[s] = struct{}{}
		dir := path.Dir(s)
		if dir == "." || dir == "" {
			dir = "."
		}

		parts := strings.Split(dir, "/")
		for i := range parts {
			prefix := strings.Join(parts[:i+1], "/")
			srcMap[prefix] = append(srcMap[prefix], src)
		}
	}

	return srcMap
}
