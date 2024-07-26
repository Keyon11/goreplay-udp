package output

import (
	"github.com/myzhan/goreplay-udp/proto"
	"os"
)

// StdOutput used for debugging, prints all incoming requests
type StdOutput struct {
}

// NewStdOutput constructor for StdOutput
func NewStdOutput() (i *StdOutput) {
	i = new(StdOutput)
	return
}

func (i *StdOutput) PluginWrite(msg *proto.Message) (int, error) {
	var n, nn int
	var err error
	n, err = os.Stdout.Write(msg.Meta)
	nn, err = os.Stdout.Write(msg.Data)
	n += nn
	nn, err = os.Stdout.Write([]byte(proto.PayloadSeparator))
	n += nn

	return n, err
}

func (i *StdOutput) String() string {
	return "Stdout Output"
}
