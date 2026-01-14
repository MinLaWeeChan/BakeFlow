package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type webviewTokenPayload struct {
	PSID string `json:"psid"`
	Exp  int64  `json:"exp"`
	Iat  int64  `json:"iat"`
}

func getWebviewTokenSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("WEBVIEW_TOKEN_SECRET"))
	if secret == "" {
		return nil
	}
	return []byte(secret)
}

// GenerateWebviewToken returns a signed token that encodes the user's PSID.
// The token is URL-safe and can be passed as query param `t`.
func GenerateWebviewToken(psid string, ttl time.Duration) (string, error) {
	secret := getWebviewTokenSecret()
	if len(secret) == 0 {
		return "", errors.New("WEBVIEW_TOKEN_SECRET is not set")
	}
	if strings.TrimSpace(psid) == "" {
		return "", errors.New("psid is empty")
	}
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour
	}

	now := time.Now().Unix()
	payload := webviewTokenPayload{PSID: psid, Iat: now, Exp: now + int64(ttl.Seconds())}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	sig := hmacSHA256(secret, []byte(payloadB64))
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)
	return payloadB64 + "." + sigB64, nil
}

// VerifyWebviewToken validates the token signature + expiry and returns the embedded PSID.
func VerifyWebviewToken(token string) (string, error) {
	secret := getWebviewTokenSecret()
	if len(secret) == 0 {
		return "", errors.New("WEBVIEW_TOKEN_SECRET is not set")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", errors.New("invalid token format")
	}
	payloadB64 := parts[0]
	sigB64 := parts[1]

	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return "", errors.New("invalid payload encoding")
	}
	providedSig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return "", errors.New("invalid signature encoding")
	}

	expectedSig := hmacSHA256(secret, []byte(payloadB64))
	if !hmac.Equal(providedSig, expectedSig) {
		return "", errors.New("invalid token signature")
	}

	var payload webviewTokenPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return "", errors.New("invalid payload")
	}
	if strings.TrimSpace(payload.PSID) == "" {
		return "", errors.New("missing psid")
	}
	if payload.Exp <= 0 {
		return "", errors.New("missing exp")
	}
	if time.Now().Unix() > payload.Exp {
		return "", errors.New("token expired")
	}

	return payload.PSID, nil
}

func hmacSHA256(secret, msg []byte) []byte {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write(msg)
	return h.Sum(nil)
}

func MustVerifyWebviewToken(token string) string {
	psid, err := VerifyWebviewToken(token)
	if err != nil {
		panic(fmt.Sprintf("invalid webview token: %v", err))
	}
	return psid
}
