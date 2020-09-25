package topics

import (
	"time"

	"github.com/branthz/devbroker/storage"
)

type Subscriber interface {
	ID() string
	Send([]byte) error
}

type delivery struct {
	topic    string
	sub      Subscriber
	readAble bool
	ticker   time.Ticker
	store    storage.Storage
	readBuf  chan []byte
	exit     chan struct{}
}

func newDelivery(tp string, s Subscriber) *delivery {
	return &delivery{
		topic:   tp,
		sub:     s,
		readBuf: make(chan []byte, 0),
		ticker:  time.Ticker{},
		exit:    make(chan struct{}),
	}
}

func (d *delivery) stop() {
	d.exit <- struct{}{}
}

func (d *delivery) onStop() {
	d.ticker.Stop()
	close(d.exit)
}

//有消息入队
func (d *delivery) setRead() {
	d.readAble = true
}

//没有可消费的了
func (d *delivery) beSilent() {
	d.readAble = false
}

//获取数据
//TODO 增加流控
func (d *delivery) ready() chan []byte {
	if d.readAble == true {
		dt := d.store.ReadMsg(d.topic, 1)
		d.readBuf <- dt
		return d.readBuf
	}
	return nil
}

func (d *delivery) start() {
	go func() {
		defer d.onStop()
		for {
			select {
			case dt := <-d.ready():
				d.sub.Send(dt)
			case <-d.exit:
				return
			}
			time.Sleep(1e9)
		}
	}()
	return
}

//这种定义只支持单一消费者模式，key为主题,想要广播效果就是多个消费者，可将key扩展成主题+频道
type Workq struct {
	cn map[string]*delivery
}

//TODO
//解决重复添加造成goroutine泄漏
//频道的创建由单独的管理入口不放在订阅事件里
func (s *Workq) AddSub(topic string, con Subscriber, st storage.Storage) {
	d := newDelivery(topic, con)
	d.store = st
	s.cn[topic] = d
	d.start()
}

func (s *Workq) UnSub(topic string, con Subscriber) {
	d := s.cn[topic]
	d.stop()
	delete(s.cn, topic)
}

//每个topic单独一个管理goroutine
//针对每个消费者，还是设置一个快递小哥比较合理，集中式的loop里发货效率比较低
//每个快递小哥周期确认是否要送货。
func (s *Workq) hello() {
	return
}
