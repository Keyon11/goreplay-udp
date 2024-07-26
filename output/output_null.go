package output

import "github.com/myzhan/goreplay-udp/proto"

// NullOutput used for debugging, prints nothing
type NullOutput struct {
}

// NullOutput constructor for NullOutput
func NewNullOutput() (o *NullOutput) {
	return new(NullOutput)
}

func (o *NullOutput) PluginWrite(msg *proto.Message) (int, error) {
	return len(msg.Meta) + len(msg.Data) + len(proto.PayloadSeparator), nil
}

func (o *NullOutput) String() string {
	return "Null Output"
}
