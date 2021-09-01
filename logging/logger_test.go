package logging

import (
	"bytes"
	"io"
	"log"
	"testing"
	"time"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		lv level
		s  string
	}{
		{TRACE, "TRACE"},
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARNING, "WARNING"},
		{ERROR, "ERROR"},
		{CRITICAL, "CRITICAL"},
		{0, "level(0)"},
	}
	for _, tt := range tests {
		s := tt.lv.String()
		if s != tt.s {
			t.Errorf("level(%d).String() = %q; want %q", tt.lv, s, tt.s)
		}
	}
}

func TestGetLogger(t *testing.T) {
	var logger = GetLogger("tag")
	if logger.tag != "tag" {
		t.Errorf("tag should be tag but %v", logger.tag)
	}
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel(INFO)
	if logLv != INFO {
		t.Errorf("tag should be tag but %v", logLv.String())
	}
}

func prepareLogger(w io.Writer, v level) func() {
	o := lgr
	lgr = log.New(w, "", 0)
	SetLogLevel(v)
	lgr.SetFlags(0)
	return func() {
		lgr = o
	}
}

// These tests do not check anything yet.
// You can see result by go test -v
func TestLogf(t *testing.T) {
	var w bytes.Buffer
	t.Cleanup(prepareLogger(&w, TRACE))

	now := time.Unix(1630482120, 0).UTC()
	ts := "2021-09-01 07:42:00 +0000 UTC"
	var logger = GetLogger("tag")
	tests := []struct {
		logf   func(format string, args ...interface{})
		format string
		args   []interface{}
		want   string
	}{
		{
			logf:   logger.Criticalf,
			format: "This is critical log: %v",
			args:   []interface{}{now},
			want:   "CRITICAL <tag> This is critical log: " + ts + "\n",
		},
		{
			logf:   logger.Errorf,
			format: "This is error log: %v",
			args:   []interface{}{now},
			want:   "ERROR <tag> This is error log: " + ts + "\n",
		},
		{
			logf:   logger.Warningf,
			format: "This is warning log: %v",
			args:   []interface{}{now},
			want:   "WARNING <tag> This is warning log: " + ts + "\n",
		},
		{
			logf:   logger.Infof,
			format: "This is info log: %v",
			args:   []interface{}{now},
			want:   "INFO <tag> This is info log: " + ts + "\n",
		},
		{
			logf:   logger.Debugf,
			format: "This is debug log: %v",
			args:   []interface{}{now},
			want:   "DEBUG <tag> This is debug log: " + ts + "\n",
		},
		{
			logf:   logger.Tracef,
			format: "This is trace log: %v",
			args:   []interface{}{now},
			want:   "TRACE <tag> This is trace log: " + ts + "\n",
		},
	}
	for _, tt := range tests {
		w.Reset()
		tt.logf(tt.format, tt.args...)
		if s := w.String(); s != tt.want {
			t.Errorf("got %q; want %q", s, tt.want)
		}
	}
}

func TestInfoLogf(t *testing.T) {
	var w bytes.Buffer
	t.Cleanup(prepareLogger(&w, INFO))

	var logger = GetLogger("tag")
	tests := []struct {
		name   string
		logf   func(format string, args ...interface{})
		output bool
	}{
		{name: "critical", logf: logger.Criticalf, output: true},
		{name: "error", logf: logger.Errorf, output: true},
		{name: "warning", logf: logger.Warningf, output: true},
		{name: "info", logf: logger.Infof, output: true},
		{name: "debug", logf: logger.Debugf, output: false},
		{name: "trace", logf: logger.Tracef, output: false},
	}
	for _, tt := range tests {
		w.Reset()
		t.Run(tt.name, func(t *testing.T) {
			tt.logf("testtest")
			if tt.output {
				if s := w.String(); s == "" {
					t.Errorf("should output the log")
				}
			} else {
				if s := w.String(); s != "" {
					t.Errorf("should avoid any outputs")
				}
			}
		})
	}
}

func TestCriticalLogf(t *testing.T) {
	var w bytes.Buffer
	t.Cleanup(prepareLogger(&w, CRITICAL))

	var logger = GetLogger("tag")
	tests := []struct {
		name   string
		logf   func(format string, args ...interface{})
		output bool
	}{
		{name: "critical", logf: logger.Criticalf, output: true},
		{name: "error", logf: logger.Errorf, output: false},
		{name: "warning", logf: logger.Warningf, output: false},
		{name: "info", logf: logger.Infof, output: false},
		{name: "debug", logf: logger.Debugf, output: false},
		{name: "trace", logf: logger.Tracef, output: false},
	}
	for _, tt := range tests {
		w.Reset()
		t.Run(tt.name, func(t *testing.T) {
			tt.logf("testtest")
			if tt.output {
				if s := w.String(); s == "" {
					t.Errorf("should output the log")
				}
			} else {
				if s := w.String(); s != "" {
					t.Errorf("should avoid any outputs")
				}
			}
		})
	}
}
