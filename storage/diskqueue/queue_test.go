package diskqueue

import (
	"testing"

	"github.com/branthz/utarrow/lib/log"
)

const sample = "today is good,tomorrow will be better4"

func init() {
	log.Setup("", "debug")
}

func TestWrite(t *testing.T) {
	db := New("./", "hangzhou")
	err := db.Write([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func BenchmarkWriteB(b *testing.B) {
	db := New("./", "lanzhou")
	log.Debugln("benchmark write start")
	for i := 0; i < b.N; i++ {
		err := db.Write([]byte(sample))
		if err != nil {
			b.Fatal(err)
		}
	}
	db.Close()
	log.Debugln("benchmark write stop")
}

func TestRead(t *testing.T) {
	db := New("./", "hangzhou")
	defer db.Close()
	ch := db.ReadChan()
	var hole []byte
	var broke int
	for v := range ch {
		hole = v
		broke++
		if broke > 3 {
			break
		}
	}
	t.Logf("read:%d,last data:%s", broke, string(hole))
}
