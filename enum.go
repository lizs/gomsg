package gomsg

const (
	Success = iota
	ExceptionCatched
	Write
	Read
	RequestDataIsEmpty
	SerialConflict
	NoHandler
	ReadErrorNo
	SessionClosed
	End
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
