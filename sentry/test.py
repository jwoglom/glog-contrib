import json
import collections

class Graph:
    def __init__(self):
        # hash(frame) -> set(hash(frame))
        self.neighbors = collections.defaultdict(set)
        # hash(frame) -> set(hash(exception_details))
        self.exception_context = collections.defaultdict(set)
        # list(hash(frame)); keep ordered
        self.starting_frames = list()
        self.hash_lookups = {}
    
    def hsh(self, frame):
        if 'filename' in frame and 'lineno' in frame:
            return '%s#%d' % (frame['filename'], frame['lineno'])

        return hash(json.dumps(frame, sort_keys=True))
    
    def connect_frames(self, a, b):
        ah = self.hsh(a)
        bh = self.hsh(b)
        
        if ah not in self.hash_lookups:
            self.hash_lookups[ah] = a

        if bh not in self.hash_lookups:
            self.hash_lookups[bh] = b
        
        self.neighbors[ah].add(bh)
    
    def start_frame(self, frame, raw_exception):
        fh = self.hsh(frame)

        exception = {"type": raw_exception["type"], "value": raw_exception["value"]}
        eh = self.hsh(exception)        
        self.hash_lookups[eh] = exception

        self.exception_context[fh].add(eh)
        self.starting_frames.append(fh)
    
    def get(self, h):
        return self.hash_lookups.get(h)
    
    def loop(self):
        seen = set()
        for f in self.starting_frames:
            if f in seen:
                continue
            seen.add(f)

            print("start: ")
            for i in self.exception_context.get(f):
                print("\tctx: %s" % self.get(i))

            self.loop_frames(f)
    
    def loop_frames(self, start):
        todo = [(start,)]

        fullPaths = []

        # DFS (TODO prevent loops)
        while todo:
            path = todo.pop(0)
            lastI = path[-1]
            lastEl = self.get(path[-1])
            

            # process
            print("frame: %s" % lastEl)

            neighbors = self.neighbors[lastI]
            if len(neighbors) == 0:
                # we've reached the end of a path
                fullPaths.append(path)

            for n in neighbors:
                todo.append(path + (n,))

        # condense paths that start with the same indicies
        # start with the longest tuple
        longestPaths = sorted(fullPaths, key=lambda x: -len(x))
        dedupPaths = list(fullPaths)
        for p in longestPaths:
            for i in range(1, len(p)):
                if p[:i] in dedupPaths:
                    dedupPaths.remove(p[:i])


        
        for p in dedupPaths:
            print("\nfull path:\n")
            for e in p:
                print("\t%s" % self.get(e))

            

            


def process_frames(values):
    graph = Graph()

    for exception in values:
        if not "stacktrace" in exception:
            print("No stacktraces: %s" % exception)
            graph.start_frame({}, exception)
            continue

        frames = exception["stacktrace"]["frames"]
        is_first_frame = True
        if len(frames) == 1:
            graph.start_frame(frames[0], exception)
            continue

        # loop through frames backwards, skipping the first
        for i in range(len(frames)-1, 0, -1):
            cur_frame = frames[i]
            prev_frame = frames[i-1]

            # for the first frame, add it to starting_frames
            if is_first_frame:
                graph.start_frame(cur_frame, exception)
                is_first_frame = False
            
            graph.connect_frames(cur_frame, prev_frame)
    
    print("\n")
    graph.loop()


