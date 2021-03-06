package mqtt

import (
	"errors"
	"fmt"
	"io"

	"github.com/branthz/utarrow/lib/log"
)

const (
	maxHeaderSize  = 6 // max MQTT header size
	maxMessageSize = 65535
)

//Message is the interface all our packets will be implementing
type Message interface {
	fmt.Stringer

	Type() uint8
	EncodeTo(w io.Writer) (int, error)
}

// Reader is the requied reader for an efficient decoding.
type Reader interface {
	io.Reader
	ReadByte() (byte, error)
}

// MQTT type
const (
	TypeOfConnect = uint8(iota + 1)
	TypeOfConnack
	TypeOfPublish
	TypeOfPuback
	TypeOfPubrec
	TypeOfPubrel
	TypeOfPubcomp
	TypeOfSubscribe
	TypeOfSuback
	TypeOfUnsubscribe
	TypeOfUnsuback
	TypeOfPingreq
	TypeOfPingresp
	TypeOfDisconnect
)

func msgDesc(tp uint8) string {
	var desc string
	switch tp {
	case TypeOfConnect:
		desc = "TypeOfConnect"
	case TypeOfConnack:
		desc = "TypeOfConnack"
	case TypeOfPublish:
		desc = "TypeOfPublish"
	case TypeOfPuback:
		desc = "TypeOfPuback"
	case TypeOfPubrec:
		desc = "TypeOfPubrec"
	case TypeOfPubrel:
		desc = "TypeOfPubrel"
	case TypeOfPubcomp:
		desc = "TypeOfPubcomp"
	case TypeOfSubscribe:
		desc = "TypeOfSubscribe"
	case TypeOfSuback:
		desc = "TypeOfSuback"
	case TypeOfUnsubscribe:
		desc = "TypeOfUnsubscribe"
	case TypeOfUnsuback:
		desc = "TypeOfUnsuback"
	case TypeOfPingreq:
		desc = "TypeOfPingreq"
	case TypeOfPingresp:
		desc = "TypeOfPingresp"
	case TypeOfDisconnect:
		desc = "TypeOfDisconnect"
	default:
		desc = "not recognize"
	}
	return desc
}

// Header as defined in http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html#fixed-header
type Header struct {
	DUP    bool
	Retain bool
	QOS    uint8
}

// Connect represents an MQTT connect packet.
type Connect struct {
	ProtoName      []byte
	Version        uint8
	UsernameFlag   bool
	PasswordFlag   bool
	WillRetainFlag bool
	WillQOS        uint8
	WillFlag       bool
	CleanSeshFlag  bool
	KeepAlive      uint16
	ClientID       []byte
	WillTopic      []byte
	WillMessage    []byte
	Username       []byte
	Password       []byte
}

// Connack represents an MQTT connack packet.
// 0x00 connection accepted
// 0x01 refused: unacceptable proto version
// 0x02 refused: identifier rejected
// 0x03 refused server unavailiable
// 0x04 bad user or password
// 0x05 not authorized
type Connack struct {
	ReturnCode uint8
}

// Publish represents an MQTT publish packet.
type Publish struct {
	Header
	Topic     []byte
	MessageID uint16
	Payload   []byte
}

//Puback is sent for QOS level one to verify the receipt of a publish
//Qoth the spec: "A PUBACK message is sent by a server in response to a PUBLISH message from a publishing client, and by a subscriber in response to a PUBLISH message from the server."
type Puback struct {
	MessageID uint16
}

//Pubrec is for verifying the receipt of a publish
//Qoth the spec:"It is the second message of the QoS level 2 protocol flow. A PUBREC message is sent by the server in response to a PUBLISH message from a publishing client, or by a subscriber in response to a PUBLISH message from the server."
type Pubrec struct {
	MessageID uint16
}

//Pubrel is a response to pubrec from either the client or server.
type Pubrel struct {
	MessageID uint16
	//QOS1
	Header Header
}

//Pubcomp is for saying is in response to a pubrel sent by the publisher
//the final member of the QOS2 flow. both sides have said "hey, we did it!"
type Pubcomp struct {
	MessageID uint16
}

//Subscribe tells the server which topics the client would like to subscribe to
type Subscribe struct {
	Header
	MessageID     uint16
	Subscriptions []TopicQOSTuple
}

