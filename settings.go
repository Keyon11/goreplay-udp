package main

import (
	"flag"
	"fmt"
	"github.com/myzhan/goreplay-udp/output"
	"time"
)

// MultiOption allows to specify multiple flags with same name and collects all values into array
type MultiOption []string

func (h *MultiOption) String() string {
	return fmt.Sprint(*h)
}

// Set gets called multiple times for each flag with same name
func (h *MultiOption) Set(value string) error {
	*h = append(*h, value)
	return nil
}

// AppSettings is the struct of main configuration
type AppSettings struct {
	exitAfter time.Duration

	splitOutput  bool
	outputStdout bool
	outputNull   bool

	inputFile        MultiOption
	inputFileLoop    bool
	outputFile       MultiOption
	outputFileConfig output.FileOutputConfig

	inputUDP              MultiOption
	inputUDPTrackResponse bool
	outputUDP             MultiOption
	outputUDPConfig       output.UDPOutputConfig

	inputHttp        MultiOption
	outputHttp       MultiOption
	outputHttpConfig output.HTTPOutputConfig
}

// Settings holds Goreplay configuration
var Settings AppSettings

func init() {
	flag.DurationVar(&Settings.exitAfter, "exit-after", 0, "exit after specified duration")

	flag.BoolVar(&Settings.splitOutput, "split-output", false, "By default each output gets same traffic. If set to `true` it splits traffic equally among all outputs")
	flag.BoolVar(&Settings.outputStdout, "output-stdout", false, "Used for testing inputs. Just prints to console data coming from inputs")
	flag.BoolVar(&Settings.outputNull, "output-null", false, "Used for testing inputs. Drops all requests")

	flag.Var(&Settings.inputFile, "input-file", "Read requests from file: \n\tgoreplay-udp --input-file ./requests.gor --output-stdout")
	flag.BoolVar(&Settings.inputFileLoop, "input-file-loop", false, "Loop input files, useful for performance testing")

	flag.Var(&Settings.outputFile, "output-file", "Write incoming requests to file: \n\tgoreplay-udp --input-udp :80 --output-file ./requests.gor")
	flag.DurationVar(&Settings.outputFileConfig.FlushInterval, "output-file-flush-interval", time.Second, "Interval for forcing buffer flush to the file, default: 1s")
	flag.BoolVar(&Settings.outputFileConfig.Append, "output-file-append", false, "The flushed chunk is appended to existence file or not")

	// Set default
	Settings.outputFileConfig.SizeLimit.Set("32mb")
	flag.Var(&Settings.outputFileConfig.SizeLimit, "output-file-size-limit", "Size of each chunk. Default: 32mb")
	flag.IntVar(&Settings.outputFileConfig.QueueLimit, "output-file-queue-limit", 25600, "The length of the chunk queue. Default: 25600")

	flag.Var(&Settings.inputUDP, "input-udp", "Capture traffic from given port (use RAW sockets and require *sudo* access):\n\t# Capture traffic from 8080 port\n\tgoreplay-udp --input-raw :8080 --output-stdout")
	flag.BoolVar(&Settings.inputUDPTrackResponse, "input-udp-track-response", false, "If turned on gorepaly-udp will track responses in addition to requests")

	flag.Var(&Settings.outputUDP, "output-udp", "Forwards incoming requests to given udp address.\n\t# Redirect all incoming requests to staging.com address \n\tgoreplay-udp --input-raw :80 --output-udp staging.com")
	flag.IntVar(&Settings.outputUDPConfig.Workers, "output-udp-workers", 0, "Goreplay-udp uses dynamic worker scaling by default.  Enter a number to run a set number of workers.")
	flag.DurationVar(&Settings.outputUDPConfig.Timeout, "output-udp-timeout", 5*time.Second, "Specify UDP request/response timeout. By default 5s. Example: --output-udp-timeout 30s")
	flag.BoolVar(&Settings.outputUDPConfig.Stats, "output-udp-stats", false, "Report udp output queue stats to console every 5 seconds")
	flag.BoolVar(&Settings.outputUDPConfig.IgnoreResponse, "output-udp-ignore-response", false, "Ignore UDP Response")

	flag.Var(&Settings.inputHttp, "input-http", "Capture traffic from given port (use RAW sockets and require *sudo* access):\n\t# Capture traffic from 8080 port\n\tgoreplay-udp --input-http :8080 --output-stdout")
	flag.Var(&Settings.outputHttp, "output-http", "Forwards incoming requests to given http address.\n\t# Redirect all incoming requests to staging.com address \n\tgor --input-raw :80 --output-http http://staging.com")

	/* outputHTTPConfig */
	//flag.Var(&Settings.outputHttpConfig.BufferSize, "output-http-response-buffer", "HTTP response buffer size, all data after this size will be discarded.")
	flag.IntVar(&Settings.outputHttpConfig.WorkersMin, "output-http-workers-min", 0, "Gor uses dynamic worker scaling. Enter a number to set a minimum number of workers. default = 1.")
	flag.IntVar(&Settings.outputHttpConfig.WorkersMax, "output-http-workers", 0, "Gor uses dynamic worker scaling. Enter a number to set a maximum number of workers. default = 0 = unlimited.")
	flag.IntVar(&Settings.outputHttpConfig.QueueLen, "output-http-queue-len", 1000, "Number of requests that can be queued for output, if all workers are busy. default = 1000")
	flag.BoolVar(&Settings.outputHttpConfig.SkipVerify, "output-http-skip-verify", false, "Don't verify hostname on TLS secure connection.")
	flag.DurationVar(&Settings.outputHttpConfig.WorkerTimeout, "output-http-worker-timeout", 2*time.Second, "Duration to rollback idle workers.")

	flag.IntVar(&Settings.outputHttpConfig.RedirectLimit, "output-http-redirects", 0, "Enable how often redirects should be followed.")
	flag.DurationVar(&Settings.outputHttpConfig.Timeout, "output-http-timeout", 5*time.Second, "Specify HTTP request/response timeout. By default 5s. Example: --output-http-timeout 30s")
	flag.BoolVar(&Settings.outputHttpConfig.TrackResponses, "output-http-track-response", false, "If turned on, HTTP output responses will be set to all outputs like stdout, file and etc.")

	flag.BoolVar(&Settings.outputHttpConfig.Stats, "output-http-stats", false, "Report http output queue stats to console every N milliseconds. See output-http-stats-ms")
	flag.IntVar(&Settings.outputHttpConfig.StatsMs, "output-http-stats-ms", 5000, "Report http output queue stats to console every N milliseconds. default: 5000")
	flag.BoolVar(&Settings.outputHttpConfig.OriginalHost, "http-original-host", false, "Normally gor replaces the Host http header with the host supplied with --output-http.  This option disables that behavior, preserving the original Host header.")
	/* outputHTTPConfig */
}
