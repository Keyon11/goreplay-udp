package output

import (
	"github.com/myzhan/goreplay-udp/client"
	"github.com/myzhan/goreplay-udp/proto"
	"github.com/myzhan/goreplay-udp/stats"
	"sync/atomic"
	"time"
)

const initialDynamicWorkers = 10

type UDPOutputConfig struct {
	Workers        int
	Timeout        time.Duration
	Stats          bool
	IgnoreResponse bool
}

type UDPOutPut struct {
	// Keep this as first element of struct because it guarantees 64bit
	// alignment. atomic.* functions crash on 32bit machines if operand is not
	// aligned at 64bit. See https://github.com/golang/go/issues/599
	activeWorkers int64

	needWorker chan int

	address   string
	queue     chan *proto.Message
	responses chan *proto.Response

	config     *UDPOutputConfig
	queueStats *stats.GorStat
}

func NewUDPOutput(address string, config *UDPOutputConfig) (o *UDPOutPut) {
	o = new(UDPOutPut)
	o.address = address
	o.config = config

	if o.config.Stats {
		o.queueStats = stats.NewGorStat("output_udp")
	}

	o.queue = make(chan *proto.Message, 10000)
	o.needWorker = make(chan int, 1)

	if !o.config.IgnoreResponse {
		o.responses = make(chan *proto.Response, 10000)
	}

	// Initial workers count
	if o.config.Workers == 0 {
		o.needWorker <- initialDynamicWorkers
	} else {
		o.needWorker <- o.config.Workers
	}

	go o.workerMaster()
	return o
}

func (o *UDPOutPut) workerMaster() {
	for {
		newWorkers := <-o.needWorker
		for i := 0; i < newWorkers; i++ {
			go o.startWorker()
		}

		// Disable dynamic scaling if workers poll fixed size
		if o.config.Workers != 0 {
			return
		}
	}
}

func (o *UDPOutPut) startWorker() {
	c := client.NewUDPClient(o.address, o.config.Timeout, o.config.IgnoreResponse)
	deathCount := 0
	atomic.AddInt64(&o.activeWorkers, 1)
	for {
		select {
		case data := <-o.queue:
			o.sendRequest(c, data)
			deathCount = 0
		case <-time.After(time.Millisecond * 100):
			// When dynamic scaling enabled workers die after 2s of inactivity
			if o.config.Workers == 0 {
				deathCount++
			} else {
				continue
			}

			if deathCount > 20 {
				workersCount := atomic.LoadInt64(&o.activeWorkers)

				// At least 1 startWorker should be alive
				if workersCount != 1 {
					atomic.AddInt64(&o.activeWorkers, -1)
					return
				}
			}
		}
	}
}

func (o *UDPOutPut) PluginWrite(msg *proto.Message) (n int, err error) {
	if !proto.IsRequestPayload(msg.Meta) {
		return len(msg.Data), nil
	}

	o.queue <- msg

	if o.config.Stats {
		o.queueStats.Write(len(o.queue))
	}

	if o.config.Workers == 0 {
		workersCount := atomic.LoadInt64(&o.activeWorkers)

		if len(o.queue) > int(workersCount) {
			o.needWorker <- len(o.queue)
		}
	}

	return len(msg.Data) + len(msg.Meta), nil
}

// PluginRead reads message from this plugin
func (o *UDPOutPut) PluginRead() (*proto.Message, error) {
	if o.config.IgnoreResponse {
		return nil, ErrorStopped
	}
	var resp *proto.Response
	var msg proto.Message
	resp = <-o.responses
	msg.Data = resp.Payload

	msg.Meta = proto.PayloadHeader(proto.ReplayedResponsePayload, resp.Uuid, resp.RoundTripTime, nil)

	return &msg, nil
}

func (o *UDPOutPut) sendRequest(client *client.UDPClient, msg *proto.Message) {
	if !proto.IsRequestPayload(msg.Meta) {
		return
	}

	//TODO: should send meta?
	client.Send(msg.Data)
}

func (o *UDPOutPut) String() string {
	return "UDP output: " + o.address
}
