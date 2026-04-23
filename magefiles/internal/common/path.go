package common

import "path/filepath"

var (
	PathRoot    = "."
	PathRuleset = filepath.Join(PathRoot, "ruleset")
	PathDist    = filepath.Join(PathRoot, "dist")
)