def _process_frames(values):
    # hashes a dict
    def hsh(d):
        return hash(json.dumps(d))
    
    def dsp(d):
        return '%s#%s' % (d['filename'], d['lineno'])
    
    starting_frames = set()
    starting_frames_map = {}
    starting_frame_exceptions = collections.defaultdict(list)
    frame_map = collections.defaultdict(list)
    no_stacktraces = set()
    no_stacktraces_map = {}

    for exception in values:
        if not "stacktrace" in exception:
            print("No stacktraces: %s" % exception)
            
            continue
        frames = exception["stacktrace"]["frames"]
        is_first_frame = True
        # loop through frames backwards, skipping the first
        for i in range(len(frames)-1, 0, -1):
            cur_frame = frames[i]
            prev_frame = frames[i-1]

            # for the first frame, add it to starting_frames
            if is_first_frame:
                starting_frame_exceptions[hsh(cur_frame)] = {
                    "type": exception["type"],
                    "value": exception["type"]
                }
                starting_frames.add(hsh(cur_frame))
                starting_frames_map[hsh(cur_frame)] = cur_frame
                is_first_frame = False
            
            frame_map[hsh(cur_frame)].append(prev_frame)
    

    seen_hashes = set()
    def recur(frame, i=0):
        if not frame:
            return
        if hsh(frame) in seen_hashes:
            return
        seen_hashes.add(hsh(frame))
        print("%s%d: %s" % ('  '*i, i, dsp(frame)))
        nexts = frame_map[hsh(frame)]
        for n in nexts:
            recur(n, i=i+1)
    
    for h in starting_frames:
        s = starting_frames_map[h]
        print("Start exc: %s" % starting_frame_exceptions[hsh(s)])
        recur(s, 0)
    






        




