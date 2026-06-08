package auth

import "net/http"

// ApplyAuthorization mirrors server middleware: exact Authorization header match.
// The key value is never logged by this package.
func ApplyAuthorization(req *http.Request, key string) {
	if key == "" {
		return
	}
	req.Header.Set("Authorization", key)
}