//TopicQOSTuple is a struct for pairing the Qos and topic together
//for the QOS' pairs in unsubscribe and subscribe
type TopicQOSTuple struct {
	Qos   uint8
	Topic []byte
}

//Suback is to say "hey, you got it buddy. I will send you messages that fit this pattern"
type Suback struct {
	MessageID uint16
	Qos       []uint8
}

//Unsubscribe is the message to send if you don't want to subscribe to a topic anymore
type Unsubscribe struct {
	Header
	MessageID uint16
	Topics    []TopicQOSTuple
}

//Unsuback is to unsubscribe as suback is to subscribe
type Unsuback struct {
	MessageID uint16
}

//Pingreq is a keepalive
type Pingreq struct {
}

//Pingresp is for saying "hey, the server is alive"
type Pingresp struct {
}

//Disconnect is to signal you want to cease communications with the server
type Disconnect struct {
}

// Type returns the MQTT message type.
func (c *Connect) Type() uint8 {
	return TypeOfConnect
}

// String returns the name of mqtt operation.
func (c *Connect) String() string {
	return "connect"
}

// EncodeTo writes the encoded message to the underlying writer.
func (c *Connect) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	// Calculate the max length
	head, buf := array.Split(maxHeaderSize)

	// pack the proto name and version
	offset := writeString(buf, c.ProtoName)
	offset += writeUint8(buf[offset:], c.Version)

	// pack the flags
	var flagByte byte
	flagByte |= byte(boolToUInt8(c.UsernameFlag)) << 7
	flagByte |= byte(boolToUInt8(c.PasswordFlag)) << 6
	flagByte |= byte(boolToUInt8(c.WillRetainFlag)) << 5
	flagByte |= byte(c.WillQOS) << 3
	flagByte |= byte(boolToUInt8(c.WillFlag)) << 2
	flagByte |= byte(boolToUInt8(c.CleanSeshFlag)) << 1

	offset += writeUint8(buf[offset:], flagByte)
	offset += writeUint16(buf[offset:], c.KeepAlive)
	offset += writeString(buf[offset:], c.ClientID)

	if c.WillFlag {
		offset += writeString(buf[offset:], c.WillTopic)
		offset += writeString(buf[offset:], c.WillMessage)
	}

	if c.UsernameFlag {
		offset += writeString(buf[offset:], c.Username)
	}

	if c.PasswordFlag {
		offset += writeString(buf[offset:], c.Password)
	}

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfConnect, nil, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// EncodeTo writes the encoded message to the underlying writer.
func (c *Connack) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	//write padding
	head, buf := array.Split(maxHeaderSize)
	offset := writeUint8(buf, byte(0))
	offset += writeUint8(buf[offset:], byte(c.ReturnCode))

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfConnack, nil, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (c *Connack) Type() uint8 {
	return TypeOfConnack
}

// String returns the name of mqtt operation.
func (c *Connack) String() string {
	return "connack"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Publish) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	length := 2 + len(p.Topic) + len(p.Payload)
	if p.QOS > 0 {
		length += 2
	}

	if length > maxMessageSize {
		return 0, ErrMessageTooLarge
	}

	// Write the packet
	offset := writeString(buf, p.Topic)
	if p.Header.QOS > 0 {
		offset += writeUint16(buf[offset:], p.MessageID)
	}

	copy(buf[offset:], p.Payload)
	offset += len(p.Payload)

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfPublish, &p.Header, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (p *Publish) Type() uint8 {
	return TypeOfPublish
}

// String returns the name of mqtt operation.
func (p *Publish) String() string {
	return "pub"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Puback) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, p.MessageID)

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfPuback, nil, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (p *Puback) Type() uint8 {
	return TypeOfPuback
}

// String returns the name of mqtt operation.
func (p *Puback) String() string {
	return "puback"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pubrec) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, p.MessageID)

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfPubrec, nil, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (p *Pubrec) Type() uint8 {
	return TypeOfPubrec
}

// String returns the name of mqtt operation.
func (p *Pubrec) String() string {
	return "pubrec"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pubrel) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, p.MessageID)

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfPubrel, &p.Header, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (p *Pubrel) Type() uint8 {
	return TypeOfPubrel
}

// String returns the name of mqtt operation.
func (p *Pubrel) String() string {
	return "pubrel"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pubcomp) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, p.MessageID)

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfPubcomp, nil, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (p *Pubcomp) Type() uint8 {
	return TypeOfPubcomp
}

