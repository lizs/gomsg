package gomsg

import "log"

type rsp struct {
	session *Session
	serial  uint16
	en      uint16
	body    []byte
}

type push struct {
	session *Session
	body    []byte
}

type req struct {
	session *Session
	serial  uint16
	body    []byte
}

// Ret response chan
type Ret struct {
	session *Session
	serial  uint16
	ret     *Result
}

// NewRet new ret
func NewRet(s *Session, serial uint16, ret *Result) *Ret {
	return &Ret{session: s, serial: serial, ret: ret}
}

// STAService sta service
type STAService struct {
	rsp  chan *rsp
	push chan *push
	req  chan *req
	Ret  chan *Ret
}

var sta *STAService

// STA STA instance
func STA() *STAService {
	if sta == nil {
		sta = &STAService{
			rsp:  make(chan *rsp),
			push: make(chan *push),
			req:  make(chan *req, 0),
			Ret:  make(chan *Ret, 0),
		}
	}

	return sta
}

// Start start the STA service
func (s *STAService) Start() {
	for {
		select {
		case push := <-s.push:
			ret := push.session.handler.OnPush(push.session, push.body)
			if ret != 0 {
				log.Printf("onPush : %d\n", ret)
			}

		case rsp := <-s.rsp:
			req, exists := rsp.session.reqPool[rsp.serial]
			if !exists {
				log.Printf("%d not exist in req pool.\n", rsp.serial)
				return
			}

			req <- &Result{En: rsp.en, Data: rsp.body}

		case req := <-s.req:
			req.session.handler.OnReq(req.session, req.serial, req.body)

		case ret := <-s.Ret:
			ret.session.response(ret.serial, ret.ret)
		}
	}
}
