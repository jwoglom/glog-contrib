package sentry_test

import (
	"flag"
	"runtime"
	"time"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/kr/pretty"
	"github.com/yext/glog"
	"github.com/yext/glog-contrib/sentry"
)

const pkgName = "github.com/yext/glog-contrib/sentry_test" // this should stay in sync with the location/pkg of the folder
const fileNameSuffix = "_test.go"                          // this should stay in sync with the name of this file

var sendToDsn = flag.String("sendToDsn", "",
	"optional sentry DSN. if set, sample exceptions will be sent to Sentry as an integration test")

var logEvents = flag.Bool("logEvents", false,
	"if set, full log messages will be pretty-printed to the screen")

func setup(ready chan interface{}, done chan *sentrygo.Event, count int, dedup bool) {
	sentry.CaptureErrors(
		"example",
		[]string{*sendToDsn},
		sentrygo.ClientOptions{Debug: true},
		wrapper(ready, done, count, dedup, glog.RegisterBackend()))
}

func wrapper(ready chan interface{}, done chan *sentrygo.Event, count int, dedup bool, ch <-chan glog.Event) <-chan glog.Event {
	ready <- nil
	i := 0
	wrap := make(chan glog.Event)
	go func() {
		for glogEvent := range ch {
			if *logEvents {
				pretty.Log("glog event:", glogEvent)
			}
			// If a DSN is provided, run as an integration test which forwards
			// the glog event to the channel used by sentry.CaptureErrors
			if *sendToDsn != "" {
				wrap <- glogEvent
				// Give sentry time to process the event.
				// If removed, on test failure Sentry won't flush its cache
				time.Sleep(2000 * time.Millisecond)
			}
			if glogEvent.Severity == "ERROR" {
				e, _ := sentry.FromGlogEvent(glogEvent, dedup)
				if *logEvents {
					pretty.Log("Sentry event:", e)
				}
				done <- e
				i++
				// If we've seen the total number of expected events, break
				if i == count {
					break
				}
			}
		}
	}()

	return wrap
}

// Returns the current line on which the method is called
func currentLine() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}
