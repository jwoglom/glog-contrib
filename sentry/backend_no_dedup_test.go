package sentry_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/yext/glog"
	"github.com/yext/yerrors"
)

func TestGlogSimpleEventNoDedup(t *testing.T) {
	methodName := "TestGlogSimpleEventNoDedup" // this should stay in sync with the name of the method

	ready := make(chan interface{})
	done := make(chan *sentrygo.Event)
	go setup(ready, done, 1, false)

	<-ready
	errorLine := 1 + currentLine() // this should point to the next line
	glog.Error("test message")
	e := <-done

	assert.NotNil(t, e)
	assert.Equal(t, sentrygo.LevelError, e.Level, "level is error")
	assert.Equal(t, "test message", e.Message, "message matches exactly")
	assert.Len(t, e.Exception, 1, "one exception")

	ex := e.Exception[0] // the exception is from the glog invocation
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) matches the error string exactly")
	assert.True(t, strings.HasPrefix(ex.Value, fmt.Sprintf("%s:%d", methodName, errorLine)),
		"value (issue subtitle) starts with the method name and error line of the glog invocation: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 1, "one stacktrace frame")

	fr := ex.Stacktrace.Frames[0]
	assert.Equal(t, methodName, fr.Function, "function name matches")
	assert.Equal(t, errorLine, fr.Lineno, "line number matches of the glog invocation")
	assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	assert.True(t, fr.InApp, "inapp flag true")
}

func TestGlogErrorfEventNoDedup(t *testing.T) {
	methodName := "TestGlogErrorfEventNoDedup" // this should stay in sync with the name of the method

	ready := make(chan interface{})
	done := make(chan *sentrygo.Event)
	go setup(ready, done, 1, false)

	<-ready
	errorLine := 1 + currentLine() // this should point to the next line
	glog.Errorf("test %s: %s", "message", "more details")
	e := <-done

	assert.NotNil(t, e)
	assert.Equal(t, sentrygo.LevelError, e.Level, "level is error")
	assert.Equal(t, "test message: more details", e.Message,
		"message matches exactly with full error text")
	assert.Len(t, e.Exception, 1, "one exception")

	ex := e.Exception[0] // the exception is from the glog invocation
	assert.Equal(t, "test", ex.Type,
		"type (primary issue title) matches first component of the error string with removed formatters")
	assert.True(t, strings.HasPrefix(ex.Value, "more details"),
		"value (issue subtitle) starts with the remainder of the error string: "+ex.Value)
	assert.True(t, strings.HasSuffix(ex.Value, fmt.Sprintf("(%s:%d)", methodName, errorLine)),
		"value (issue subtitle) ends with the method name and error line of the glog invocation: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 1, "one stacktrace frame")

	fr := ex.Stacktrace.Frames[0]
	assert.Equal(t, methodName, fr.Function, "function name matches")
	assert.Equal(t, errorLine, fr.Lineno, "line number matches of the glog invocation")
	assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	assert.True(t, fr.InApp, "inapp flag true")
}

func TestGlogRawErrorEventNoDedup(t *testing.T) {
	methodName := "TestGlogRawErrorEventNoDedup" // this should stay in sync with the name of the method

	ready := make(chan interface{})
	done := make(chan *sentrygo.Event)
	go setup(ready, done, 1, false)

	<-ready
	// We cannot track where the raw error occurred because it uses a raw error type
	err := errors.New("test message")
	errorLine := 1 + currentLine() // this should point to the next line
	glog.Error(err)
	e := <-done

	assert.NotNil(t, e)
	assert.Equal(t, sentrygo.LevelError, e.Level, "level is error")
	assert.Equal(t, "test message", e.Message, "message matches exactly")
	assert.Len(t, e.Exception, 2, "two exceptions (first is from glog, second is from the raw err)")

	ex := e.Exception[0] // the first exception is from the glog invocation
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) matches the error string exactly")
	assert.True(t, strings.HasPrefix(ex.Value, fmt.Sprintf("%s:%d", methodName, errorLine)),
		"value (issue subtitle) starts with the method name and error line of the glog invocation: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 1, "one stacktrace frame")

	fr := ex.Stacktrace.Frames[0]
	assert.Equal(t, methodName, fr.Function, "function name matches")
	assert.Equal(t, errorLine, fr.Lineno, "line number matches of the glog invocation")
	assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	assert.True(t, fr.InApp, "inapp flag true")

	ex = e.Exception[1] // the second exception is from the raw error
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) matches the error string exactly")
	assert.Empty(t, ex.Value, "value of raw error is empty")
	// the raw error has no stacktrace or stack frames
	assert.Nil(t, ex.Stacktrace, "stacktrace of raw error is nil")
}

