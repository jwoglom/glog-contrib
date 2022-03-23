package sentry

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/getsentry/sentry-go"
)

func printException(ex sentry.Exception) {
	fmt.Printf("Type: %s\n", ex.Type)
	fmt.Printf("Value: %s\n", ex.Value)
	if ex.Stacktrace == nil {
		fmt.Printf("Stacktrace.Frames: nil")
	} else {
		fmt.Printf("Stacktrace.Frames:\n")
		for i, f := range ex.Stacktrace.Frames {
			fmt.Printf("\t%d:\n", i)
			fmt.Printf("\t\tFunction: %s (line %d col %d)\n", f.Function, f.Lineno, f.Colno)
			fmt.Printf("\t\tModule: %s Package: %s\n", f.Module, f.Package)
			fmt.Printf("\t\tFilename: %s Abspath: %s\n", f.Filename, f.AbsPath)
		}
	}
	fmt.Printf("\n")
}

func hash(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func hashStacktrace(s *sentry.Stacktrace) string {
	if hasFrames(s) {
		return hash(s.Frames)
	}

	return "nil"
}

func hasFrames(s *sentry.Stacktrace) bool {
	return s != nil && len(s.Frames) > 0
}

func valueBeforeLastParen(v string) string {
	if !strings.Contains(v, " (") {
		return v
	}

	i := strings.LastIndex(v, " (")
	return v[:i]
}

func (d *ExceptionDeduplicator) shouldDropExceptionWithNoFrames(ids []exceptionIndex) bool {
	// If an exception exists with this stack frame, and we have an equivalent
	// type already present (with value empty), then:
	// - if another entry contains an exception, drop our exception in favor of it.
	// - if no other entries contain an exception, drop our exception only if all
	//   of the other exceptions have not been dropped.
	foundNonDropped := 0
	foundWithFrames := 0
	for _, mid := range ids {
		if otherEx := d.exceptions[mid]; hasFrames(otherEx.Stacktrace) {
			foundWithFrames++
		} else if _, dropped := d.droppedExceptions[mid]; !dropped {
			foundNonDropped++
		}
	}
	fmt.Printf("shouldDropExceptionWithNoFrames withFrames: %d nonDropped: %d\n", foundWithFrames, foundNonDropped)
	return foundWithFrames > 0 || foundNonDropped > 0
}

func (d *ExceptionDeduplicator) shouldDropExceptionWithFrames(ids []exceptionIndex) bool {
	// If an exception exists with this stack frame, and we have an equivalent
	// type already present (with value empty), then:
	// - if another entry contains an exception, drop our exception in favor of it.
	// - if no other entries contain an exception, drop our exception only if all
	//   of the other exceptions have not been dropped.
	foundNonDropped := 0
	for _, mid := range ids {
		if _, dropped := d.droppedExceptions[mid]; !dropped {
			foundNonDropped++
		}
	}
	fmt.Printf("shouldDropExceptionWithFrames nonDropped: %d\n", foundNonDropped)
	return foundNonDropped > 0
}

type exceptionIndex int
type frameIndex int

type exceptionFrameIndex struct {
	exception exceptionIndex
	frame     frameIndex
}

type ExceptionDeduplicator struct {
	exceptions []sentry.Exception

	typesMap      map[string][]exceptionIndex
	valuesMap     map[string][]exceptionIndex
	framesMap     map[string][]exceptionIndex
	frameGroupMap map[string][]exceptionFrameIndex

	droppedExceptions      map[exceptionIndex]interface{}
	droppedExceptionFrames map[exceptionFrameIndex]interface{}
}

func NewExceptionDeduplicator(exceptions []sentry.Exception) *ExceptionDeduplicator {
	d := &ExceptionDeduplicator{exceptions: exceptions}
	d.processExceptions()
	return d
}

func (d *ExceptionDeduplicator) processExceptions() {
	d.typesMap = make(map[string][]exceptionIndex)
	d.valuesMap = make(map[string][]exceptionIndex)
	d.framesMap = make(map[string][]exceptionIndex)
	d.frameGroupMap = make(map[string][]exceptionFrameIndex)
	for i, e := range d.exceptions {
		exIndex := exceptionIndex(i)
		// Make a look-up between the type/value and the exception index
		d.typesMap[e.Type] = append(d.typesMap[e.Type], exIndex)
		d.valuesMap[e.Value] = append(d.valuesMap[e.Value], exIndex)

		// Add a look-up between the value, stripped before the last parentheses,
		// and the exception index. This strips away "(methodName:lineNumber)" suffixes.
		valBeforeParen := valueBeforeLastParen(e.Value)
		d.valuesMap[valBeforeParen] = append(d.valuesMap[valBeforeParen], exIndex)

		// Make a look-up between a hash of the stacktrace and the exception index
		stacktraceHash := hashStacktrace(e.Stacktrace)
		d.framesMap[stacktraceHash] = append(d.framesMap[stacktraceHash], exIndex)

		if hasFrames(e.Stacktrace) {
			// Loop through the set of frames backwards from the end,
			// and add a hash of the frames from the end to that point
			// to the frame map. e.g., adds: [3], [3,2], [3,2,1] for an
			// exception with frames 1 2 3.
			fmt.Printf("frameGroupMap add\n")
			var frames []sentry.Frame
			for j := len(e.Stacktrace.Frames) - 1; j >= 0; j-- {
				frames = append(frames, e.Stacktrace.Frames[j])
				fmt.Printf("\t%d: %d,%d\n", j, exIndex, frameIndex(j))

				frHash := hash(frames)
				frIndex := frameIndex(j)
				d.frameGroupMap[frHash] = append(d.frameGroupMap[frHash], exceptionFrameIndex{exIndex, frIndex})
			}
		}
	}

	d.droppedExceptions = make(map[exceptionIndex]interface{})
	d.droppedExceptionFrames = make(map[exceptionFrameIndex]interface{})

}

func (d *ExceptionDeduplicator) findMatchingTypes(typeIds, valIds []exceptionIndex, excluding exceptionIndex) []exceptionIndex {
	matchedIds := make(map[exceptionIndex]interface{})
	for _, tid := range typeIds {
		for _, vid := range valIds {
			// Find exceptions containing both a matching type and value, excluding ourselves
			if tid == vid && tid != excluding {
				matchedIds[tid] = nil
			}
		}
	}

	return exceptionIndexMapToArr(matchedIds)
}

func exceptionIndexMapToArr(ids map[exceptionIndex]interface{}) []exceptionIndex {
	var arr []exceptionIndex
	for id, _ := range ids {
		arr = append(arr, id)
	}
	return arr
}


func (d *ExceptionDeduplicator) dedupNames() {
	// If exceptions exist which match, then drop the less specific of the two exceptions
	// wherever possible. If the exceptions are identical, then drop all but one.
	// The lowest IDs corresponding to exceptions present earlier in the exceptions list
	// are dropped first.

	for frHash, ids := range d.framesMap {
		fmt.Printf("frame hash %s has %d ids: %#v\n", frHash, len(ids), ids)
		// Work to remove all exceptions which have no stacktrace, assuming a similar
		// exception exists that contains the same data or a more detailed stacktrace
		if frHash == "nil" {
			for _, i := range ids {
				e := d.exceptions[i]
				if len(e.Value) == 0 {
					// If this exception has no value attribute, then check first for identical
					// exceptions containing the same Type, ignoring the value.
					// Since the given exception already has this type, only check if there is
					// more than one result.
					if ids, ok := d.typesMap[e.Type]; ok && len(ids) > 1 {
						if d.shouldDropExceptionWithNoFrames(ids) {
							fmt.Printf("emptyValueMatchingType dropping %d\n", i)
							d.droppedExceptions[i] = nil
						}
					} else if ids, ok := d.valuesMap[e.Type]; ok && len(ids) > 0 {
						// Check for exceptions which have an identical value to this type.
						// Since the given exception does not match this criteria, check if there
						// are any results.
						if d.shouldDropExceptionWithNoFrames(ids) {
							fmt.Printf("emptyValueMatchingValue dropping %d\n", i)
							d.droppedExceptions[i] = nil
						}
					}
				} else {
					// If an exception exists with this stack frame, and we have an equivalent
					// type + value already present, then:
					// - if another entry contains an exception, drop our exception in favor of it.
					// - if no other entries contain an exception, drop our exception only if all of
					//   the other exceptions have not been dropped.
					if typeIds, ok := d.typesMap[e.Type]; ok && len(typeIds) > 1 {
						if valIds, ok := d.valuesMap[e.Value]; ok && len(valIds) > 0 {
							matchedIds := d.findMatchingTypes(typeIds, valIds, i)
							if d.shouldDropExceptionWithNoFrames(matchedIds) {
								fmt.Printf("dropping %d\n", i)
								d.droppedExceptions[i] = nil
							}
						}
					}
				}
			}
		} else if len(ids) > 1 {
			// If exceptions have identical frames, with the same type + value,
			// then drop one of the exceptions.
			for _, i := range ids {
				e := d.exceptions[i]
				// If an exception exists with this stack frame, and we have an equivalent
				// type + value already present, then:
				// - if another entry contains an exception, drop our exception in favor of it.
				// - if no other entries contain an exception, drop our exception only if all of
				//   the other exceptions have not been dropped.
				if typeIds, ok := d.typesMap[e.Type]; ok && len(typeIds) > 1 {
					if valIds, ok := d.valuesMap[e.Value]; ok && len(valIds) > 0 {
						matchedIds := d.findMatchingTypes(typeIds, valIds, i)
						if d.shouldDropExceptionWithFrames(matchedIds) {
							fmt.Printf("dropping %d\n", i)
							d.droppedExceptions[i] = nil
						}
					}
				}
			}
		}
	}
}

func (d *ExceptionDeduplicator) dedupFrames() {
	for groupHash, indexes := range d.frameGroupMap {
		if len(indexes) > 1 {
			fmt.Printf("groupHash: %s indexes: %+v\n", groupHash, indexes)
			// Drop all except for the last index, which will be for the highest-numbered exception
			for i := 0; i < len(indexes) - 1; i++ {
				// Drop the exception, frame pair.
				orig := indexes[i]
				d.droppedExceptionFrames[orig] = nil

				// Drop all of the frames before the given frame for the given exception.
				for j := 0; j < int(orig.frame); j++ {
					f := exceptionFrameIndex{exception: orig.exception, frame: frameIndex(j)}
					d.droppedExceptionFrames[f] = nil
				}
			}
		}
	}
}

func (d *ExceptionDeduplicator) Dedup() []sentry.Exception {
	// Deduplicate based on name and frame contents
	d.dedupNames()
	d.dedupFrames()

	// Remove dropped frames and exceptions from d.exceptions,
	// then reprocess the input.
	d.drop()
	d.processExceptions()

	// Once dropped, process any remaining duplicate names
	d.dedupNames()

	// Remove any additionally dropped frames
	d.drop()

	return d.exceptions
}

func (d *ExceptionDeduplicator) drop() {
	// Removes any dropped frames from the stacktrace, and sets d.exceptions to the new stacktrace set
	var returnedExceptions []sentry.Exception
	for i, e := range d.exceptions {
		if _, ok := d.droppedExceptions[exceptionIndex(i)]; !ok {
			if e.Stacktrace != nil {
				var frames []sentry.Frame
				for f := 0; f < len(e.Stacktrace.Frames); f++ {
					if _, drop := d.droppedExceptionFrames[exceptionFrameIndex{exception: exceptionIndex(i), frame: frameIndex(f)}]; !drop {
						fmt.Printf("dropping %d %d\n", i, f)
						frames = append(frames, e.Stacktrace.Frames[f])
					}
				}
				e.Stacktrace.Frames = frames
			}
			returnedExceptions = append(returnedExceptions, e)
		}
	}
	d.exceptions = returnedExceptions
}

func DedupExceptions(exceptions []sentry.Exception) []sentry.Exception {
	fmt.Println("DEDUP input:")
	for _, e := range exceptions {
		printException(e)
	}

	dedup := NewExceptionDeduplicator(exceptions)

	returnedExceptions := dedup.Dedup()
	fmt.Println("DEDUP output:")
	for _, e := range returnedExceptions {
		printException(e)
	}
	return returnedExceptions
}
