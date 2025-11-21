package jwtcustomverifiers

import "net/http"

func VerifyApiKeyHeader(r *http.Request) string {
	return r.Header.Get("Api-key")
}
