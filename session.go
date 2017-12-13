package gomsg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

// ISessionHandler 会话处理器
type ISessionHandler interface {
	OnReq(s *Session, data []byte, ch chan *Result)
	OnPush(s *Session, data []byte) uint16
}

// Result (ErrorNum,Data)
type Result struct {
	En   uint16
	Data []byte
}

// Session 会话
type Session struct {
	ID       int32
	conn     net.Conn
	handler  ISessionHandler
	bodyLen  uint16
	err      error
	reqSeed  uint16
	ppSeed   uint8
	requests map[uint16]chan *Result
}

// SetHandler 设置会话处理器
func (s *Session) SetHandler(handler ISessionHandler) {
	s.handler = handler
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
	input := bufio.NewScanner(s.conn)
	input.Split(s.split)

	for input.Scan() {
		// dispatch
		s.dispatch(input.Bytes())
	}

	s.Close()
}

// Close 关闭会话
func (s *Session) Close() {
	s.conn.Close()
	log.Printf("conn [%d] closed.\n", s.ID)
}

func (s *Session) dispatch(data []byte) {
	//log.Printf("conn : %d=> Read [% x]\n", s.ID, data)

	if len(data) < 2 {
		return
	}

	reader := bytes.NewBuffer(data)
	_, err := reader.ReadByte() // cnt
	if err != nil {
		s.Close()
		log.Println(err.Error())
		return
	}

	pattern, err := reader.ReadByte()
	left := len(data) - 2

	switch pattern {
	case Push:
		s.onPush(reader, left)
	case Request:
		go s.onReq(reader, left)
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
		s.Close()
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
		s.Close()
		log.Println("")
		return
	}

	if s.handler == nil {
		return
	}

	ret := s.handler.OnPush(s, body)
	if ret != 0 {
		log.Printf("onPush : %d\n", ret)
	}
}

func (s *Session) onResponse(reader *bytes.Buffer, left int) {
	var serial uint16
	var en uint16
	err := binary.Read(reader, binary.LittleEndian, &serial)
	if err != nil {
		s.Close()
		log.Println(err.Error())
		return
	}

	left -= 2
	err = binary.Read(reader, binary.LittleEndian, &en)
	if err != nil {
		s.Close()
		log.Println(err.Error())
		return
	}

	left -= 2
	body := make([]byte, left)
	n, err := reader.Read(body)
	if n != left {
		s.Close()
		log.Println("")
		return
	}

	if err != nil {
		s.Close()
		log.Println(err.Error())
		return
	}

	var req chan *Result
	var exists bool
	if req, exists = s.requests[serial]; !exists {
		log.Printf("%d not exist in request.\n", serial)
		return
	}

	req <- &Result{En: en, Data: body}
}

func (s *Session) onReq(reader *bytes.Buffer, left int) {
	var serial uint16
	err := binary.Read(reader, binary.LittleEndian, &serial)
	if err != nil {
		s.Close()
		log.Println(err.Error())
		return
	}

	left -= 2
	body := make([]byte, left)
	n, err := reader.Read(body)
	if n != left {
		s.Close()
		log.Println("")
		return
	}

	if err != nil {
		s.Close()
		log.Println(err.Error())
		return
	}

	if s.handler == nil {
		s.response(serial, &Result{En: NoHandler, Data: nil})
		return
	}

	ch := make(chan *Result, 1)
	defer close(ch)
	s.handler.OnReq(s, body, ch)
	ret := <-ch

	s.response(serial, ret)
}

// Write raw send interface
func (s *Session) Write(data []byte) {
	n, err := s.conn.Write(data)
	if n != len(data) || err != nil {
		log.Println("Write error")
	}

	//log.Printf("conn : %d=> Write [% x]\n", s.ID, data)
}

// Request request remote to response
func (s *Session) Request(data []byte) *Result {
	if len(data) == 0 {
		return &Result{En: RequestDataIsEmpty}
	}

	s.reqSeed++
	if _, exists := s.requests[s.reqSeed]; exists {
		return &Result{En: SerialConflict}
	}

	req := make(chan *Result)
	defer close(req)

	// record, let 'response' package know which chan to notify
	s.requests[s.reqSeed] = req

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+2+len(data)))
	buf.WriteByte(1)
	buf.WriteByte(Request)
	binary.Write(buf, binary.LittleEndian, s.reqSeed)
	if len(data) != 0 {
		buf.Write(data)
	}

	ret := <-req
	delete(s.requests, s.reqSeed)

	return ret
}

// Push push to remote without response
func (s *Session) Push(data []byte) uint16 {
	if len(data) == 0 {
		return PushDataIsEmpty
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+len(data)))
	buf.WriteByte(1)
	buf.WriteByte(Push)
	if len(data) != 0 {
		buf.Write(data)
	}

	s.Write(buf.Bytes())
	return Success
}

// Pub ...
func (s *Session) Pub(subject string, data []byte) {
	subBytes := []byte(subject)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+len(subBytes)+1+len(data)))
	buf.WriteByte(1)
	buf.WriteByte(Pub)
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
	buf.WriteByte(Sub)
	buf.Write(subBytes)

	s.Write(buf.Bytes())
}

// Ping ping remote, returns delay seconds
func (s *Session) Ping() uint32 {
	return 0
}

func (s *Session) pong(serial byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+1))
	buf.WriteByte(1)
	buf.WriteByte(Pong)
	buf.WriteByte(serial)

	s.Write(buf.Bytes())
}

func (s *Session) response(serial uint16, ret *Result) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(1+1+2+2+len(ret.Data)))
	buf.WriteByte(1)
	buf.WriteByte(Response)
	binary.Write(buf, binary.LittleEndian, serial)
	binary.Write(buf, binary.LittleEndian, ret.En)
	if len(ret.Data) != 0 {
		buf.Write(ret.Data)
	}

	s.Write(buf.Bytes())
}
