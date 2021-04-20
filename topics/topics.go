package topics

import (
	"time"

	"github.com/branthz/devbroker/message"
	"github.com/branthz/devbroker/storage"
	"github.com/branthz/utarrow/lib/log"
)

type Subscriber interface {
	ID() string
	Send(*message.Message) error
}

//快递员
type delivery struct {
	topic   string
	sub     Subscriber
	window  int
	canRead bool
	ticker  time.Ticker
	store   storage.Storage
	readBuf chan []byte
	exit    chan struct{}
}

func newDelivery(tp string, s Subscriber) *delivery {
	return &delivery{
		topic:   tp,
		sub:     s,
		window:  3,
		canRead: true,
		readBuf: make(chan []byte, 1),
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
	d.canRead = true
}

//没有可消费的了
func (d *delivery) beSilent() {
	d.canRead = false
}

func (d *delivery) readAble() bool {
	if d.window > 0 && d.canRead {
		return true
	}
	return false
}

//订阅者消费携程
func (d *delivery) start() {
	go func() {
		var dataRead []byte
		defer d.onStop()
		for {
			select {
			case <-d.exit:
				return
			default:
				if d.readAble() {
					dataRead = d.store.ReadMsg(d.topic, 1)
					if dataRead == nil {
						d.beSilent()
					} else {
						log.Info("delivery get msg:%s", string(dataRead))
						m, _ := message.DecodeMessage(dataRead)
						d.sub.Send(m)
						if m.Qos == 0 {
							d.store.CommitRead(string(m.Topic), 10)
						}
					}
				}
			}
		}
	}()
}

//这种定义只支持单一消费者模式，key为主题,想要广播效果就是多个消费者，可将key扩展成主题+频道
type TopicPool struct {
	store storage.Storage
	cn    map[string]*delivery
}

var TopicHandler *TopicPool

func New(sto storage.Storage) *TopicPool {
	if TopicHandler == nil {
		TopicHandler = &TopicPool{cn: make(map[string]*delivery)}
		TopicHandler.store = sto
	}
	return TopicHandler
}

//TODO
//解决重复添加造成goroutine泄漏
//频道的创建由单独的管理入口不放在订阅事件里
func (s *TopicPool) AddSub(topic string, con Subscriber, st storage.Storage) {
	if v, ok := s.cn[topic]; ok {
		v.sub = con
	} else {
		d := newDelivery(topic, con)
		d.store = st
		s.cn[topic] = d
		d.setRead()
		d.start()
	}
	log.Info("client sub-----------:%s", topic)
}

func (s *TopicPool) UnSub(topic string, con Subscriber) {
	d := s.cn[topic]
	d.stop()
	delete(s.cn, topic)
}

//每个topic单独一个管理goroutine
//针对每个消费者，还是设置一个快递小哥比较合理，集中式的loop里发货效率比较低
//每个快递小哥周期确认是否要送货。
func (s *TopicPool) SaveMsg(topic string, data []byte) error {
	if v, ok := s.cn[topic]; ok {
		v.setRead()
	}
	return s.store.SaveMsg(topic, data)
}
