package gomsg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"time"
)

// Result (ErrorNum,Data)
type Result struct {
	En   int16
	Data []byte
}

// Succeed if ret is success
func (ret *Result) Succeed() bool {
	return ret.En == 0
}

// Session 会话
type Session struct {
	ID   int32
	conn net.Conn
	node *Node

	bodyLen      uint16
	err          error
	reqSeed      uint16
	ppSeed       uint8
	reqPool      map[uint16]chan *Result
	closed       bool
	responseTime int64
}

// NewSession make session
func newSession(id int32, conn net.Conn, n *Node) *Session {
	return &Session{
		ID:           id,
		conn:         conn,
		node:         n,
		reqPool:      make(map[uint16]chan *Result),
		responseTime: time.Now().Unix(),
	}
}

func (s *Session) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	len := len(data)
	offset := 0
	if len == 0 {
		return 0, nil, nil
	}

	if atEOF {
		return len, nil, nil
	}

	if s.bodyLen == 0 {
		if len < 2 {
			// Request more data.
			return 0, nil, nil
		}

		s.bodyLen = binary.LittleEndian.Uint16(data[offset:2])
		len -= 2
		offset += 2

		if len < int(s.bodyLen) {
			return 2, nil, nil
		}

	} else if len < int(s.bodyLen) {
		// Request more data.
		return 0, nil, nil
	}

	advance = int(s.bodyLen) + offset
	s.bodyLen = 0
	return advance, data[offset:advance], nil
}

func (s *Session) scan() {
	defer Recover()

	input := bufio.NewScanner(s.conn)
	input.Split(s.split)

	for input.Scan() {
		// dispatch
		s.dispatch(input.Bytes())

		s.node.ReadCounter <- 1
	}

	s.Close(false)
}

// Close 关闭会话
func (s *Session) Close(force bool) {
	if s.closed {
		return
	}

	s.closed = true
	s.conn.Close()

	log.Printf("conn [%d] closed.\n", s.ID)
	s.node.internalHandler.OnClose(s, force)
}

// IsClosed 是否关闭
func (s *Session) Closed() bool {
	return s == nil || s.closed
}

func (s *Session) elapsedSinceLastResponse() int {
	return int(time.Now().Unix() - s.responseTime)
}

func (s *Session) dispatch(data []byte) {
	//log.Printf("conn : %d=> Read [% x]\n", s.ID, data)
	s.responseTime = time.Now().Unix()

	if len(data) < 2 {
		return
	}

	reader := bytes.NewBuffer(data)
	_, err := reader.ReadByte() // cnt
	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	pattern, err := reader.ReadByte()
	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	left := len(data) - 2

	switch Pattern(pattern) {
	case Push:
		s.onPush(reader, left)
	case Request:
		s.onReq(reader, left)
	case Response:
		s.onResponse(reader, left)
	case Ping:
		s.onPing(reader)
	case Pong:
		s.onPong(reader)
	case Sub:
	case Unsub:
	case Pub:
	}
}

func (s *Session) onPing(reader *bytes.Buffer) {
	serial, err := reader.ReadByte()
	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	s.pong(serial)
}

// todo lizs
func (s *Session) onPong(reader *bytes.Buffer) {
}

func (s *Session) onPush(reader *bytes.Buffer, left int) {
	body := make([]byte, left)
	n, err := reader.Read(body)
	if n != left || err != nil {
		s.Close(false)
		log.Println("")
		return
	}

	ret := s.node.internalHandler.OnPush(s, body)
	if ret != 0 {
		log.Printf("onPush : %d\n", ret)
	}
}

func (s *Session) onResponse(reader *bytes.Buffer, left int) {
	var serial uint16
	var en int16
	err := binary.Read(reader, binary.LittleEndian, &serial)
	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	left -= 2
	err = binary.Read(reader, binary.LittleEndian, &en)
	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	left -= 2
	body := make([]byte, left)
	n, err := reader.Read(body)
	if n != left {
		s.Close(false)
		log.Println("")
		return
	}

	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	req, exists := s.reqPool[serial]
	if !exists {
		log.Printf("%d not exist in req pool.\n", serial)
		return
	}

	delete(s.reqPool, serial)
	req <- &Result{En: en, Data: body}
}

