package service

type msgIndex struct {
	ids   map[uint16]uint64
	point uint16
}

func newMsgroute() *msgIndex {
	m := new(msgIndex)
	return m
}

func (m *msgIndex) getroute(id uint16) uint64 {
	return m.ids[id]
}

func (m *msgIndex) getid() uint16 {
	m.point++
	return m.point
}

func (m *msgIndex) newroute(path uint64) {
	m.point++
	m.ids[m.point] = path
	return
}
