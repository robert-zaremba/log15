package log15setup

import (
	"errors"
	"regexp"
)

// ReName is a regular which tests valid app name for logger configuration
var ReName = regexp.MustCompile(`^[[:alnum:]\-_.]{2,200}`)

// CheckAppName validates the application name against `ReName`
// `what` is the an optional argument to specify name category / family.
func CheckAppName(name, what string) error {
	if ok := ReName.Match([]byte(name)); !ok {
		return errors.New("Wrong" + what + " name. Should match the following regexp: " +
			ReName.String())
	}
	return nil
}

// MustAppName validates if `name` matches `ReName` and panics (through logger.Fail)
// if it doesn't match.
func MustAppName(name, what string) {
	if ok := ReName.Match([]byte(name)); !ok {
		root.Fatal("Wrong name. Should match the following reg. expression",
			"regexp", ReName.String(), "what", what)
	}
}