// String returns the name of mqtt operation.
func (p *Pubcomp) String() string {
	return "pubcomp"
}

// EncodeTo writes the encoded message to the underlying writer.
func (s *Subscribe) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, s.MessageID)
	for _, t := range s.Subscriptions {
		offset += writeString(buf[offset:], t.Topic)
		offset += writeUint8(buf[offset:], byte(t.Qos))
	}

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfSubscribe, &s.Header, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (s *Subscribe) Type() uint8 {
	return TypeOfSubscribe
}

// String returns the name of mqtt operation.
func (s *Subscribe) String() string {
	return "sub"
}

// EncodeTo writes the encoded message to the underlying writer.
func (s *Suback) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, s.MessageID)
	for _, q := range s.Qos {
		offset += writeUint8(buf[offset:], byte(q))
	}

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfSuback, nil, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (s *Suback) Type() uint8 {
	return TypeOfSuback
}

// String returns the name of mqtt operation.
func (s *Suback) String() string {
	return "suback"
}

// EncodeTo writes the encoded message to the underlying writer.
func (u *Unsubscribe) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, u.MessageID)
	for _, toptup := range u.Topics {
		offset += writeString(buf[offset:], toptup.Topic)
	}

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfUnsubscribe, &u.Header, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (u *Unsubscribe) Type() uint8 {
	return TypeOfUnsubscribe
}

// String returns the name of mqtt operation.
func (u *Unsubscribe) String() string {
	return "unsub"
}

// EncodeTo writes the encoded message to the underlying writer.
func (u *Unsuback) EncodeTo(w io.Writer) (int, error) {
	array := buffers.Get()
	defer buffers.Put(array)

	head, buf := array.Split(maxHeaderSize)
	offset := writeUint16(buf, u.MessageID)

	// Write the header in front and return the buffer
	start := writeHeader(head, TypeOfUnsuback, nil, offset)
	return w.Write(array.Slice(start, maxHeaderSize+offset))
}

// Type returns the MQTT message type.
func (u *Unsuback) Type() uint8 {
	return TypeOfUnsuback
}

// String returns the name of mqtt operation.
func (u *Unsuback) String() string {
	return "unsuback"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pingreq) EncodeTo(w io.Writer) (int, error) {
	return w.Write([]byte{0xc0, 0x0})
}

// Type returns the MQTT message type.
func (p *Pingreq) Type() uint8 {
	return TypeOfPingreq
}

// String returns the name of mqtt operation.
func (p *Pingreq) String() string {
	return "pingreq"
}

// EncodeTo writes the encoded message to the underlying writer.
func (p *Pingresp) EncodeTo(w io.Writer) (int, error) {
	return w.Write([]byte{0xd0, 0x0})
}

// Type returns the MQTT message type.
func (p *Pingresp) Type() uint8 {
	return TypeOfPingresp
}

// String returns the name of mqtt operation.
func (p *Pingresp) String() string {
	return "pingresp"
}

// EncodeTo writes the encoded message to the underlying writer.
func (d *Disconnect) EncodeTo(w io.Writer) (int, error) {
	return w.Write([]byte{0xe0, 0x0})
}

// Type returns the MQTT message type.
func (d *Disconnect) Type() uint8 {
	return TypeOfDisconnect
}

// String returns the name of mqtt operation.
func (d *Disconnect) String() string {
	return "disconnect"
}

// decodeHeader decodes the header
func decodeHeader(rdr Reader) (hdr Header, length uint32, messageType uint8, err error) {
	firstByte, err := rdr.ReadByte()
	if err != nil {
		log.Errorln("decodeHeader-----readbytes:", err)
		return Header{}, 0, 0, err
	}
	messageType = (firstByte & 0xf0) >> 4

	// Set the header depending on the message type
	switch messageType {
	case TypeOfPublish, TypeOfSubscribe, TypeOfUnsubscribe, TypeOfPubrel:
		DUP := firstByte&0x08 > 0
		QOS := firstByte & 0x06 >> 1
		retain := firstByte&0x01 > 0

		hdr = Header{
			DUP:    DUP,
			QOS:    QOS,
			Retain: retain,
		}
	}

	multiplier := uint32(1)
	digit := byte(0x80)

	// Read the length
	for (digit & 0x80) != 0 {
		b, err := rdr.ReadByte()
		if err != nil {
			log.Debugln("decodeHeader-----read 2 fail")
			return Header{}, 0, 0, err
		}

		digit = b
		length += uint32(digit&0x7f) * multiplier
		multiplier *= 128
	}

	return hdr, uint32(length), messageType, nil
}

