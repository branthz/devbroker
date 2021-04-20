package diskqueue

import (
	"testing"

	"github.com/branthz/utarrow/lib/log"
	"github.com/stretchr/testify/assert"
)

func newQueue() DB {
	log.Setup("", "debug")
	var topic = "hello"
	db := New("/tmp/test", topic)
	return db
}

var msg = []string{
	"nihao1",
	"nihao2",
	"nihao3",
}

func TestTopicWrite(t *testing.T) {
	q := newQueue()
	for _, v := range msg {
		err := q.Write([]byte(v))
		if err != nil {
			t.Fatal(err)
		}
	}
	err := q.Close()
	if err != nil {
		t.Fatal(err)
	}
	//err = q.Empty()
	//if err != nil {
	//	t.Fatal(err)
	//}
}

func TestTopicRead(t *testing.T) {
	q := newQueue()
	dt := q.ReadMsg()
	assert.Equal(t, string(dt), msg[0])
	err := q.ReadCommit()
	if err != nil {
		t.Fatal(err)
	}
}
