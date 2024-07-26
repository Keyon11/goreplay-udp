package proto

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestPayload(t *testing.T) {
	sIp := "192.168.1.102"
	ip := net.ParseIP(sIp)
	uid := uuid.New()
	st := time.Now().UnixNano()
	meta := PayloadHeader(RequestPayload, []byte(uid.String()), st, ip)
	t.Logf("meta:%s", string(meta))

	es := PayloadMeta(meta)
	assert.Equal(t, 4, len(es))
	assert.Equal(t, byte(RequestPayload), es[0][0])
	assert.Equal(t, uid.String(), string(es[1]))
	assert.Equal(t, strconv.Itoa(int(st)), string(es[2]))
	assert.Equal(t, sIp, string(es[3]))
}
