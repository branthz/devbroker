package diskqueue

import (
	"bytes"
	"encoding/binary"
	"os"
)

type persist struct {
	wfileSeq    int64
	writeBuffer bytes.Buffer
	writeFile   *os.File
	wPosition   int64
	parent      *fileSeg
}

func (p *persist) Close() {
	if p.writeFile != nil {
		p.writeFile.Close()
		p.writeFile = nil
	}
}

func (p *persist) write(data []byte) error {
	var err error
	if p.writeFile == nil {
		if p.writeFile, err = os.OpenFile(p.parent.fileName(p.wfileSeq), os.O_RDWR|os.O_CREATE, 0600); err != nil {
			return err
		}
		if p.wPosition > 0 {
			_, err = p.writeFile.Seek(p.wPosition, 0)
			if err != nil {
				p.Close()
				return err
			}
		}
	}
	dataLen := int32(len(data))
	p.writeBuffer.Reset()
	err = binary.Write(&p.writeBuffer, binary.BigEndian, dataLen)
	if err != nil {
		return err
	}
	_, err = p.writeBuffer.Write(data)
	if err != nil {
		p.Close()
		return err
	}
	_, err = p.writeFile.Write(p.writeBuffer.Bytes())
	if err != nil {
		p.Close()
		return err
	}

	p.wPosition = p.wPosition + 4 + int64(dataLen)
	//split file
	if p.wPosition > p.parent.segSize {
		p.Close()
		p.wfileSeq++
		p.wPosition = 0
		p.parent.sync()
	}
	return err
}
