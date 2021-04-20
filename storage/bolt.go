package storage

import (
	"fmt"
	"log"

	"github.com/branthz/devbroker/storage/diskqueue"
	bolt "github.com/etcd-io/bbolt"
)

const presave = "qos2"

type Boltdb struct {
	db   *bolt.DB
	ques map[string]diskqueue.DB
}

func NewBolt(path string) *Boltdb {
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
	d.ques = make(map[string]diskqueue.DB)
	q := diskqueue.New(path, "hangzhou")
	d.ques["hangzhou"] = q
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

func (b *Boltdb) SaveMsg(topic string, data []byte) error {
	err := b.ques[topic].Write(data)
	return err
}

func (b *Boltdb) ReadMsg(topic string, batch int) []byte {
	dt := b.ques[topic].ReadMsg()
	return dt
}

func (b *Boltdb) CommitRead(topic string, offset uint64) error {
	b.ques[topic].ReadCommit()
	return nil
}
