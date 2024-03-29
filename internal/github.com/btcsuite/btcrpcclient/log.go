// Copyright (c) 2014-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcrpcclient

import (
	"errors"
	"io"

	"github.com/whoisjeremylam/enu/internal/github.com/btcsuite/btclog"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log btclog.Logger

// The default amount of logging is none.
func init() {
	DisableLog()
}

// DisableLog disables all library log output.  Logging output is disabled
// by default until either UseLogger or SetLogWriter are called.
func DisableLog() {
	log = btclog.Disabled
}

// UseLogger uses a specified Logger to output package logging info.
// This should be used in preference to SetLogWriter if the caller is also
// using btclog.
func UseLogger(logger btclog.Logger) {
	log = logger
}

// SetLogWriter uses a specified io.Writer to output package logging info.
// This allows a caller to direct package logging output without needing a
// dependency on seelog.  If the caller is also using btclog, UseLogger should
// be used instead.
func SetLogWriter(w io.Writer, level string) error {
	if w == nil {
		return errors.New("nil writer")
	}

	lvl, ok := btclog.LogLevelFromString(level)
	if !ok {
		return errors.New("invalid log level")
	}

	l, err := btclog.NewLoggerFromWriter(w, lvl)
	if err != nil {
		return err
	}

	UseLogger(l)
	return nil
}

// LogClosure is a closure that can be printed with %v to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type logClosure func() string

// String invokes the log closure and returns the results string.
func (c logClosure) String() string {
	return c()
}

// newLogClosure returns a new closure over the passed function which allows
// it to be used as a parameter in a logging function that is only invoked when
// the logging level is such that the message will actually be logged.
func newLogClosure(c func() string) logClosure {
	return logClosure(c)
}
