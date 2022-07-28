
package rpc

import (
	"fmt"
	"io"
	"reflect"

	encoding "github.com/2gui/rpc/encoding"
)

type PingCmd struct{
	Data uint64
}

var _ Command = (*PingCmd)(nil)

func (c *PingCmd)WriteTo(w io.Writer)(n int64, err error){
	return encoding.WriteUint64(w, c.Data)
}

func (c *PingCmd)ReadFrom(r io.Reader, p *Point)(n int64, err error){
	c.Data, n, err = encoding.ReadUint64(r)
	if err != nil {
		return
	}
	err = p.SendCommand(CmdPong, &PongCmd{
		Data: c.Data,
	})
	return
}

type PongCmd struct{
	Data uint64
}

var _ Command = (*PongCmd)(nil)

func (c *PongCmd)WriteTo(w io.Writer)(n int64, err error){
	return encoding.WriteUint64(w, c.Data)
}

func (c *PongCmd)ReadFrom(r io.Reader, p *Point)(n int64, err error){
	c.Data, n, err = encoding.ReadUint64(r)
	if err != nil {
		return
	}
	go func(){ p.pingch <- struct{}{} }()
	return
}

type DefCmd struct{
	Id uint32
	Name string
}

var _ Command = (*DefCmd)(nil)

func (c *DefCmd)WriteTo(w io.Writer)(n int64, err error){
	var n0 int64
	n, err = encoding.WriteUint32(w, c.Id)
	if err != nil {
		return
	}
	n0, err = encoding.WriteString(w, c.Name)
	n += n0
	return
}

func (c *DefCmd)ReadFrom(r io.Reader, p *Point)(n int64, err error){
	var n0 int64
	c.Id, n, err = encoding.ReadUint32(r)
	if err != nil {
		return
	}
	c.Name, n0, err = encoding.ReadString(r)
	n += n0
	if err != nil {
		return
	}
	p.defs[c.Name] = c.Id
	return
}

type CallCmd struct{
	Id uint32
	Session uint32
	Args []reflect.Value
}

var _ Command = (*CallCmd)(nil)

func (c *CallCmd)WriteTo(w io.Writer)(n int64, err error){
	var n0 int64
	n, err = encoding.WriteUint32(w, c.Id)
	if err != nil {
		return
	}
	n0, err = encoding.WriteUint32(w, c.Session)
	n += n0
	if err != nil {
		return
	}
	n0, err = encoding.WriteUint16(w, (uint16)(len(c.Args)))
	n += n0
	if err != nil {
		return
	}
	for _, v := range c.Args {
		n0, err = encoding.WriteValue(w, v)
		n += n0
		if err != nil {
			return
		}
	}
	return
}

func (c *CallCmd)ReadFrom(r io.Reader, p *Point)(n int64, err error){
	var n0 int64
	c.Id, n, err = encoding.ReadUint32(r)
	if err != nil {
		return
	}
	c.Session, n0, err = encoding.ReadUint32(r)
	n += n0
	if err != nil {
		return
	}
	var l uint16
	l, n0, err = encoding.ReadUint16(r)
	n += n0
	if err != nil {
		return
	}
	if c.Id > (uint32)(len(p.funcs)) {
		err = p.SendCommand(CmdError, &ErrorCmd{
			Session: c.Session,
			Errid: ErrNotExists,
			Msg: fmt.Sprintf("Unbinded function id '%d'", c.Id),
		})
		return
	}
	fuc := p.funcs[c.Id]
	fuct := fuc.Type()
	if fuct.NumIn() != (int)(l) {
		err = p.SendCommand(CmdError, &ErrorCmd{
			Session: c.Session,
			Errid: ErrArgs,
			Msg: fmt.Sprintf("Arguments length not same for '%d', except %d but have %d", c.Id, fuct.NumIn(), l),
		})
		return
	}
	c.Args = make([]reflect.Value, l)
	for i := 0; i < (int)(l); i++ {
		c.Args[i], n0, err = encoding.ReadValue(r, fuct.In(i))
		n += n0
		if err != nil {
			return
		}
	}
	go func(){
		var err error
		if c.Id >= (uint32)(len(p.funcs)) {
			return
		}
		out := fuc.Call(c.Args)
		var res reflect.Value
		if len(out) != 0 {
			var (
				er error
				ie bool
			)
			if len(out) == 2 {
				er = out[1].Interface().(error)
				ie = true
			}else{
				er, ie = out[0].Interface().(error)
			}
			if ie {
				if er != nil {
					err = p.SendCommand(CmdError, &ErrorCmd{
						Session: c.Session,
						Errid: ErrString,
						Msg: er.Error(),
					})
					if err != nil {
						panic(err)
					}
					return
				}
			}else{
				res = out[0]
			}
		}
		err = p.SendCommand(CmdReturn, &ReturnCmd{
			Session: c.Session,
			Res: res,
		})
		if err != nil {
			panic(err)
		}
	}()
	return
}

