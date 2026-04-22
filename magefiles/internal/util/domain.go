package util

import "strings"

func IsValidDomain(raw string) bool {
	if raw == "" {
		return false
	}

	raw = strings.ToLower(raw)

	if strings.ContainsAny(raw, "\t\n\r ,/") {
		return false
	}

	if strings.HasPrefix(raw, ".") || strings.HasSuffix(raw, ".") || strings.Contains(raw, "..") {
		return false
	}

	labels := strings.Split(raw, ".")
	if len(labels) < 2 {
		return false
	}

	totalLen := len(raw)
	if totalLen > 253 {
		return false
	}

	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}

		for i := 0; i < len(label); i++ {
			c := label[i]
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
				continue
			}
			return false
		}
	}

	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return false
	}
	allDigit := true
	for i := 0; i < len(tld); i++ {
		if tld[i] < '0' || tld[i] > '9' {
			allDigit = false
			break
		}
	}
	if allDigit {
		return false
	}

	return true
}
