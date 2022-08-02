
package rpc

import (
	"fmt"
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

const (
	SesCall uint32 = (1 << 16) - 1
	SesThread uint32 = ^SesCall
)

type Errid = uint64

const (
	ErrString    Errid = 0x00 << 32
	ErrNotExists Errid = 0x01 << 32
	ErrArgs      Errid = 0x02 << 32
	ErrCustom    Errid = 0xff << 32

	ErrIdMask     Errid = 0xffffffff << 32
	ErrCustomMask Errid = ^ErrIdMask
)

type RemoteErr struct{
	id Errid
	msg string
}

func (e *RemoteErr)ErrId()(uint64){
	return e.id
}

func (e *RemoteErr)CustomId()(uint64){
	if e.id & ErrIdMask == ErrCustom {
		return e.id & ErrCustomMask
	}
	return 0
}

func (e *RemoteErr)Error()(string){
	switch e.id & ErrIdMask {
	case ErrString: return "remote: " + e.msg
	case ErrNotExists: return "remote: id not exists: " + e.msg
	case ErrArgs: return "arguments error: " + e.msg
	case ErrCustom: return fmt.Sprintf("remote: custom id: %d: %s", e.CustomId(), e.msg)
	default: return "remote unknown error: " + e.msg
	}
}
