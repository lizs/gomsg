package gomsg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

const (
	Push     = byte(0)
	Request  = byte(1)
	Response = byte(2)
	Ping     = byte(3)
	Pong     = byte(4)
	Sub      = byte(5)
	Unsub    = byte(6)
	Pub      = byte(7)
)

const (
	PatternSize = 1
)

// ISessionHandler 会话处理器
type ISessionHandler interface {
	OnReq(data []byte) Result
	OnPush(data []byte) uint16
}

type Result struct {
	en   uint16
	data []byte
}

// Session 会话
type Session struct {
	ID   int32
	Conn net.Conn

	bodyLen uint16
	err     error
}

func (s *Session) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
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

func (s *Session) scan() {
	input := bufio.NewScanner(s.Conn)
	input.Split(s.split)

	for input.Scan() {
		// dispatch
		s.dispatch(input.Bytes())
	}

	log.Printf("conn [%d] closed.\n", s.ID)
}

// Close 关闭会话
func (s *Session) Close() {

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
	switch pattern {
	case Push:
	case Request:
		var serial uint16
		err = binary.Read(reader, binary.LittleEndian, serial)
		if err != nil {
			s.Close()
			log.Println(err.Error())
			return
		}

		go dispatchReq()
	case Response:
	case Ping:
		serial, err := reader.ReadByte()
		if err != nil {
			s.Close()
			log.Println(err.Error())
			return
		}

		s.pong(serial)
	case Pong:
	case Sub:
	case Unsub:
	case Pub:
	}
}

func (s *Session) dispatchReq() {

}

func (s *Session) Write(data []byte) {
	n, err := s.Conn.Write(data)
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

/*
  using (var ms = new MemoryStream(message))
            using (var br = new BinaryReader(ms))
            {
                var pattern = (Pattern) br.ReadByte();
                left -= PatternSize;
                switch (pattern)
                {
                    case Pattern.Push:
                        _dispatcher.OnPush(this, br.ReadBytes(left));
                        break;

                    case Pattern.Request:
                    {
                        var serial = br.ReadUInt16();
                        left -= SerialSize;
                        try
                        {
#if NET35
                                _dispatcher.OnRequest(this, br.ReadBytes(left), ret =>
                                {
                                    Response(ret.Err, serial, ret.Data, null);
                                    ret.Callback?.Invoke(ret.CallbackData);
                                });
#else
                            var ret = await _dispatcher.OnRequest(this, br.ReadBytes(left));
                            Response(ret.Err, serial, ret.Data, null);
                            ret.Callback?.Invoke(ret.CallbackData);
#endif
                        }
                        catch (Exception e)
                        {
                            Response((ushort) NetError.ExceptionCatched, serial, null, null);
                            Logger.Ins.Exception("Pattern.Request", e);
                        }
                        break;
                    }

                    case Pattern.Response:
                    {
                        var serial = br.ReadUInt16();
                        left -= SerialSize;
                        var err = br.ReadUInt16();
                        left -= ErrorNoSize;
                        OnResponse(serial, err, br.ReadBytes(left));
                        break;
                    }

                    case Pattern.Ping:
                    {
                        var serial = br.ReadByte();
                        Pong(serial);
                        break;
                    }

                    case Pattern.Pong:
                    {
                        var serial = br.ReadByte();
                        if (_pings.ContainsKey(serial))
                        {
                            var info = _pings[serial];
                            info.Callback?.Invoke((int) (_watch.ElapsedMilliseconds - info.PingTime));
                            _pings.Remove(serial);
                        }

                        if (KeepAliveCounter > 0)
                            --KeepAliveCounter;

                        break;
                    }

                    case Pattern.Sub:
                        break;

                    case Pattern.Unsub:
                        break;

                    default:
                        return false;
                }
            }

            return true;
        }

*/