func TestGlogRawErrorEventWithColonNoDedup(t *testing.T) {
	methodName := "TestGlogRawErrorEventWithColonNoDedup" // this should stay in sync with the name of the method

	ready := make(chan interface{})
	done := make(chan *sentrygo.Event)
	go setup(ready, done, 1, false)

	<-ready
	// We cannot track where the raw error occurred because it uses a raw error type
	err := errors.New("test message: more details")
	errorLine := 1 + currentLine() // this should point to the next line
	glog.Error(err)
	e := <-done

	assert.NotNil(t, e)
	assert.Equal(t, sentrygo.LevelError, e.Level, "level is error")
	assert.Equal(t, "test message: more details", e.Message,
		"message matches exactly containing detail after colon")
	assert.Len(t, e.Exception, 2, "two exceptions (first is from glog, second is from the raw err)")

	ex := e.Exception[0] // the first exception is from the glog invocation
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) matches the error string exactly")
	assert.True(t, strings.HasSuffix(ex.Value, fmt.Sprintf("(%s:%d)", methodName, errorLine)),
		"value (issue subtitle) ends with the method name and error line of the glog invocation: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 1, "one stacktrace frame")

	fr := ex.Stacktrace.Frames[0]
	assert.Equal(t, methodName, fr.Function, "function name matches")
	assert.Equal(t, errorLine, fr.Lineno, "line number matches of the glog invocation")
	assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	assert.True(t, fr.InApp, "inapp flag true")

	ex = e.Exception[1] // the second exception is from the raw error
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) matches the error string before the colon exactly")
	assert.Equal(t, "more details", ex.Value,
		"value of raw error matches the error string after the colon exactly")
	// the raw error has no stacktrace or stack frames
	assert.Nil(t, ex.Stacktrace, "stacktrace of raw error is nil")
}

func TestGlogYerrorsEventNoDedup(t *testing.T) {
	methodName := "TestGlogYerrorsEventNoDedup" // this should stay in sync with the name of the method

	ready := make(chan interface{})
	done := make(chan *sentrygo.Event)
	go setup(ready, done, 1, false)

	<-ready
	errorLine := 1 + currentLine() // this should point to the next line
	err := yerrors.New("test message")
	glogErrorLine := 1 + currentLine() // this should point to the next line
	glog.Error(err)
	e := <-done

	assert.NotNil(t, e)
	assert.Equal(t, sentrygo.LevelError, e.Level, "level is error")
	assert.Equal(t, "test message", strings.SplitN(e.Message, "\n", 2)[0],
		"first line of the message equals the error string exactly")
	assert.Len(t, e.Exception, 2, "two exceptions (first is from glog, second is from the raw err)")

	ex := e.Exception[0] // first exception is from the glog invocation
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) matches the error string exactly")
	assert.Equal(t, fmt.Sprintf("%s:%d", methodName, glogErrorLine), ex.Value,
		"value (issue subtitle) equals the method name and error line of the glog invocation exactly: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 1, "one stacktrace frame")

	fr := ex.Stacktrace.Frames[0]
	assert.True(t, strings.HasSuffix(fr.Function, methodName), "function name has suffix: "+fr.Function)
	assert.Equal(t, glogErrorLine, fr.Lineno, "line number matches of the glog invocation")
	assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)

	ex = e.Exception[1] // second exception is passed from the error argument
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) equals the error string exactly: "+ex.Type)
	assert.Equal(t, fmt.Sprintf("%s.%s:%d", pkgName, methodName, errorLine), ex.Value,
		"value (issue subtitle) equals the method name and error line of the yerrors invocation exactly: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 2, "two stacktrace frames")

	for _, fr := range ex.Stacktrace.Frames {
		assert.True(t, strings.HasSuffix(fr.Function, methodName), "function name has suffix: "+fr.Function)
		assert.Equal(t, errorLine, fr.Lineno, "line number matches of the yerrors invocation")
		assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	}
}