func (s *Session) onReq(reader *bytes.Buffer, left int) {
	var serial uint16
	err := binary.Read(reader, binary.LittleEndian, &serial)
	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	left -= 2
	body := make([]byte, left)
	n, err := reader.Read(body)
	if n != left {
		s.Close(false)
		log.Println("")
		return
	}

	if err != nil {
		s.Close(false)
		log.Println(err.Error())
		return
	}

	go s.node.internalHandler.OnReq(s, body, func(r *Result) {
		s.response(serial, r)
	})
}

// Write raw send interface
func (s *Session) Write(data []byte) error {
	n, err := s.conn.Write(data)
	if n != len(data) {
		if err != nil {
			log.Println("Write error =>", err.Error())
			s.Close(false)
		} else {
			log.Printf("Write error => writed : %d != expected : %d", n, len(data))
		}
	} else {
		s.node.WriteCounter <- 1
	}

	return err
}

// Request request remote to response
func (s *Session) Request(data []byte) *Result {
	if len(data) == 0 {
		return &Result{En: int16(RequestDataIsEmpty)}
	}

	s.reqSeed++
	if _, exists := s.reqPool[s.reqSeed]; exists {
		return &Result{En: int16(SerialConflict)}
	}

	req := make(chan *Result)
	// record, let 'response' package know which chan to notify
	s.reqPool[s.reqSeed] = req

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+2+len(data)))
	buf.WriteByte(1)
	buf.WriteByte(byte(Request))
	binary.Write(buf, binary.LittleEndian, s.reqSeed)
	if len(data) != 0 {
		buf.Write(data)
	}

	err := s.Write(buf.Bytes())
	if err != nil {
		return &Result{En: int16(Write)}
	}

	return <-req
}

// Push push to remote without response
func (s *Session) Push(data []byte) NetError {
	if len(data) == 0 {
		return PushDataIsEmpty
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+len(data)))
	buf.WriteByte(1)
	buf.WriteByte(byte(Push))
	if len(data) != 0 {
		buf.Write(data)
	}

	err := s.Write(buf.Bytes())
	if err != nil {
		return Write
	}

	return Success
}

// Pub ...
func (s *Session) Pub(subject string, data []byte) {
	subBytes := []byte(subject)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+len(subBytes)+1+len(data)))
	buf.WriteByte(1)
	buf.WriteByte(byte(Pub))
	buf.Write(subBytes)
	buf.WriteByte(0) // append \0 to string
	buf.Write(data)

	s.Write(buf.Bytes())
}

// Sub ...
func (s *Session) Sub(subject string) {
	subBytes := []byte(subject)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+len(subBytes)))
	buf.WriteByte(1)
	buf.WriteByte(byte(Sub))
	buf.Write(subBytes)

	s.Write(buf.Bytes())
}

// Ping ping remote, returns delay seconds
func (s *Session) Ping() {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+1))
	buf.WriteByte(1)
	buf.WriteByte(byte(Ping))
	buf.WriteByte(0)

	s.Write(buf.Bytes())
}

func (s *Session) pong(serial byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+1))
	buf.WriteByte(1)
	buf.WriteByte(byte(Pong))
	buf.WriteByte(serial)

	s.Write(buf.Bytes())
}

func (s *Session) response(serial uint16, ret *Result) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+2+2+len(ret.Data)))
	buf.WriteByte(1)
	buf.WriteByte(byte(Response))
	binary.Write(buf, binary.LittleEndian, serial)
	binary.Write(buf, binary.LittleEndian, ret.En)
	if len(ret.Data) != 0 {
		buf.Write(ret.Data)
	}

	s.Write(buf.Bytes())
}
