package proto

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"strconv"
	"time"
)

// Message represents data across plugins
type Message struct {
	Meta []byte // metadata
	Data []byte // actual data
}

type Response struct {
	Payload       []byte
	Uuid          []byte
	RoundTripTime int64
	StartedAt     int64
}

type UDPMessage struct {
	IsIncoming bool
	Start      time.Time
	SrcIp      []byte
	DstIp      []byte
	SrcPort    uint16
	DstPort    uint16
	length     uint16
	checksum   uint16
	data       []byte
}

func NewUDPMessage(data []byte, srcIp, dstIp []byte, isIncoming bool) (m *UDPMessage) {
	m = &UDPMessage{}
	udp := &layers.UDP{}
	err := udp.DecodeFromBytes(data, gopacket.NilDecodeFeedback)
	if err != nil {
		log.Printf("Error decode udp message, %v\n", err)
	}
	m.SrcIp = make([]byte, len(srcIp))
	copy(m.SrcIp, srcIp)
	m.DstIp = make([]byte, len(dstIp))
	copy(m.DstIp, dstIp)
	m.SrcPort = uint16(udp.SrcPort)
	m.DstPort = uint16(udp.DstPort)
	m.length = udp.Length
	m.checksum = udp.Checksum
	m.data = udp.Payload
	m.IsIncoming = isIncoming

	return
}

func (m *UDPMessage) UUID() []byte {
	var key []byte

	key = strconv.AppendInt(key, m.Start.UnixNano(), 10)
	key = strconv.AppendUint(key, uint64(m.SrcPort), 10)
	key = strconv.AppendUint(key, uint64(m.DstPort), 10)
	key = strconv.AppendUint(key, uint64(m.length), 10)

	uuid := make([]byte, 40)
	sha := sha1.Sum(key)
	hex.Encode(uuid, sha[:20])

	return uuid
}

func (m *UDPMessage) Data() []byte {
	return m.data
}

func (m *UDPMessage) String() string {
	return fmt.Sprintf("SrcPort: %d | DstPort: %d | Length: %d | Checksum: %d | Data: %s",
		m.SrcPort, m.DstPort, m.length, m.checksum, string(m.data))
}