type ReturnCmd struct{
	Session uint32
	Res reflect.Value
}

var _ Command = (*ReturnCmd)(nil)

func (c *ReturnCmd)WriteTo(w io.Writer)(n int64, err error){
	var n0 int64
	n, err = encoding.WriteUint32(w, c.Session)
	if err != nil {
		return
	}
	ok := c.Res.IsValid()
	n0, err = encoding.WriteBool(w, ok)
	n += n0
	if err != nil {
		return
	}
	if ok {
		n0, err = encoding.WriteType(w, c.Res.Type())
		n += n0
		if err != nil {
			return
		}
		n0, err = encoding.WriteValue(w, c.Res)
		n += n0
	}
	return
}

func (c *ReturnCmd)ReadFrom(r io.Reader, p *Point)(n int64, err error){
	var n0 int64
	c.Session, n, err = encoding.ReadUint32(r)
	if err != nil {
		return
	}
	var ok bool
	ok, n0, err = encoding.ReadBool(r)
	n += n0
	if err != nil {
		return
	}
	if ok {
		var typ reflect.Type
		typ, n0, err = encoding.ReadType(r)
		n += n0
		if err != nil {
			return
		}
		c.Res, n0, err = encoding.ReadValue(r, typ)
		n += n0
		if err != nil {
			return
		}
	}
	go func(){
		p.lock.Lock()
		ses, ok := p.sessions[c.Session]
		if ok {
			delete(p.sessions, c.Session)
		}
		p.lock.Unlock()
		if ok {
			if c.Res.IsValid() {
				ses.ret <- c.Res.Interface()
			}else{
				ses.ret <- struct{}{}
			}
		}
	}()
	return
}

type ErrorCmd struct{
	Session uint32
	Errid Errid
	Msg string
}

var _ Command = (*ErrorCmd)(nil)

func (c *ErrorCmd)WriteTo(w io.Writer)(n int64, err error){
	var n0 int64
	n, err = encoding.WriteUint32(w, c.Session)
	if err != nil {
		return
	}
	n0, err = encoding.WriteUint16(w, c.Errid)
	n += n0
	if err != nil {
		return
	}
	n0, err = encoding.WriteString(w, c.Msg)
	n += n0
	return
}

func (c *ErrorCmd)ReadFrom(r io.Reader, p *Point)(n int64, err error){
	var n0 int64
	c.Session, n, err = encoding.ReadUint32(r)
	if err != nil {
		return
	}
	c.Errid, n0, err = encoding.ReadUint16(r)
	if err != nil {
		return
	}
	n += n0
	c.Msg, n0, err = encoding.ReadString(r)
	n += n0
	if err != nil {
		return
	}
	rerr := &RemoteErr{c.Errid, c.Msg}
	go func(){
		p.lock.Lock()
		ses, ok := p.sessions[c.Session]
		if ok {
			delete(p.sessions, c.Session)
		}
		p.lock.Unlock()
		if ok {
			ses.ret <- rerr
		}
	}()
	return
}
