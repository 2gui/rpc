
package rpc

import (
	"io"
)

type Cmd = uint8

const (
	CmdPing   Cmd = 0x00
	CmdPong   Cmd = 0x01
	CmdDef    Cmd = 0x02
	CmdCall   Cmd = 0x03
	CmdReturn Cmd = 0x04
	CmdError  Cmd = 0x05
)

type Command interface{
	WriteTo(io.Writer)(int64, error)
	ReadFrom(io.Reader, *Point)(int64, error)
}

type CommandNewer = func()(Command)

type Errid = uint16

const (
	ErrString    Errid = 0x00
	ErrNotExists Errid = 0x01
	ErrArgs      Errid = 0x02
)

type RemoteErr struct{
	id Errid
	msg string
}

func (e *RemoteErr)Error()(string){
	switch e.id {
	case ErrString: return "remote: " + e.msg
	case ErrNotExists: return "remote: id not exists: " + e.msg
	case ErrArgs: return "arguments error: " + e.msg
	default: return "remote unknown error: " + e.msg
	}
}
