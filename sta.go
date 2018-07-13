package gomsg

// import (
// 	"log"
// )

// type rsp struct {
// 	session *Session
// 	serial  uint16
// 	en      int16
// 	body    []byte
// }

// type push struct {
// 	session *Session
// 	body    []byte
// }

// type req struct {
// 	session *Session
// 	serial  uint16
// 	body    []byte
// }

// // Ret response chan
// type Ret struct {
// 	session *Session
// 	serial  uint16
// 	ret     *Result
// }

// // NewRet new ret
// func NewRet(s *Session, serial uint16, ret *Result) *Ret {
// 	return &Ret{session: s, serial: serial, ret: ret}
// }

// // STAService sta service
// type STAService struct {
// 	rsp chan *rsp
// }

// var sta *STAService

// // STA STA instance
// func STA() *STAService {
// 	if sta == nil {
// 		sta = &STAService{
// 			rsp: make(chan *rsp, 10000),
// 		}
// 	}

// 	return sta
// }

// func (s *STAService) startImp() {
// 	defer Recover()

// 	for {
// 		select {
// 		case rsp := <-s.rsp:
// 			req, exists := rsp.session.reqPool[rsp.serial]
// 			if !exists {
// 				log.Printf("%d not exist in req pool.\n", rsp.serial)
// 				return
// 			}

// 			delete(rsp.session.reqPool, rsp.serial)
// 			req <- &Result{En: rsp.en, Data: rsp.body}
// 		}
// 	}
// }

// // Start start the STA service
// func (s *STAService) Start() {
// 	go s.startImp()
// }
