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
	OnReq(s *Session, data []byte) *Result
	OnPush(s *Session, data []byte) uint16
}

// Result (ErrorNum,Data)
type Result struct {
	En   uint16
	Data []byte
}

// Session 会话
type Session struct {
	ID      int32
	conn    net.Conn
	handler ISessionHandler
	bodyLen uint16
	err     error
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

	ad := int(s.bodyLen)
	s.bodyLen = 0
	return offset + ad, data[offset : offset+ad], nil
}

func (s *Session) scan() {
	input := bufio.NewScanner(s.conn)
	input.Split(s.split)

	for input.Scan() {
		// dispatch
		s.dispatch(input.Bytes())
	}

	log.Printf("conn [%d] closed.\n", s.ID)
}

// Close 关闭会话
func (s *Session) Close() {
	s.conn.Close()
}

func (s *Session) dispatch(data []byte) {
	log.Printf("conn : %d=> Read [% x]\n", s.ID, data)

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
	case Request:
		s.onReq(reader, left)
	case Response:
	case Ping:
		s.onPing(reader)
	case Pong:
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

	ret := s.handler.OnReq(s, body)
	s.response(serial, ret)
}

// Write raw send interface
func (s *Session) Write(data []byte) {
	n, err := s.conn.Write(data)
	if n != len(data) || err != nil {
		log.Println("Write error")
	}

	log.Printf("conn : %d=> Write [% x]\n", s.ID, data)
}

func (s *Session) Request(data []byte) {

}

func (s *Session) Push(data []byte) {

}

func (s *Session) Pub(data []byte) {

}

func (s *Session) Sub(data []byte) {

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
