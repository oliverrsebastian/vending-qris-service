package controller

import (
	"errors"
	"strings"
)

var (
	errCallbackNotConfigured  = errors.New("payment callback use case not configured")
	errMissingCallbackGateway = errors.New("callback gateway path segment is required")
)

func isUnknownCallbackGateway(err error) bool {
	return strings.Contains(err.Error(), "unknown callback handler")
}

func isCallbackClientError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "invalid json") ||
		strings.Contains(msg, "is required")
}

func isTransactionNotFound(err error) bool {
	return strings.Contains(err.Error(), "transaction") && strings.Contains(err.Error(), "not found")
}
