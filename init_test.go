package log15

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

func init() {
	Suite(&FormatSuite{})
	Suite(&CallerSuite{})
}
