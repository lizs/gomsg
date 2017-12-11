package gomsg

import (
	"encoding/binary"
	"net"
)

// Session 会话
type Session struct {
	ID   int32
	Conn net.Conn

	bodyLen uint16
	err     error
}

// Scan scan header/body
func (s *Session) Scan(data []byte, atEOF bool) (advance int, token []byte, err error) {
	len := len(data)
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

		s.bodyLen = binary.LittleEndian.Uint16(data[:2])
		return 2, nil, nil
	}

	if len < int(s.bodyLen) {
		// Request more data.
		return 0, nil, nil
	}

	ad := int(s.bodyLen)
	s.bodyLen = 0
	return ad, data[:ad], nil
}
