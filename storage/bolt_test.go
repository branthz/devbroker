package storage

import (
	"testing"

	"github.com/branthz/utarrow/lib/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Setup("", "debug")
}
func TestBolt(t *testing.T) {
	b := NewBolt()
	var topic = "hangzhou"
	var msg = []byte("just do it")
	err := b.SaveMsg(topic, msg)
	if err != nil {
		t.Fatal(err)
	}
	dt := b.ReadMsg(topic, 1)
	assert.Equal(t, dt, msg)
}
