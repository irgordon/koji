package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
)

const (
	requestIDHeader     = "X-Request-ID"
	requestIDMinLength  = 8
	requestIDMaxLength  = 128
	generatedIDBytes    = 16
	generatedIDFallback = "00000000000000000000000000000000"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)

func requestID(r *http.Request) string {
	if id, ok := requestIDFromRequest(r); ok {
		return id
	}
	return ""
}

func requestIDForRequest(r *http.Request) string {
	inbound := r.Header.Get(requestIDHeader)
	if validRequestID(inbound) {
		return inbound
	}
	return generateRequestID()
}

func validRequestID(id string) bool {
	if len(id) < requestIDMinLength || len(id) > requestIDMaxLength {
		return false
	}
	return requestIDPattern.MatchString(id)
}

func generateRequestID() string {
	bytes := make([]byte, generatedIDBytes)
	if _, err := rand.Read(bytes); err != nil {
		return generatedIDFallback
	}
	return hex.EncodeToString(bytes)
}
