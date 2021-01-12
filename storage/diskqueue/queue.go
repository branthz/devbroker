package diskqueue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/branthz/utarrow/lib/log"
)

type dqueue struct {
	fdes       fileSeg
	exitChan   chan int
	exitWait   chan int
	writeChan  chan []byte
	writeResp  chan error
	writePos   int64
	readChan   chan []byte
	itemCounts int64
	sync.RWMutex
	exitFlag int
}

//Empty 清空该消息队列
//通过重置meta文件
func (d *dqueue) Empty() error {
	return nil
}

//Delete 清除topic相关的一切环境
func (d *dqueue) Delete() error { return nil }

//Depth
func (d *dqueue) Depth() int64 { return 0 }

func (d *dqueue) ReadChan() <-chan []byte {
	return d.readChan
}

func (d *dqueue) writeOne(data []byte) error {
	return d.fdes.writer.write(data)
}

//移动reader标记，清除历史文件
func (d *dqueue) moveForward() error {
	d.itemCounts--
	if err := d.fdes.reader.walkfile(); err != nil {
		log.Errorln(err)
	}
	return nil
}

//为了实现对存储管理的并发操作，通过后台线程ioLoop统一管理
func (d *dqueue) run() {
	var dataItem []byte
	var r chan []byte
	var err error
	syncTicker := time.NewTicker(2 * 1e9)
	for {
		if d.fdes.needSync {
			d.fdes.sync()
		}
		if d.fdes.checkReadAble() && d.fdes.reader.match() {
			dataItem, err = d.fdes.reader.readOne()
			if err != nil {
				log.Errorln(err)
				time.Sleep(1e9)
				continue
			}
			r = d.readChan
		} else {
			r = nil
		}
		select {
		case recv := <-d.writeChan:
			d.writeResp <- d.writeOne(recv)
		case r <- dataItem:
			log.Debugln("get in read channel")
			d.moveForward()
		case <-syncTicker.C:
			//log.Info("ioloop ticker hanppend")
		case <-d.exitChan:
			goto exit
		}
	}
exit:
	syncTicker.Stop()
	d.exitWait <- 1
}

func (d *dqueue) exit() {
	d.exitFlag = 1
	d.exitChan <- 1
	<-d.exitWait
	close(d.exitChan)
	close(d.exitWait)
}

func (d *dqueue) Write(data []byte) error {
	d.RLock()
	defer d.RUnlock()
	if d.exitFlag == 1 {
		return errors.New("exiting")
	}
	d.writeChan <- data
	return <-d.writeResp
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

//New 构造一个实例
func New(path, name string) DB {
	fd := newfileSeg(path, name)
	d := dqueue{
		fdes:      *fd,
		exitChan:  make(chan int, 0),
		exitWait:  make(chan int, 0),
		writeChan: make(chan []byte),
		writeResp: make(chan error),
		readChan:  make(chan []byte),
	}
	err := d.fdes.load()
	if err != nil {
		fmt.Println(err)
	}
	log.Info("queue:%s", d.fdes.String())
	go d.run()
	return &d
}
