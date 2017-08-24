package log15

import (
	. "gopkg.in/check.v1"
)

type handlerRecorder struct {
	records []*Record
}

func (hr *handlerRecorder) Log(r *Record) error {
	hr.records = append(hr.records, r)
	return nil
}

func logSomething(logger Logger) {
	logger.Error("message 2", "param1", 2)
}

type CallerSuite struct{}

func (suite CallerSuite) TestCaller(c *C) {
	h := &handlerRecorder{}
	logger := New()
	logger.SetHandler(CallerFileHandler(h, true))

	logger.Error("message 1", "param1", 1)
	logSomething(logger)
	logger.Crit("message 3", "param1", 3)

	c.Assert(h.records, HasLen, 3)
	r := h.records[0]
	c.Check(r.Msg, Equals, "message 1")
	c.Check(r.Ctx, DeepEquals, []interface{}{"param1", 1, CallerCtx("log15/caller_test.go:27")})
	c.Check(r.Lvl, Equals, LvlError)
	r = h.records[1]
	c.Check(r.Msg, Equals, "message 2")
	c.Check(r.Ctx, DeepEquals, []interface{}{"param1", 2, CallerCtx("log15/caller_test.go:17")})
	r = h.records[2]
	c.Check(r.Msg, Equals, "message 3")
	c.Check(r.Ctx, DeepEquals, []interface{}{"param1", 3, CallerCtx("log15/caller_test.go:29")})
	c.Check(r.Lvl, Equals, LvlCrit)
}
