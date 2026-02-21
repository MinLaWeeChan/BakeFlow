package controllers

import (
	"net/http"
	"strings"
)

func requirePSIDFromToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	tok := strings.TrimSpace(r.URL.Query().Get("t"))
	if tok == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return "", false
	}
	psid, err := VerifyWebviewToken(tok)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return "", false
	}
	return psid, true
}
