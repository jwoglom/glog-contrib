package sentry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"

	"github.com/yext/glog-contrib/sentry"
)

func TestDedupMergeException_EquivalentTypeAndValue_KeepsSingleStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:  "type",
		Value: "value",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: []sentrygo.Frame{{
				Filename: "filename",
			}},
		},
	}, {
		Type:       "type",
		Value:      "value",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "one exception merged")

	assert.Equal(t, "type", output[0].Type, "type matches")
	assert.Equal(t, "value", output[0].Value, "value matches")
	assert.NotNil(t, output[0].Stacktrace, "non-nil stacktrace")

	assert.Equal(t, "othertype", output[1].Type, "other type exists")
	assert.Equal(t, "othervalue", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_EquivalentTypeAndValue_NoStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:       "type",
		Value:      "value",
		Stacktrace: nil,
	}, {
		Type:       "type",
		Value:      "value",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "one exception merged")

	assert.Equal(t, "type", output[0].Type, "type matches")
	assert.Equal(t, "value", output[0].Value, "value matches")
	assert.Nil(t, output[0].Stacktrace, "nil stacktrace")

	assert.Equal(t, "othertype", output[1].Type, "other type exists")
	assert.Equal(t, "othervalue", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_EquivalentType_NoValue_KeepsSingleStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:  "type",
		Value: "value",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: []sentrygo.Frame{{
				Filename: "filename",
			}},
		},
	}, {
		Type:       "type",
		Value:      "",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "one exception merged")

	assert.Equal(t, "type", output[0].Type, "type matches")
	assert.Equal(t, "value", output[0].Value, "value matches")
	assert.NotNil(t, output[0].Stacktrace, "non-nil stacktrace")

	assert.Equal(t, "othertype", output[1].Type, "other type exists")
	assert.Equal(t, "othervalue", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_EquivalentType_NoValue_NoStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:       "type",
		Value:      "value",
		Stacktrace: nil,
	}, {
		Type:       "type",
		Value:      "",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "one exception merged")

	assert.Equal(t, "type", output[0].Type, "type matches")
	assert.Equal(t, "value", output[0].Value, "value matches")
	assert.Nil(t, output[0].Stacktrace, "nil stacktrace")

	assert.Equal(t, "othertype", output[1].Type, "other type exists")
	assert.Equal(t, "othervalue", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_TypeMatchingOtherValue_NoStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:       "i/o timeout",
		Value:      "",
		Stacktrace: nil,
	}, {
		Type:       "dial tcp 1.1.1.1:1111",
		Value:      "i/o timeout",
		Stacktrace: nil,
	}, {
		Type:       "dial tcp 1.1.1.1:1111",
		Value:      "i/o timeout",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 1, "all exceptions merged")

	assert.Equal(t, "dial tcp 1.1.1.1:1111", output[0].Type, "type matches")
	assert.Equal(t, "i/o timeout", output[0].Value, "value matches")
	assert.Nil(t, output[0].Stacktrace, "nil stacktrace")
}

func TestDedupMergeException_TypeMatchingOtherValueAndType_NoStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:       "i/o timeout",
		Value:      "i/o timeout",
		Stacktrace: nil,
	}, {
		Type:       "i/o timeout",
		Value:      "i/o timeout",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 1, "all exceptions merged")

	assert.Equal(t, "i/o timeout", output[0].Type, "type matches")
	assert.Equal(t, "i/o timeout", output[0].Value, "value matches")
	assert.Nil(t, output[0].Stacktrace, "nil stacktrace")
}

