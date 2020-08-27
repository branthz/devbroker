//对于qos1
//根据commitIndex readmsg

package storage

type earth interface {
	SaveMsg(topic string, msg []byte) error
	ReadMsg(topic string, batch int) []byte
	CommitMsg(topic string, index uint32)
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
