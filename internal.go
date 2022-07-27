
package rpc

import (
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

func (c *PingCmd)ReadFrom(r io.Reader, ctx *Context)(n int64, err error){
	c.Data, n, err = encoding.ReadUint64(r)
	if err != nil {
		return
	}
	err = ctx.SendCommand(PONG, &PongCmd{
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

func (c *PongCmd)ReadFrom(r io.Reader, ctx *Context)(n int64, err error){
	c.Data, n, err = encoding.ReadUint64(r)
	if err != nil {
		return
	}
	go func(){ ctx.pingch <- struct{}{} }()
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

func (c *DefCmd)ReadFrom(r io.Reader, ctx *Context)(n int64, err error){
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
	ctx.defs[c.Name] = c.Id
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

func (c *CallCmd)ReadFrom(r io.Reader, ctx *Context)(n int64, err error){
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
	if c.Id > (uint32)(len(ctx.funcs)) {
		panic("Unbinded function id")
	}
	fuc := ctx.funcs[c.Id]
	fuct := fuc.Type()
	if fuct.NumIn() != (int)(l) {
		panic("Arguments length not same")
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
		if c.Id >= (uint32)(len(ctx.funcs)) {
			return
		}
		out := fuc.Call(c.Args)
		var res reflect.Value
		if len(out) != 0 {
			er, ie := out[0].Interface().(error)
			if len(out) == 2 {
				er = out[1].Interface().(error)
				ie = true
			}
			if ie {
				if er != nil {
					err := ctx.SendCommand(ERROR, &ErrorCmd{
						Session: c.Session,
						Err: er.Error(),
					})
					if err != nil {
						panic(err)
					}
				}
			}else{
				res = out[0]
			}
		}
		err := ctx.SendCommand(RETURN, &ReturnCmd{
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

func (c *ReturnCmd)ReadFrom(r io.Reader, ctx *Context)(n int64, err error){
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
		ctx.lock.Lock()
		ch, ok := ctx.waiting[c.Session]
		if ok {
			delete(ctx.waiting, c.Session)
		}
		ctx.lock.Unlock()
		if ok {
			if c.Res.IsValid() {
				ch <- c.Res.Interface()
			}else{
				ch <- struct{}{}
			}
		}
	}()
	return
}

type ErrorCmd struct{
	Session uint32
	Err string
}

var _ Command = (*ErrorCmd)(nil)

func (c *ErrorCmd)WriteTo(w io.Writer)(n int64, err error){
	var n0 int64
	n, err = encoding.WriteUint32(w, c.Session)
	if err != nil {
		return
	}
	n0, err = encoding.WriteString(w, c.Err)
	n += n0
	return
}

func (c *ErrorCmd)ReadFrom(r io.Reader, ctx *Context)(n int64, err error){
	var n0 int64
	c.Session, n, err = encoding.ReadUint32(r)
	if err != nil {
		return
	}
	c.Err, n0, err = encoding.ReadString(r)
	n += n0
	if err != nil {
		return
	}
	go func(){
		ctx.lock.Lock()
		ch, ok := ctx.waiting[c.Session]
		if ok {
			delete(ctx.waiting, c.Session)
		}
		ctx.lock.Unlock()
		if ok {
			ch <- &RemoteErr{c.Err}
		}
	}()
	return
}
