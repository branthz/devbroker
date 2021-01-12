package diskqueue

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path"

	"github.com/branthz/utarrow/lib/log"
)

type fileSeg struct {
	path       string
	name       string
	segSize    int64
	reader     *query
	writer     *persist
	needSync   bool //是否已经刷盘
	itemCounts int64
}

func newfileSeg(path, name string) *fileSeg {
	f := new(fileSeg)
	f.name = name
	f.path = path
	f.segSize = 1 << 24
	f.reader = new(query)
	f.writer = new(persist)
	f.reader.parent = f
	f.writer.parent = f
	return f
}

type persist struct {
	wfileSeq    int64 //文件序列号
	writeBuffer bytes.Buffer
	writeFile   *os.File //filer
	wPosition   int64
	parent      *fileSeg
}

func (p *persist) Close() {
	if p.writeFile != nil {
		p.writeFile.Close()
		p.writeFile = nil
	}
}

//元信息存储路由
func (f *fileSeg) metaDataFileName() string {
	return fmt.Sprintf(path.Join(f.path, "%s.meta.dat"), f.name)
}

//文件实体路由
func (f *fileSeg) fileName(seq int64) string {
	return fmt.Sprintf(path.Join(f.path, "%s.diskqueue.%06d.dat"), f.name, seq)
}

func (f *fileSeg) String() string {
	var desc = fmt.Sprintf("the running file:%s,write index:[%d,%d],read index:[%d,%d]", f.metaDataFileName(), f.writer.wfileSeq, f.writer.wPosition, f.reader.rfileSeq, f.reader.rPosition)
	return desc
}

//启动时加载元信息
func (f *fileSeg) load() error {
	name := f.metaDataFileName()
	fs, err := os.OpenFile(name, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer fs.Close()
	var counts int64
	_, err = fmt.Fscanf(fs, "%d\n%d,%d\n%d,%d\n",
		&counts,
		&f.writer.wfileSeq, &f.writer.wPosition,
		&f.reader.rfileSeq, &f.reader.rPosition)
	if err != nil {
		return err
	}
	f.itemCounts = counts
	return nil
}

//文件对象关闭
func (f *fileSeg) Shutdown() {
	f.writer.Close()
	f.reader.Close()
}

//文件是否可读
func (f *fileSeg) checkReadAble() bool {
	if f.reader.rfileSeq < f.writer.wfileSeq || f.reader.rPosition < f.writer.wPosition {
		return true
	}
	return false
}

//数据写入
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

//元数据保存
func (f *fileSeg) saveMeta() error {
	tmpfile := fmt.Sprintf("%s/%s.%d.tmp", f.path, f.name, rand.Int())
	filename := f.metaDataFileName()
	fs, err := os.OpenFile(tmpfile, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(fs, "%d\n%d,%d\n%d,%d\n", f.itemCounts, f.writer.wfileSeq, f.writer.wPosition, f.reader.rfileSeq, f.reader.rPosition)
	if err != nil {
		fs.Close()
		return err
	}
	fs.Sync()
	fs.Close()
	return os.Rename(tmpfile, filename)
}

//文件同步磁盘,包含消息日志和元数据
func (f *fileSeg) sync() {
	if f.writer.writeFile != nil {
		err := f.writer.writeFile.Sync()
		if err != nil {
			f.writer.Close()
			log.Error("file sync failed:%v", err)
			return
		}
	}
	log.Debugln("fileseg sync")
	if err := f.saveMeta(); err != nil {
		log.Error("file save meta failed:%v", err)
		return
	}
	f.needSync = false
	return
}
