package input

import (
	"github.com/myzhan/goreplay-udp/listener"
	"github.com/myzhan/goreplay-udp/proto"
	"log"
	"net"
)

type UDPInput struct {
	data          chan *proto.UDPMessage
	address       string
	quit          chan bool
	listener      *listener.UDPListener
	trackResponse bool
}

func NewUDPInput(address string, trackResponse bool) (i *UDPInput) {
	i = new(UDPInput)
	i.data = make(chan *proto.UDPMessage)
	i.address = address
	i.quit = make(chan bool)
	i.trackResponse = trackResponse
	i.listen(address)
	return
}

func (i *UDPInput) PluginRead() (*proto.Message, error) {
	var msg proto.Message
	msgUdp := <-i.data
	msg.Data = msgUdp.Data()

	if msgUdp.IsIncoming {
		msg.Meta = proto.PayloadHeader(proto.RequestPayload, msgUdp.UUID(), msgUdp.Start.UnixNano(), msgUdp.SrcIp)
	} else {
		msg.Meta = proto.PayloadHeader(proto.ResponsePayload, msgUdp.UUID(), msgUdp.Start.UnixNano(), msgUdp.SrcIp)
	}
	msgUdp = nil
	return &msg, nil
}

func (i *UDPInput) listen(address string) {
	log.Println("Listening for traffic on: " + address)

	host, port, err := net.SplitHostPort(address)

	if err != nil {
		log.Fatal("input-raw: error while parsing address", err)
	}

	i.listener = listener.NewUDPListener(host, port, i.trackResponse)

	ch := i.listener.Receiver()

	go func() {
		for {
			select {
			case <-i.quit:
				return
			default:
			}
			// Receiving UDPMessage
			m := <-ch
			i.data <- m
		}
	}()
}
