package main

import (
	"github.com/myzhan/goreplay-udp/input"
	"github.com/myzhan/goreplay-udp/output"
	"github.com/myzhan/goreplay-udp/proto"
	"reflect"
	"strings"
	"sync"
)

// PluginReader is an interface for input plugins
type PluginReader interface {
	PluginRead() (msg *proto.Message, err error)
}

// PluginWriter is an interface for output plugins
type PluginWriter interface {
	PluginWrite(msg *proto.Message) (n int, err error)
}

// PluginReadWriter is an interface for plugins that support reading and writing
type PluginReadWriter interface {
	PluginReader
	PluginWriter
}

// InOutPlugins struct for holding references to plugins
type InOutPlugins struct {
	Inputs  []PluginReader
	Outputs []PluginWriter
	All     []interface{}
}

var pluginMu sync.Mutex

// Plugins holds all the plugin objects
var Plugins = new(InOutPlugins)

// extractLimitOptions detects if plugin get called with limiter support
// Returns address and limit
func extractLimitOptions(options string) (string, string) {
	split := strings.Split(options, "|")

	if len(split) > 1 {
		return split[0], split[1]
	}

	return split[0], ""
}

// Automatically detects type of plugin and initialize it
//
// See this article if curious about reflect stuff below: http://blog.burntsushi.net/type-parametric-functions-golang
func registerPlugin(constructor interface{}, options ...interface{}) {
	var path, limit string
	vc := reflect.ValueOf(constructor)

	// Pre-processing options to make it work with reflect
	vo := []reflect.Value{}
	for _, oi := range options {
		vo = append(vo, reflect.ValueOf(oi))
	}

	if len(vo) > 0 {
		// Removing limit options from path
		path, limit = extractLimitOptions(vo[0].String())

		// Writing value back without limiter "|" options
		vo[0] = reflect.ValueOf(path)
	}

	// Calling our constructor with list of given options
	plugin := vc.Call(vo)[0].Interface()

	if limit != "" {
		plugin = NewLimiter(plugin, limit)
	}

	_, isR := plugin.(PluginReader)
	_, isW := plugin.(PluginWriter)

	// Some of the output can be Readers as well because return responses
	if isR && !isW {
		Plugins.Inputs = append(Plugins.Inputs, plugin.(PluginReader))
	}

	if isW {
		Plugins.Outputs = append(Plugins.Outputs, plugin.(PluginWriter))
	}

	Plugins.All = append(Plugins.All, plugin)
}

// InitPlugins specify and initialize all available plugins
func InitPlugins() {
	pluginMu.Lock()
	defer pluginMu.Unlock()

	if Settings.outputStdout {
		registerPlugin(output.NewStdOutput)
	}

	if Settings.outputNull {
		registerPlugin(output.NewNullOutput)
	}

	for _, options := range Settings.inputUDP {
		registerPlugin(input.NewUDPInput, options, Settings.inputUDPTrackResponse)
	}

	for _, options := range Settings.inputFile {
		registerPlugin(input.NewFileInput, options, Settings.inputFileLoop)
	}

	for _, options := range Settings.outputFile {
		registerPlugin(output.NewFileOutput, options, &Settings.outputFileConfig)
	}

	for _, options := range Settings.outputUDP {
		registerPlugin(output.NewUDPOutput, options, &Settings.outputUDPConfig)
	}

	for _, options := range Settings.inputHttp {
		registerPlugin(input.NewHTTPInput, options)
	}

	for _, options := range Settings.outputHttp {
		registerPlugin(output.NewHTTPOutput, options, &Settings.outputHttpConfig)
	}
}