func TestGlogYerrorsEventWithColonNoDedup(t *testing.T) {
	methodName := "TestGlogYerrorsEventWithColonNoDedup" // this should stay in sync with the name of the method

	ready := make(chan interface{})
	done := make(chan *sentrygo.Event)
	go setup(ready, done, 1, false)

	<-ready
	errorLine := 1 + currentLine() // this should point to the next line
	err := yerrors.New("test message: more details")
	glogErrorLine := 1 + currentLine() // this should point to the next line
	glog.Error(err)
	e := <-done

	assert.NotNil(t, e)
	assert.Equal(t, sentrygo.LevelError, e.Level, "level is error")
	assert.Equal(t, "test message: more details", strings.SplitN(e.Message, "\n", 2)[0],
		"first line of the message equals the error string exactly")
	assert.Len(t, e.Exception, 2, "two exceptions (first is from glog, second is from the raw err)")

	ex := e.Exception[0] // first exception is from the glog invocation
	assert.Equal(t, "test message", ex.Type,
		"type (primary issue title) matches the error string exactly")
	assert.True(t, strings.HasPrefix(ex.Value, "more details"),
		"value (issue subtitle) starts with the second half of the error string: "+ex.Value)
	assert.True(t, strings.HasSuffix(ex.Value, fmt.Sprintf("(%s:%d)", methodName, glogErrorLine)),
		"value (issue subtitle) ends with the method name and error line of the glog invocation: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 1, "one stacktrace frame")

	fr := ex.Stacktrace.Frames[0]
	assert.True(t, strings.HasSuffix(fr.Function, methodName), "function name has suffix: "+fr.Function)
	assert.Equal(t, glogErrorLine, fr.Lineno, "line number matches of the glog invocation")
	assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)

	ex = e.Exception[1] // second exception is passed from the error argument
	assert.True(t, strings.HasPrefix(ex.Type, "test message"),
		"type (primary issue title) starts with the error string: "+ex.Type)
	assert.True(t, strings.HasPrefix(ex.Value, "more details"),
		"type (primary issue title) starts with the second half of the error string: "+ex.Value)
	assert.True(t, strings.HasSuffix(ex.Value, fmt.Sprintf("(%s.%s:%d)", pkgName, methodName, errorLine)),
		"value (issue subtitle) ends with the method name and error line of the yerrors invocation: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 2, "two stacktrace frames")

	for _, fr := range ex.Stacktrace.Frames {
		assert.True(t, strings.HasSuffix(fr.Function, methodName), "function name has suffix: "+fr.Function)
		assert.Equal(t, errorLine, fr.Lineno, "line number matches of the yerrors invocation")
		assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	}
}

func TestGlogYerrorsWrappedEventNoDedup(t *testing.T) {
	methodName := "TestGlogYerrorsWrappedEventNoDedup" // this should stay in sync with the name of the method

	ready := make(chan interface{})
	done := make(chan *sentrygo.Event)
	go setup(ready, done, 1, false)

	<-ready
	errorLine := 1 + currentLine() // this should point to the next line
	err := yerrors.New("test message")
	errorWrappedLine := 1 + currentLine() // this should point to the next line
	wrap := yerrors.Wrap(err)
	glogErrorLine := 1 + currentLine() // this should point to the next line
	glog.Error(wrap)
	e := <-done

	assert.NotNil(t, e)
	assert.Equal(t, sentrygo.LevelError, e.Level, "level is error")
	assert.True(t, strings.HasPrefix(e.Message, "test message"),
		"message starts with the error string")
	assert.Len(t, e.Exception, 3, "three exceptions (first is from glog, second two are from err)")

	ex := e.Exception[0] // first exception is from the glog invocation
	assert.True(t, strings.HasPrefix(ex.Type, "test message"),
		"type (primary issue title) starts with the error string: "+ex.Type)
	assert.True(t, strings.HasPrefix(ex.Value, fmt.Sprintf("%s:%d", methodName, glogErrorLine)),
		"value (issue subtitle) starts with the method name and error line of the glog invocation: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 1, "one stacktrace frame")

	fr := ex.Stacktrace.Frames[0]
	assert.True(t, strings.HasSuffix(fr.Function, methodName), "function name has suffix: "+fr.Function)
	assert.Equal(t, glogErrorLine, fr.Lineno, "line number matches")
	assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)

	ex = e.Exception[1] // second exception contains the inner frame from the invoked error
	assert.True(t, strings.HasPrefix(ex.Type, "test message"),
		"type (primary issue title) starts with the error string: "+ex.Type)
	assert.True(t, strings.HasPrefix(ex.Value, fmt.Sprintf("%s.%s:%d", pkgName, methodName, errorLine)),
		"value (issue subtitle) starts with the method name and error line: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 2, "two stacktrace frames")

	for _, fr := range ex.Stacktrace.Frames {
		assert.True(t, strings.HasSuffix(fr.Function, methodName), "function name has suffix: "+fr.Function)
		assert.Equal(t, errorLine, fr.Lineno, "line number matches")
		assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	}

	ex = e.Exception[2] // third exception contains the outer frame from the called error
	assert.True(t, strings.HasPrefix(ex.Type, "test message"),
		"type (primary issue title) starts with the error string: "+ex.Type)
	assert.True(t, strings.HasPrefix(ex.Value, fmt.Sprintf("%s.%s:%d", pkgName, methodName, errorLine)),
		"value (issue subtitle) starts with the method name and error line: "+ex.Value)
	assert.NotNil(t, ex.Stacktrace)
	assert.Len(t, ex.Stacktrace.Frames, 3, "three stacktrace frames")
	for _, fr := range ex.Stacktrace.Frames {
		assert.True(t, strings.HasSuffix(fr.Function, methodName), "function name has suffix: "+fr.Function)
		assert.True(t, strings.HasSuffix(fr.AbsPath, fileNameSuffix), "abspath matches: "+fr.AbsPath)
	}
	assert.Equal(t, errorWrappedLine, ex.Stacktrace.Frames[0].Lineno, "first frame line number matches")
	assert.Equal(t, errorWrappedLine, ex.Stacktrace.Frames[1].Lineno, "second frame line number matches")
	assert.Equal(t, errorLine, ex.Stacktrace.Frames[2].Lineno, "third frame line number matches")
}
