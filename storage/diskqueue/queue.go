package diskqueue

import (
	"errors"
	"fmt"
	"sync"

	"github.com/branthz/utarrow/lib/log"
)

type dqueue struct {
	fdes     fileSeg
	exitChan chan int
	exitWait chan int
	writePos int64
	sync.RWMutex
	exitFlag   int
	itemCounts int64
}

//New 构造一个实例
func New(path, name string) DB {
	fd := newfileSeg(path, name)
	d := dqueue{
		fdes:     *fd,
		exitChan: make(chan int, 0),
		exitWait: make(chan int, 0),
	}
	//文件元信息加载
	err := d.fdes.load()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Info("queue:%s", d.fdes.String())
	return &d
}

//Delete 清除topic相关的一切环境
func (d *dqueue) Delete() error {
	return nil
}

//Empty 清空该消息队列
//通过重置meta文件
func (d *dqueue) Empty() error {
	return nil
}

func (d *dqueue) ReadAble() bool {
	return d.fdes.checkReadAble()
}

func (d *dqueue) ReadMsg() []byte {
	data, err := d.fdes.reader.readOne()
	if err != nil {
		log.Errorln(err)
		return nil
	}
	return data
}

func (d *dqueue) ReadCommit() error {
	d.itemCounts--
	return d.fdes.reader.walkfile()
}
func (d *dqueue) Write(data []byte) error {
	d.RLock()
	defer d.RUnlock()
	if d.exitFlag == 1 {
		return errors.New("write request during existing")
	}
	return d.fdes.writer.write(data)
}

//queue object close
func (d *dqueue) Close() error {
	//d.RLock()
	//defer d.RUnlock()
	log.Warnln("queue close")
	d.exit()
	d.fdes.sync()
	d.fdes.Shutdown()
	return nil
}

func (d *dqueue) exit() {
	d.exitFlag = 1
	d.exitChan <- 1
	<-d.exitWait
	close(d.exitChan)
	close(d.exitWait)
}