func TestDedupMergeException_EquivalentType_ValueWithShortContext_KeepsSingleStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:  "type",
		Value: "value (fooBar:123)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: []sentrygo.Frame{{
				Filename: "filename",
			}},
		},
	}, {
		Type:       "type",
		Value:      "value",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "one exception merged")

	assert.Equal(t, "type", output[0].Type, "type matches")
	assert.Equal(t, "value (fooBar:123)", output[0].Value, "value matches")
	assert.NotNil(t, output[0].Stacktrace, "non-nil stacktrace")

	assert.Equal(t, "othertype", output[1].Type, "other type exists")
	assert.Equal(t, "othervalue", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_EquivalentType_ValueWithLongContext_KeepsSingleStacktrace(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:  "type",
		Value: "value (exit code: 1) (yext/foo/bar.(*server).Status:123)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: []sentrygo.Frame{{
				Filename: "filename",
			}},
		},
	}, {
		Type:       "type",
		Value:      "value (exit code: 1)",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "one exception merged")

	assert.Equal(t, "type", output[0].Type, "type matches")
	assert.Equal(t, "value (exit code: 1) (yext/foo/bar.(*server).Status:123)", output[0].Value, "value matches")
	assert.NotNil(t, output[0].Stacktrace, "non-nil stacktrace")

	assert.Equal(t, "othertype", output[1].Type, "other type exists")
	assert.Equal(t, "othervalue", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_EquivalentType_ValueWithDifferentContext(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:  "type",
		Value: "value (exit code: 1) (yext/foo/bar.(*server).Status:123)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: []sentrygo.Frame{{
				Filename: "filename",
			}},
		},
	}, {
		Type:       "type",
		Value:      "value (exit code: 2)",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 3, "one exception merged")

	assert.Equal(t, "type", output[0].Type, "type matches")
	assert.Equal(t, "value (exit code: 1) (yext/foo/bar.(*server).Status:123)", output[0].Value, "value matches")
	assert.NotNil(t, output[0].Stacktrace, "non-nil stacktrace")

	assert.Equal(t, "type", output[1].Type, "other type exists")
	assert.Equal(t, "value (exit code: 2)", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")

	assert.Equal(t, "othertype", output[2].Type, "other type exists")
	assert.Equal(t, "othervalue", output[2].Value, "other value exists")
	assert.Nil(t, output[2].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_ExceptionWithNoStacktrace_KeepsBoth(t *testing.T) {
	input := []sentrygo.Exception{{
		Type:       "type",
		Value:      "value (exit code: 2)",
		Stacktrace: nil,
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "no exceptions merged")

	assert.Equal(t, "type", output[0].Type, "other type exists")
	assert.Equal(t, "value (exit code: 2)", output[0].Value, "other value exists")
	assert.Nil(t, output[0].Stacktrace, "other exception still has nil stacktrace")

	assert.Equal(t, "othertype", output[1].Type, "other type exists")
	assert.Equal(t, "othervalue", output[1].Value, "other value exists")
	assert.Nil(t, output[1].Stacktrace, "other exception still has nil stacktrace")
}

func TestDedupMergeException_ExceptionWithIdenticalStacktraces_IdenticalTypeAndValue(t *testing.T) {
	frames := []sentrygo.Frame{{
		Filename: "filename",
		Lineno:   123,
	}, {
		Filename: "filename2",
		Lineno:   432,
	}}
	input := []sentrygo.Exception{{
		Type:  "type",
		Value: "value (exit code: 2)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: frames,
		},
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}, {
		Type:  "type",
		Value: "value (exit code: 2)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: frames,
		},
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "identical stacktrace and type+value exceptions merged")

	assert.Equal(t, "othertype", output[0].Type, "other type exists")
	assert.Equal(t, "othervalue", output[0].Value, "other value exists")
	assert.Nil(t, output[0].Stacktrace, "other exception still has nil stacktrace")

	assert.Equal(t, "type", output[1].Type, "type exists")
	assert.Equal(t, "value (exit code: 2)", output[1].Value, "value exists")
	assert.NotNil(t, output[1].Stacktrace, "exception has stacktrace")
	assert.Equal(t, frames, output[1].Stacktrace.Frames, "frames match")
}

func TestDedupMergeException_ExceptionWithIdenticalStacktraces_ShouldMergeTwice(t *testing.T) {
	frames := []sentrygo.Frame{{
		Filename: "filename",
		Lineno:   123,
	}, {
		Filename: "filename2",
		Lineno:   432,
	}}
	input := []sentrygo.Exception{{
		Type:  "type",
		Value: "value (exit code: 2)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: frames,
		},
	}, {
		Type:       "othertype",
		Value:      "othervalue",
		Stacktrace: nil,
	}, {
		Type:  "type",
		Value: "value (exit code: 2)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: frames,
		},
	}, {
		Type:  "type",
		Value: "value (exit code: 2)",
		Stacktrace: &sentrygo.Stacktrace{
			Frames: frames,
		},
	}}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "identical stacktrace and type+value exceptions merged")

	assert.Equal(t, "othertype", output[0].Type, "other type exists")
	assert.Equal(t, "othervalue", output[0].Value, "other value exists")
	assert.Nil(t, output[0].Stacktrace, "other exception still has nil stacktrace")

	assert.Equal(t, "type", output[1].Type, "type exists")
	assert.Equal(t, "value (exit code: 2)", output[1].Value, "value exists")
	assert.NotNil(t, output[1].Stacktrace, "exception has stacktrace")
	assert.Equal(t, frames, output[1].Stacktrace.Frames, "frames match")

}

