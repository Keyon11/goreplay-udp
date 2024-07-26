package input

import (
	"errors"
	"github.com/google/uuid"
	"github.com/myzhan/goreplay-udp/proto"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// ErrorStopped is the error returned when the go routines reading the input is stopped.
var ErrorStopped = errors.New("reading stopped")

// HTTPInput used for sending requests to Gor via http
type HTTPInput struct {
	data     chan *proto.Message
	address  string
	listener net.Listener
	stop     chan bool // Channel used only to indicate goroutine should shutdown
}

// NewHTTPInput constructor for HTTPInput. Accepts address with port which it will listen on.
func NewHTTPInput(address string) (i *HTTPInput) {
	i = new(HTTPInput)
	i.data = make(chan *proto.Message, 1000)
	i.stop = make(chan bool)

	i.listen(address)

	return
}

// PluginRead reads message from this plugin
func (i *HTTPInput) PluginRead() (*proto.Message, error) {
	select {
	case <-i.stop:
		return nil, ErrorStopped
	case msg := <-i.data:
		return msg, nil
	}
}

// Close closes this plugin
func (i *HTTPInput) Close() error {
	close(i.stop)
	return nil
}

func (i *HTTPInput) handler(w http.ResponseWriter, r *http.Request) {
	r.URL.Scheme = "http"
	r.URL.Host = i.address

	var msg proto.Message

	remoteIp := r.Header.Get("X-Real-IP")
	if len(remoteIp) == 0 {
		remoteIp = r.RemoteAddr
		if strings.Contains(remoteIp, ":") {
			ss := strings.Split(remoteIp, ":")
			remoteIp = ss[0]
		}
	}

	msg.Meta = proto.PayloadHeader(proto.RequestPayload, []byte(uuid.New().String()), time.Now().UnixNano(), []byte(net.ParseIP(remoteIp)))
	msg.Data, _ = io.ReadAll(r.Body)
	//buf, _ := httputil.DumpRequestOut(r, true)
	http.Error(w, http.StatusText(200), 200)
	i.data <- &msg
}

func (i *HTTPInput) listen(address string) {
	var err error

	mux := http.NewServeMux()

	mux.HandleFunc("/", i.handler)

	i.listener, err = net.Listen("tcp", address)
	if err != nil {
		log.Fatal("HTTP input listener failure:", err)
	}
	i.address = i.listener.Addr().String()

	go func() {
		err = http.Serve(i.listener, mux)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP input serve failure ", err)
		}
	}()
}

func (i *HTTPInput) String() string {
	return "HTTP input: " + i.address
}
