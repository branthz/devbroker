package diskqueue

type DB interface {
	Close() error
	Write(data []byte) error
	Delete() error
	Empty() error
	ReadChan() <-chan []byte
}

//至少一次
//after-read,wait for commit(ack),if no ack next time send from commit index