func decodeConnect(data []byte) Message {
	//TODO: Decide how to recover rom invalid packets (offsets don't equal actual reading?)
	bookmark := uint32(0)

	protoname := readString(data, &bookmark)
	ver := uint8(data[bookmark])
	bookmark++
	flags := data[bookmark]
	bookmark++
	keepalive := readUint16(data, &bookmark)
	cliID := readString(data, &bookmark)
	connect := &Connect{
		ProtoName:      protoname,
		Version:        ver,
		KeepAlive:      keepalive,
		ClientID:       cliID,
		UsernameFlag:   flags&(1<<7) > 0,
		PasswordFlag:   flags&(1<<6) > 0,
		WillRetainFlag: flags&(1<<5) > 0,
		WillQOS:        (flags & (1 << 4)) + (flags & (1 << 3)),
		WillFlag:       flags&(1<<2) > 0,
		CleanSeshFlag:  flags&(1<<1) > 0,
	}

	if connect.WillFlag {
		connect.WillTopic = readString(data, &bookmark)
		connect.WillMessage = readString(data, &bookmark)
	}

	if connect.UsernameFlag {
		connect.Username = readString(data, &bookmark)
	}

	if connect.PasswordFlag {
		connect.Password = readString(data, &bookmark)
	}
	return connect
}

func decodeConnack(data []byte, _ Header) Message {
	//first byte is weird in connack
	bookmark := uint32(1)
	retcode := data[bookmark]

	return &Connack{
		ReturnCode: retcode,
	}
}

func decodePublish(data []byte, hdr Header) Message {
	bookmark := uint32(0)
	topic := readString(data, &bookmark)
	var msgID uint16
	if hdr.QOS > 0 {
		msgID = readUint16(data, &bookmark)
	}

	return &Publish{
		Header:    hdr,
		Topic:     topic,
		Payload:   data[bookmark:],
		MessageID: msgID,
	}
}

func decodePuback(data []byte) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Puback{
		MessageID: msgID,
	}
}

func decodePubrec(data []byte) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrec{
		MessageID: msgID,
	}
}

func decodePubrel(data []byte, hdr Header) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubrel{
		Header:    hdr,
		MessageID: msgID,
	}
}

func decodePubcomp(data []byte) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Pubcomp{
		MessageID: msgID,
	}
}

func decodeSubscribe(data []byte, hdr Header) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var topics []TopicQOSTuple
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		t.Topic = readString(data, &bookmark)
		qos := data[bookmark]
		bookmark++
		t.Qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Subscribe{
		Header:        hdr,
		MessageID:     msgID,
		Subscriptions: topics,
	}
}

func decodeSuback(data []byte) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	var qoses []uint8
	maxlen := uint32(len(data))
	//is this efficient
	for bookmark < maxlen {
		qos := data[bookmark]
		bookmark++
		qoses = append(qoses, qos)
	}
	return &Suback{
		MessageID: msgID,
		Qos:       qoses,
	}
}

func decodeUnsubscribe(data []byte, hdr Header) Message {
	bookmark := uint32(0)
	var topics []TopicQOSTuple
	msgID := readUint16(data, &bookmark)
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		//		qos := data[bookmark]
		//		bookmark++
		t.Topic = readString(data, &bookmark)
		//		t.qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Unsubscribe{
		Header:    hdr,
		MessageID: msgID,
		Topics:    topics,
	}
}

func decodeUnsuback(data []byte) Message {
	bookmark := uint32(0)
	msgID := readUint16(data, &bookmark)
	return &Unsuback{
		MessageID: msgID,
	}
}

func decodePingreq() Message {
	return &Pingreq{}
}

func decodePingresp() Message {
	return &Pingresp{}
}

func decodeDisconnect() Message {
	return &Disconnect{}
}

