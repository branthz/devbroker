//消息的持久化存储，消息包含：topic/data/id
//消息的消费包括read和清理；read需要标记读到哪里，清理就像日志压缩。
//对于qos1根据commitIndex readmsg

package storage

import (
	"github.com/branthz/devbroker/message"
)

type earth interface {
	SaveMsg(topic string, msg []byte) error
	ReadMsg(topic string, batch int) []byte
	CommitMsg(topic string, index uint32)
}

type Storage struct {
}

func (s *Storage) Store(m *message.Message) {

}

// --------------------------
// | commit |msg1/msg2/msg3/...
// --------------------------
//

func ioLoop() {
	//copy msg1 ->
	//copy msg2
}

//sever服务器主动下推,一次取出一条等待client段响应结果；成功后再处理下一条------->吞吐量太低了些。
//一次取出若干条,支持批量传输.
//
func readmsg() {
	//根据commit信息读取下一条
}
