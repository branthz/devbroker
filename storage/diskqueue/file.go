package diskqueue

import (
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
	writer     *persist
	reader     *query
	needSync   bool
	itemCounts int64
}

func newfileSeg(path, name string) *fileSeg {
	f := new(fileSeg)
	f.name = name
	f.path = path
	f.segSize = 1 << 24
	f.writer = new(persist)
	f.reader = new(query)
	f.writer.parent = f
	f.reader.parent = f
	return f
}

//元信息存储路由
func (f *fileSeg) metaDataFileName() string {
	return fmt.Sprintf(path.Join(f.path, "%s.meta.dat"), f.name)
}

//文件实体路由
func (f *fileSeg) fileName(seq int64) string {
	return fmt.Sprintf(path.Join(f.path, "%s.diskqueue.%06d.dat"), f.name, seq)
}

//启动时加载元信息
func (f *fileSeg) load() error {
	name := f.metaDataFileName()
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return nil
	}
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

//内容文件对象关闭
func (f *fileSeg) Shutdown() {
	f.writer.Close()
	f.reader.Close()
}

func (f *fileSeg) String() string {
	var desc = fmt.Sprintf("the running file:%s,write index:[%d,%d],read index:[%d,%d]", f.metaDataFileName(), f.writer.wfileSeq, f.writer.wPosition, f.reader.rfileSeq, f.reader.rPosition)
	return desc
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

func (f *fileSeg) sync() {
	if f.writer.writeFile != nil {
		err := f.writer.writeFile.Sync()
		if err != nil {
			f.writer.Close()
			log.Error("file sync failed:%v", err)
			return
		}
	}
	if err := f.saveMeta(); err != nil {
		log.Error("file save meta failed:%v", err)
		return
	}
	f.needSync = false
	return
}

//文件是否可读
func (f *fileSeg) checkReadAble() bool {
	if f.reader.rfileSeq < f.writer.wfileSeq || f.reader.rPosition < f.writer.wPosition {
		return true
	}
	return false
}
