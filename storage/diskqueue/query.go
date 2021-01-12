package diskqueue

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type query struct {
	nextRPosition int64
	nextRSeq      int64
	rPosition     int64
	rfileSeq      int64 //文件序列号
	parent        *fileSeg
	readFile      *os.File
	reader        *bufio.Reader
}

func (q *query) Close() {
	if q.readFile != nil {
		q.readFile.Close()
		q.readFile = nil
	}
}

func (q *query) match() bool {
	return q.rPosition == q.nextRPosition
}

func (q *query) walkfile() (err error) {
	oldfSeq := q.rfileSeq
	q.rfileSeq = q.nextRSeq
	q.rPosition = q.nextRPosition
	//jump to next file
	if oldfSeq != q.nextRSeq {
		q.parent.needSync = true
		name := q.parent.fileName(oldfSeq)
		err = os.Remove(name)
	}
	return
}

func (q *query) readOne() (data []byte, err error) {
	if q.readFile == nil {
		fname := q.parent.fileName(q.rfileSeq)
		q.readFile, err = os.OpenFile(fname, os.O_RDONLY, 0600)
		if err != nil {
			return
		}
		if q.rPosition > 0 {
			_, err = q.readFile.Seek(q.rPosition, 0)
			if err != nil {
				q.Close()
				return
			}
		}
		q.reader = bufio.NewReader(q.readFile)
	}
	//defer fs.Close()
	var rsize int32
	err = binary.Read(q.reader, binary.BigEndian, &rsize)
	if err != nil {
		return
	}
	//TODO
	//read size check
	if rsize < 1 || rsize > 655350 {
		err = fmt.Errorf("invalid message read size (%d)", rsize)
		q.Close()
		return
	}
	data = make([]byte, rsize)
	_, err = io.ReadFull(q.reader, data)
	if err != nil {
		q.Close()
		return
	}
	q.nextRPosition = q.rPosition + int64(rsize) + 4
	q.nextRSeq = q.rfileSeq
	if q.nextRPosition > q.parent.segSize {
		q.Close()
		q.nextRSeq++
		q.nextRPosition = 0
	}
	return
}
