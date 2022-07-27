
package rpc

import (
	"io"
)

type Cmd = uint8

const (
	PING   Cmd = 0x00
	PONG   Cmd = 0x01
	DEF    Cmd = 0x02
	CALL   Cmd = 0x03
	RETURN Cmd = 0x04
	ERROR  Cmd = 0x05
)

type Command interface{
	WriteTo(io.Writer)(int64, error)
	ReadFrom(io.Reader, *Context)(int64, error)
}

type CommandNewer = func()(Command)

type RemoteErr struct{
	msg string
}

func (e *RemoteErr)Error()(string){
	return "remote: " + e.msg
}