func TestDedupMergeException_SimulatedExample1(t *testing.T) {
	input := []sentrygo.Exception{
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (logServerError:97)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*Server).serveStreams.func1.2",
						Module:   "google.golang.org/grpc",
						Filename: "external/org_golang_google_grpc/server.go",
						AbsPath:  "external/org_golang_google_grpc/server.go",
						Lineno:   878,
						InApp:    true,
					},
					{
						Function: "(*Server).handleStream",
						Module:   "google.golang.org/grpc",
						Filename: "external/org_golang_google_grpc/server.go",
						AbsPath:  "external/org_golang_google_grpc/server.go",
						Lineno:   1540,
						InApp:    true,
					},
					{
						Function: "(*Server).processUnaryRPC",
						Module:   "google.golang.org/grpc",
						Filename: "external/org_golang_google_grpc/server.go",
						AbsPath:  "external/org_golang_google_grpc/server.go",
						Lineno:   1217,
						InApp:    true,
					},
					{
						Function: "_exampleService_Invokeexample_Handler",
						Module:   "alpha/src/com/yext/platform/examples",
						Filename: "com/yext/foo/bar/examples.pb.go",
						AbsPath:  "bazel-out/k8-fastbuild/bin/src/com/yext/foo/bar/examples_go_proto.withoutprotosourcefiles_/alpha/src/com/yext/foo/bar/examples.pb.go",
						Lineno:   980,
						InApp:    true,
					},
					{
						Function: "ChainUnaryServer.func1",
						Module:   "github.com/grpc-ecosystem/go-grpc-middleware",
						Filename: "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						AbsPath:  "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						Lineno:   34,
						InApp:    true,
					},
					{
						Function: "ChainUnaryServer.func1.1.1",
						Module:   "github.com/grpc-ecosystem/go-grpc-middleware",
						Filename: "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						AbsPath:  "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						Lineno:   25,
						InApp:    true,
					},
					{
						Function: "InterceptServerUnary",
						Module:   "yext/net/grpc/grpctrace",
						Filename: "yext/net/grpc/grpctrace/server.go",
						AbsPath:  "gocode/src/yext/net/grpc/grpctrace/server.go",
						Lineno:   53,
						InApp:    true,
					},
					{
						Function: "logServerError",
						Module:   "yext/net/grpc/grpctrace",
						Filename: "yext/net/grpc/grpctrace/server.go",
						AbsPath:  "gocode/src/yext/net/grpc/grpctrace/server.go",
						Lineno:   97,
						InApp:    true,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1)",
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*server).ready",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   436,
						InApp:    true,
					},
					{
						Function: "(*server).Status",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*server).awaitLoaded",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   412,
						InApp:    true,
					},
					{
						Function: "(*server).ready",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*example).Invoke",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   247,
						InApp:    true,
					},
					{
						Function: "(*server).Load",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Load",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*Server).Invokeexample",
						Module:   "yext/foo/bar/internal/server",
						Filename: "yext/foo/bar/internal/server/server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/server/server.go",
						Lineno:   154,
						InApp:    true,
					},
					{
						Function: "(*example).Invoke",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   248,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*example).Invoke",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   248,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Load",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1)",
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "_exampleService_Invokeexample_Handler.func1",
						Module:   "alpha/src/com/yext/platform/examples",
						Filename: "com/yext/foo/bar/examples.pb.go",
						AbsPath:  "bazel-out/k8-fastbuild/bin/src/com/yext/foo/bar/examples_go_proto.withoutprotosourcefiles_/alpha/src/com/yext/foo/bar/examples.pb.go",
						Lineno:   978,
						InApp:    true,
					},
					{
						Function: "(*Server).Invokeexample",
						Module:   "yext/foo/bar/internal/server",
						Filename: "yext/foo/bar/internal/server/server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/server/server.go",
						Lineno:   160,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/server.(*Server).Invokeexample",
						Filename: "yext/foo/bar/internal/server/server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/server/server.go",
						Lineno:   160,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*example).Invoke",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   248,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Load",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1)",
		},
	}

	output := sentry.DedupExceptions(input)
	require.Len(t, output, 2, "exceptions were merged")


	expected := []sentrygo.Exception{
		{
			// First exception matches identically
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (logServerError:97)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*Server).serveStreams.func1.2",
						Module:   "google.golang.org/grpc",
						Filename: "external/org_golang_google_grpc/server.go",
						AbsPath:  "external/org_golang_google_grpc/server.go",
						Lineno:   878,
						InApp:    true,
					},
					{
						Function: "(*Server).handleStream",
						Module:   "google.golang.org/grpc",
						Filename: "external/org_golang_google_grpc/server.go",
						AbsPath:  "external/org_golang_google_grpc/server.go",
						Lineno:   1540,
						InApp:    true,
					},
					{
						Function: "(*Server).processUnaryRPC",
						Module:   "google.golang.org/grpc",
						Filename: "external/org_golang_google_grpc/server.go",
						AbsPath:  "external/org_golang_google_grpc/server.go",
						Lineno:   1217,
						InApp:    true,
					},
					{
						Function: "_exampleService_Invokeexample_Handler",
						Module:   "alpha/src/com/yext/platform/examples",
						Filename: "com/yext/foo/bar/examples.pb.go",
						AbsPath:  "bazel-out/k8-fastbuild/bin/src/com/yext/foo/bar/examples_go_proto.withoutprotosourcefiles_/alpha/src/com/yext/foo/bar/examples.pb.go",
						Lineno:   980,
						InApp:    true,
					},
					{
						Function: "ChainUnaryServer.func1",
						Module:   "github.com/grpc-ecosystem/go-grpc-middleware",
						Filename: "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						AbsPath:  "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						Lineno:   34,
						InApp:    true,
					},
					{
						Function: "ChainUnaryServer.func1.1.1",
						Module:   "github.com/grpc-ecosystem/go-grpc-middleware",
						Filename: "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						AbsPath:  "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
						Lineno:   25,
						InApp:    true,
					},
					{
						Function: "InterceptServerUnary",
						Module:   "yext/net/grpc/grpctrace",
						Filename: "yext/net/grpc/grpctrace/server.go",
						AbsPath:  "gocode/src/yext/net/grpc/grpctrace/server.go",
						Lineno:   53,
						InApp:    true,
					},
					{
						Function: "logServerError",
						Module:   "yext/net/grpc/grpctrace",
						Filename: "yext/net/grpc/grpctrace/server.go",
						AbsPath:  "gocode/src/yext/net/grpc/grpctrace/server.go",
						Lineno:   97,
						InApp:    true,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*server).ready",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   436,
						InApp:    true,
					},
					{
						Function: "(*server).Status",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*server).awaitLoaded",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   412,
						InApp:    true,
					},
					{
						Function: "(*server).ready",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*example).Invoke",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   247,
						InApp:    true,
					},
					{
						Function: "(*server).Load",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Load",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "(*Server).Invokeexample",
						Module:   "yext/foo/bar/internal/server",
						Filename: "yext/foo/bar/internal/server/server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/server/server.go",
						Lineno:   154,
						InApp:    true,
					},
					{
						Function: "(*example).Invoke",
						Module:   "yext/foo/bar/internal/runtime",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   248,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*example).Invoke",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   248,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Load",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1)",
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1) (yext/foo/bar/internal/runtime.(*server).Status:213)",
			Stacktrace: &sentrygo.Stacktrace{
				Frames: []sentrygo.Frame{
					{
						Function: "_exampleService_Invokeexample_Handler.func1",
						Module:   "alpha/src/com/yext/platform/examples",
						Filename: "com/yext/foo/bar/examples.pb.go",
						AbsPath:  "bazel-out/k8-fastbuild/bin/src/com/yext/foo/bar/examples_go_proto.withoutprotosourcefiles_/alpha/src/com/yext/foo/bar/examples.pb.go",
						Lineno:   978,
						InApp:    true,
					},
					{
						Function: "(*Server).Invokeexample",
						Module:   "yext/foo/bar/internal/server",
						Filename: "yext/foo/bar/internal/server/server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/server/server.go",
						Lineno:   160,
						InApp:    true,
					},
					{
						Function: "yext/foo/bar/internal/server.(*Server).Invokeexample",
						Filename: "yext/foo/bar/internal/server/server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/server/server.go",
						Lineno:   160,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*example).Invoke",
						Filename: "yext/foo/bar/internal/runtime/example.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example.go",
						Lineno:   248,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Load",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   171,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).ready",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   443,
						InApp:    false,
					},
					{
						Function: "yext/foo/bar/internal/runtime.(*server).Status",
						Filename: "yext/foo/bar/internal/runtime/example_server.go",
						AbsPath:  "gocode/src/yext/foo/bar/internal/runtime/example_server.go",
						Lineno:   213,
						InApp:    false,
					},
				},
			},
		},
		{
			Type:  "status \"InternalError\"",
			Value: "Worker stopped unexpectedly (exit code: 1)",
		},
	}

	assert.Equal(t, expected, output)
}
