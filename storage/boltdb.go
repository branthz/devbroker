package storage

import (
	"fmt"
	"log"

	"github.com/branthz/devbroker/storage/diskqueue"
	bolt "github.com/etcd-io/bbolt"
)

const presave = "qos2"

type Boltdb struct {
	db  *bolt.DB
	que diskqueue.DB
}

func NewBolt() *Boltdb {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	d := new(Boltdb)
	db, err := bolt.Open("my.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(presave))
		if err != nil {
			return fmt.Errorf("create bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	d.db = db
	q := diskqueue.New("/tmp/data", "hangzhou")
	d.que = q
	return d
}

//pid = producerid+msgid
func (b *Boltdb) PreSaveMsg(pid []byte, data []byte) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(presave))
		err := b.Put(pid, data)
		return err
	})
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

func (b *Boltdb) CommitMsg(pid []byte) error {
	var data []byte
	b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(presave))
		data = b.Get(pid)
		return nil
	})
	if len(data) == 0 {
		//alrady commited
		return nil
	}
	topic := "hello" //parse from pid
	b.SaveMsg(topic, data)
	b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(presave)).Delete(pid)
	})
	return nil
}

func (n *Boltdb) SaveMsg(topic string, data []byte) error {
	err := n.que.Write(data)
	return err
}
func (n *Boltdb) ReadMsg(topic string, batch int) []byte {
	return nil
}

func (n *Boltdb) CommitRead(topic string, index uint64) error {
	return nil
}