# python is basically json, right?
true = True
false = False
data = {
    "values": [
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1) (logServerError:97)",
            "stacktrace": {
                "frames": [
                    {
                        "function": "(*Server).serveStreams.func1.2",
                        "module": "google.golang.org/grpc",
                        "filename": "external/org_golang_google_grpc/server.go",
                        "abs_path": "external/org_golang_google_grpc/server.go",
                        "lineno": 878,
                        "in_app": true
                    },
                    {
                        "function": "(*Server).handleStream",
                        "module": "google.golang.org/grpc",
                        "filename": "external/org_golang_google_grpc/server.go",
                        "abs_path": "external/org_golang_google_grpc/server.go",
                        "lineno": 1540,
                        "in_app": true
                    },
                    {
                        "function": "(*Server).processUnaryRPC",
                        "module": "google.golang.org/grpc",
                        "filename": "external/org_golang_google_grpc/server.go",
                        "abs_path": "external/org_golang_google_grpc/server.go",
                        "lineno": 1217,
                        "in_app": true
                    },
                    {
                        "function": "_PluginService_InvokePlugin_Handler",
                        "module": "alpha/src/com/yext/platform/plugins",
                        "filename": "com/yext/platform/plugins/plugins.pb.go",
                        "abs_path": "bazel-out/k8-fastbuild/bin/src/com/yext/platform/plugins/plugins_go_proto.withoutprotosourcefiles_/alpha/src/com/yext/platform/plugins/plugins.pb.go",
                        "lineno": 980,
                        "in_app": true
                    },
                    {
                        "function": "ChainUnaryServer.func1",
                        "module": "github.com/grpc-ecosystem/go-grpc-middleware",
                        "filename": "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
                        "abs_path": "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
                        "lineno": 34,
                        "in_app": true
                    },
                    {
                        "function": "ChainUnaryServer.func1.1.1",
                        "module": "github.com/grpc-ecosystem/go-grpc-middleware",
                        "filename": "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
                        "abs_path": "external/com_github_grpc_ecosystem_go_grpc_middleware/chain.go",
                        "lineno": 25,
                        "in_app": true
                    },
                    {
                        "function": "InterceptServerUnary",
                        "module": "yext/net/grpc/grpctrace",
                        "filename": "yext/net/grpc/grpctrace/server.go",
                        "abs_path": "gocode/src/yext/net/grpc/grpctrace/server.go",
                        "lineno": 53,
                        "in_app": true
                    },
                    {
                        "function": "logServerError",
                        "module": "yext/net/grpc/grpctrace",
                        "filename": "yext/net/grpc/grpctrace/server.go",
                        "abs_path": "gocode/src/yext/net/grpc/grpctrace/server.go",
                        "lineno": 97,
                        "in_app": true
                    }
                ]
            }
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1)"
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1) (yext/platform/plugins/internal/runtime.(*server).Status:213)",
            "stacktrace": {
                "frames": [
                    {
                        "function": "(*server).ready",
                        "module": "yext/platform/plugins/internal/runtime",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 436,
                        "in_app": true
                    },
                    {
                        "function": "(*server).Status",
                        "module": "yext/platform/plugins/internal/runtime",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 213,
                        "in_app": true
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Status",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 213,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    }
                ]
            }
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1) (yext/platform/plugins/internal/runtime.(*server).Status:213)",
            "stacktrace": {
                "frames": [
                    {
                        "function": "(*server).awaitLoaded",
                        "module": "yext/platform/plugins/internal/runtime",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 412,
                        "in_app": true
                    },
                    {
                        "function": "(*server).ready",
                        "module": "yext/platform/plugins/internal/runtime",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 443,
                        "in_app": true
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).ready",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 443,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Status",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 213,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    }
                ]
            }
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1) (yext/platform/plugins/internal/runtime.(*server).Status:213)",
            "stacktrace": {
                "frames": [
                    {
                        "function": "(*Plugin).Invoke",
                        "module": "yext/platform/plugins/internal/runtime",
                        "filename": "yext/platform/plugins/internal/runtime/plugin.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin.go",
                        "lineno": 247,
                        "in_app": true
                    },
                    {
                        "function": "(*server).Load",
                        "module": "yext/platform/plugins/internal/runtime",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 171,
                        "in_app": true
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Load",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 171,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).ready",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 443,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Status",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 213,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    }
                ]
            }
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1) (yext/platform/plugins/internal/runtime.(*server).Status:213)",
            "stacktrace": {
                "frames": [
                    {
                        "function": "(*Server).InvokePlugin",
                        "module": "yext/platform/plugins/internal/server",
                        "filename": "yext/platform/plugins/internal/server/server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/server/server.go",
                        "lineno": 154,
                        "in_app": true
                    },
                    {
                        "function": "(*Plugin).Invoke",
                        "module": "yext/platform/plugins/internal/runtime",
                        "filename": "yext/platform/plugins/internal/runtime/plugin.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin.go",
                        "lineno": 248,
                        "in_app": true
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*Plugin).Invoke",
                        "filename": "yext/platform/plugins/internal/runtime/plugin.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin.go",
                        "lineno": 248,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Load",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 171,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).ready",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 443,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Status",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 213,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    }
                ]
            }
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1)"
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1) (yext/platform/plugins/internal/runtime.(*server).Status:213)",
            "stacktrace": {
                "frames": [
                    {
                        "function": "_PluginService_InvokePlugin_Handler.func1",
                        "module": "alpha/src/com/yext/platform/plugins",
                        "filename": "com/yext/platform/plugins/plugins.pb.go",
                        "abs_path": "bazel-out/k8-fastbuild/bin/src/com/yext/platform/plugins/plugins_go_proto.withoutprotosourcefiles_/alpha/src/com/yext/platform/plugins/plugins.pb.go",
                        "lineno": 978,
                        "in_app": true
                    },
                    {
                        "function": "(*Server).InvokePlugin",
                        "module": "yext/platform/plugins/internal/server",
                        "filename": "yext/platform/plugins/internal/server/server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/server/server.go",
                        "lineno": 160,
                        "in_app": true
                    },
                    {
                        "function": "yext/platform/plugins/internal/server.(*Server).InvokePlugin",
                        "filename": "yext/platform/plugins/internal/server/server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/server/server.go",
                        "lineno": 160,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*Plugin).Invoke",
                        "filename": "yext/platform/plugins/internal/runtime/plugin.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin.go",
                        "lineno": 248,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Load",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 171,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).ready",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 443,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    },
                    {
                        "function": "yext/platform/plugins/internal/runtime.(*server).Status",
                        "filename": "yext/platform/plugins/internal/runtime/plugin_server.go",
                        "abs_path": "gocode/src/yext/platform/plugins/internal/runtime/plugin_server.go",
                        "lineno": 213,
                        "in_app": false,
                        "data": {
                            "orig_in_app": -1
                        }
                    }
                ]
            }
        },
        {
            "type": "status \"InternalError\"",
            "value": "Worker stopped unexpectedly (exit code: 1)"
        }
    ]
}

if __name__ == '__main__':
    process_frames(data["values"])