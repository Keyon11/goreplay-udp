package proto

import (
	"bytes"
	"net"
	"strconv"
)

const (
	RequestPayload          = '1'
	ResponsePayload         = '2'
	ReplayedResponsePayload = '3'
)

var PayloadSeparator = "\n🐵🙈🙉\n"

func PayloadHeader(payloadType byte, uuid []byte, timing int64, srcIp []byte) (header []byte) {
	var sTime string
	var sIp string

	sTime = strconv.FormatInt(timing, 10)
	if len(srcIp) == 4 {
		sIp = net.IPv4(srcIp[0], srcIp[1], srcIp[2], srcIp[3]).String()
	} else if len(srcIp) > 0 {
		sIp = net.IP(srcIp).String()
	}

	//Example:
	// 3 f45590522cd1838b4a0d5c5aab80b77929dea3b3 1231 192.168.1.102\n
	// `+ 1` indicates space characters or end of line
	headerLen := 1 + 1 + len(uuid) + 1 + len(sTime) + 1 + len(sIp) + 1

	header = make([]byte, headerLen)
	header[0] = payloadType
	header[1] = ' '
	header[2+len(uuid)] = ' '
	header[3+len(uuid)+len(sTime)] = ' '
	header[len(header)-1] = '\n'

	copy(header[2:], uuid)
	copy(header[3+len(uuid):], sTime)
	copy(header[4+len(uuid)+len(sTime):], sIp)

	return header
}

func PayloadBody(payload []byte) []byte {
	headerSize := bytes.IndexByte(payload, '\n')
	return payload[headerSize+1:]
}

func PayloadMeta(payload []byte) [][]byte {
	headerSize := bytes.IndexByte(payload, '\n')
	if headerSize < 0 {
		headerSize = 0
	}
	return bytes.Split(payload[:headerSize], []byte{' '})
}

func PayloadMetaWithBody(payload []byte) (meta, body []byte) {
	if i := bytes.IndexByte(payload, '\n'); i > 0 && len(payload) > i+1 {
		meta = payload[:i+1]
		body = payload[i+1:]
		return
	}
	// we assume the message did not have meta data
	return nil, payload
}

func IsRequestPayload(payload []byte) bool {
	return payload[0] == RequestPayload
}
