// Package env reads configuration from environment variables.
package env

import (
	"fmt"
	"os"
)

// MustGet returns the value of the environment variable key. It panics if
// the variable is unset or empty, so a misconfigured server never starts.
func MustGet(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %s is not set", key))
	}
	return v
}
