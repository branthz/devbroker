//消息的持久化存储，消息包含：topic/data/id
//消息的消费包括read和清理；read需要标记读到哪里，清理就像日志压缩。
//对于qos1根据commitIndex readmsg

package storage

import (
	"container/list"
)

//读取数据由外部控制，可以做流控，外部需要一个后台loop检测是否可读
//consumer groutine维护consumer连接，如果有消息发送下去，同步等待返回结果。
//元信息的管理放入引擎内部
type Storage interface {
	SaveMsg(topic string, data []byte) error
	ReadMsg(topic string, batch int) []byte
	CommitRead(topic string, offset uint64) error
	PreSaveMsg(pid []byte, data []byte) error
	CommitMsg(pid []byte) error
}

// --------------------------
// topic/{filename} | msg1/msg2/msg3/...
// filename=fid.data;fid和offset需要保存在元信息里，用来定位到写入位置。
// 元信息可以放在外部，当参数传入进去;这里不需要考虑操作原子性,因为不回滚不造成数据错乱。
// --------------------------
// 消息读取传入index
// 消息确认，提供index;消息index怎么和文件id和offset关联起来？uint32+uint32
// ---------------------
type Noop struct {
	//topic---msgs
	data map[string]*list.List
}

func NewNoop() *Noop {
	n := &Noop{}
	n.data = make(map[string]*list.List)
	return n
}

func (n *Noop) SaveMsg(topic string, data []byte) error {
	if v, ok := n.data[topic]; ok {
		v.PushBack(data)
	} else {
		l := list.New()
		l.PushBack(data)
		n.data[topic] = l
	}
	return nil
}

func (n *Noop) ReadMsg(topic string, batch int) []byte {
	if v, ok := n.data[topic]; ok {
		dt := v.Front()
		//还没有实现批量逻辑
		if dt != nil {
			return dt.Value.([]byte)
		}
	}
	return nil
}

func (n *Noop) CommitRead(topic string) error {
	return nil
}

func (n *Noop) PreSaveMsg(pid []byte, data []byte) error {
	return nil
}

func (n *Noop) CommitMsg(pid []byte) error {
	return nil
}

//sever服务器主动下推,一次取出一条等待client段响应结果；成功后再处理下一条------->吞吐量太低了些。
//一次取出若干条,支持批量传输.
//
