// Package erkstrict controls if erk is running in strict mode.
package erkstrict

import (
	"fmt"
	"os"
	"strings"
)

var (
	isStrictModeSet bool
	isStrictMode    bool
)

// IsStrictMode reports if erk is running in strict mode.
//
// On the first run or after UnsetStrictMode is called, it reparses the strict mode.
// If the ERK_STRICT environment variable is set, it uses that value ("true" to enable, "false" to disable).
// Otherwise, it checks if it is running under tests by looking for a -test.* flag, which is automatically added by `go test`.
func IsStrictMode() bool {
	if isStrictModeSet {
		return isStrictMode
	}

	SetStrictMode(parseStrictMode())
	return isStrictMode
}

// UnsetStrictMode returns strict mode to the pristine state.
// It will check again for the ERK_STRICT environment variable and -test.* flag.
func UnsetStrictMode() {
	isStrictModeSet = false
}

// SetStrictMode to the provided state.
func SetStrictMode(enabled bool) {
	isStrictMode = enabled
	isStrictModeSet = true
}

func parseStrictMode() bool {
	strict, isSet := os.LookupEnv("ERK_STRICT")
	if isSet {
		return strict == "true"
	}

	// Check the args for -test.* flags
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			fmt.Println("WARNING: Detected using erk in tests, so strict mode is enabled. To disable strict mode for tests, set ERK_STRICT=false.")
			return true
		}
	}

	return false
}