// DecodePacket decodes the packet from the provided reader.
func DecodePacket(rdr Reader, maxMessageSize int64) (Message, error) {
	hdr, sizeOf, messageType, err := decodeHeader(rdr)
	if err != nil {
		return nil, err
	}

	log.Infoln("defined", "server get msgtype:", msgDesc(messageType))
	// Check for empty packets
	switch messageType {
	case TypeOfPingreq:
		return &Pingreq{}, nil
	case TypeOfPingresp:
		return &Pingresp{}, nil
	case TypeOfDisconnect:
		return &Disconnect{}, nil
	}

	//check to make sure packet isn't above size limit
	if int64(sizeOf) > maxMessageSize {
		return nil, ErrMessageTooLarge
	}

	// Now we can decode the buffer. The problem here is that we have to create
	// a new buffer for the body as we're going to simply create slices around it.
	// There's probably a way to use a "read buffer" provided to reduce allocations.
	buffer := make([]byte, sizeOf)
	_, err = io.ReadFull(rdr, buffer)
	if err != nil {
		return nil, err
	}

	// Decode the body
	var msg Message
	switch messageType {
	case TypeOfConnect:
		msg = decodeConnect(buffer)
	case TypeOfConnack:
		msg = decodeConnack(buffer, hdr)
	case TypeOfPublish:
		msg = decodePublish(buffer, hdr)
	case TypeOfPuback:
		msg = decodePuback(buffer)
	case TypeOfPubrec:
		msg = decodePubrec(buffer)
	case TypeOfPubrel:
		msg = decodePubrel(buffer, hdr)
	case TypeOfPubcomp:
		msg = decodePubcomp(buffer)
	case TypeOfSubscribe:
		msg = decodeSubscribe(buffer, hdr)
	case TypeOfSuback:
		msg = decodeSuback(buffer)
	case TypeOfUnsubscribe:
		msg = decodeUnsubscribe(buffer, hdr)
	case TypeOfUnsuback:
		msg = decodeUnsuback(buffer)
	default:
		return nil, fmt.Errorf("Invalid zero-length packet with type %d", messageType)
	}

	return msg, nil
}

// -------------------------------------------------------------

// encodeParts sews the whole packet together
func writeHeader(buf []byte, msgType uint8, h *Header, length int) int {
	var firstByte byte
	firstByte |= msgType << 4
	if h != nil {
		firstByte |= boolToUInt8(h.DUP) << 3
		firstByte |= h.QOS << 1
		firstByte |= boolToUInt8(h.Retain)
	}

	// get the length first
	numBytes, bitField := encodeLength(uint32(length))
	offset := 6 - numBytes - 1 //to account for the first byte

	// now we blit it in
	buf[offset] = byte(firstByte)
	for i := offset + 1; i < 6; i++ {
		buf[i] = byte(bitField >> ((numBytes - 1) * 8))
		numBytes--
	}

	return int(offset)
}

func writeString(buf, v []byte) int {
	length := len(v)
	writeUint16(buf, uint16(length))
	copy(buf[2:], v)
	return 2 + length
}

func writeUint16(buf []byte, v uint16) int {
	buf[0] = byte((v & 0xff00) >> 8)
	buf[1] = byte(v & 0x00ff)
	return 2
}

func writeUint8(buf []byte, v uint8) int {
	buf[0] = v
	return 1
}

func readString(b []byte, startsAt *uint32) []byte {
	l := readUint16(b, startsAt)
	v := b[*startsAt : uint32(l)+*startsAt]
	*startsAt += uint32(l)
	return v
}

func readUint16(b []byte, startsAt *uint32) uint16 {
	b0 := uint16(b[*startsAt])
	b1 := uint16(b[*startsAt+1])
	*startsAt += 2

	return (b0 << 8) + b1
}

func boolToUInt8(v bool) uint8 {
	if v {
		return 0x1
	}

	return 0x0
}

// encodeLength encodes the length formatting (see: http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html#fixed-header)
// and tells us how many bytes it takes up.
func encodeLength(bodyLength uint32) (uint8, uint32) {
	if bodyLength == 0 {
		return 1, 0
	}

	var bitField uint32
	var numBytes uint8
	for bodyLength > 0 {
		bitField <<= 8
		dig := uint8(bodyLength % 128)
		bodyLength /= 128
		if bodyLength > 0 {
			dig = dig | 0x80
		}

		bitField |= uint32(dig)
		numBytes++
	}
	return numBytes, bitField
}

// ErrMessageTooLarge occurs when a message encoded/decoded is larger than max MQTT frame.
var ErrMessageTooLarge = errors.New("mqtt: message size exceeds 64K")
