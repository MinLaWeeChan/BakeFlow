package controllers

import "strings"

func SanitizeMyanmarPhoneInput(raw string) string {
	v := strings.TrimSpace(raw)
	hasPlus := strings.HasPrefix(v, "+")
	var b strings.Builder
	for _, r := range v {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if digits == "" {
		if hasPlus {
			return "+"
		}
		return ""
	}
	if hasPlus {
		return "+" + digits
	}
	return digits
}

func NormalizeMyanmarPhoneE164(raw string) (string, bool) {
	s := SanitizeMyanmarPhoneInput(raw)
	if s == "" {
		return "", false
	}
	if strings.HasPrefix(s, "+") {
		if strings.HasPrefix(s, "+959") && len(s) == len("+959")+9 {
			for _, r := range s[len("+959"):] {
				if r < '0' || r > '9' {
					return "", false
				}
			}
			return s, true
		}
		return "", false
	}
	if strings.HasPrefix(s, "09") && len(s) == len("09")+9 {
		for _, r := range s[len("09"):] {
			if r < '0' || r > '9' {
				return "", false
			}
		}
		return "+959" + s[len("09"):], true
	}
	return "", false
}
