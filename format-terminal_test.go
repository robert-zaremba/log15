package log15

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	. "gopkg.in/check.v1"
)

type FormatSuite struct{}

func (suite FormatSuite) check(ctx []interface{}, expected string, c *C, comment CommentInterface) {
	b := &bytes.Buffer{}
	logfmt(b, ctx, 0)
	c.Check(b.String(), Equals, expected, comment)
}

func (suite FormatSuite) TestLogfmtCorrectOutput(c *C) {
	ctx := []interface{}{
		"key", 2,
		"", "value",
	}
	suite.check(ctx, "key=2 =\"value\"\n", c,
		Commentf("simple key value pairs should be formatted correctly"))

	// Test with nil error converted to interface
	var err error
	suite.check([]interface{}{err}, "\n", c,
		Commentf("nil entries in key position should be ignored"))

	// Test with error
	err = errors.New("this is an error\nwith lot of lines\n\t and tabs")
	errStr := fmt.Sprint(err) + "\n"
	suite.check([]interface{}{err}, "\n"+errHeader+errStr, c,
		Commentf("Single error should be formatted correctly"))

	// Test with error and values
	composed := []string{"one", "two"}
	ctx = []interface{}{
		"key", 2, err, Spew(composed, "context"),
		"time", time.Date(2015, 11, 20, 1, 34, 22, 0, time.UTC),
	}
	suite.check(ctx,
		"key=2   time=2015-11-20T01:34:22+0000\n"+
			"-------- context --------\n"+
			spew.Sdump(composed)+
			errHeader+errStr,
		c, Commentf("Spew and error around key-value paris should be formatted correctly"))
}

func (suite *FormatSuite) TestLogfmtMalformedOutput(c *C) {
	ctx := []interface{}{
		1, 2, 3,
	}
	suite.check(ctx, "MALFORMED_LOGFMT_KEY=2 MALFORMED_LOGFMT_KEY=MALFORMED_LOGFMT: no value for last key\n", c,
		Commentf("Malformed context should include malformed key in log"))
}
