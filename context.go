
package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"

	encoding "github.com/2gui/rpc/encoding"
)

type Context struct{
	flag bool

	w io.Writer
	wl sync.Mutex
	r io.ReadCloser
	err error
	ctx context.Context
	cancel context.CancelFunc

	cmds map[Cmd]CommandNewer
	funcs []reflect.Value
	defs map[string]uint32

	lock sync.Mutex
	sesinc uint32
	waiting map[uint32]chan <- any

	pingch chan struct{}
}

func NewContext(w io.Writer, r io.ReadCloser)(c *Context){
	return NewContextWithCtx(w, r, context.Background())
}

func NewContextWithCtx(w io.Writer, r io.ReadCloser, ctx context.Context)(c *Context){
	ctx0, cancel := context.WithCancel(ctx)
	c = &Context{
		flag: false,
		w: w,
		r: r,
		ctx: ctx0,
		cancel: cancel,
		cmds: make(map[Cmd]CommandNewer),
		defs: make(map[string]uint32),
		waiting: make(map[uint32]chan <- any),
		pingch: make(chan struct{}),
	}
	c.DefineCmd(PING, func()(Command){ return &PingCmd{} })
	c.DefineCmd(PONG, func()(Command){ return &PongCmd{} })
	c.DefineCmd(DEF, func()(Command){ return &DefCmd{} })
	c.DefineCmd(CALL, func()(Command){ return &CallCmd{} })
	c.DefineCmd(RETURN, func()(Command){ return &ReturnCmd{} })
	c.DefineCmd(ERROR, func()(Command){ return &ErrorCmd{} })
	return
}

func (c *Context)Ping()(err error){
	if !c.flag {
		panic("Listener not start")
	}
	err = c.SendCommand(PING, &PingCmd{
		Data: 0xff,
	})
	if err != nil {
		return
	}
	<-c.pingch
	return
}

func (c *Context)DefineCmd(id Cmd, newer CommandNewer){
	if _, ok := c.cmds[id]; ok {
		panic("Command id already exists")
	}
	if newer == nil {
		panic("Newer cannot be nil")
	}
	c.cmds[id] = newer
}

func (c *Context)Register(name string, fuc any)(err error){
	fv := reflect.ValueOf(fuc)
	if fv.Kind() != reflect.Func {
		panic("Except a method")
	}
	c.funcs = append(c.funcs, fv)
	i := (uint32)(len(c.funcs) - 1)
	return c.SendCommand(DEF, &DefCmd{
		Id: i,
		Name: name,
	})
}

func (c *Context)Call(name string, args ...any)(res any, err error){
	vals := make([]reflect.Value, len(args))
	for i, v := range args {
		vals[i] = reflect.ValueOf(v)
	}
	id, ok := c.defs[name]
	if !ok {
		fmt.Println("c.defs", c.defs, id)
		return nil, fmt.Errorf("No method named '%s'", name)
	}
	ch := make(chan any, 1)
	c.lock.Lock()
	c.sesinc++
	session := c.sesinc
	for {
		if _, ok := c.waiting[session]; !ok {
			break
		}
		session++
	}
	c.waiting[session] = ch
	c.lock.Unlock()
	err = c.SendCommand(CALL, &CallCmd{
		Id: id,
		Session: session,
		Args: vals,
	})
	if err != nil {
		return
	}
	res0 := <- ch
	if err0, ok := res0.(*RemoteErr); ok {
		err = err0
		return
	}
	res = res0
	return
}

func (c *Context)SendCommand(id Cmd, cmd Command)(err error){
	buf := bytes.NewBuffer(nil)
	_, err = buf.Write([]byte{id})
	if err != nil {
		return
	}
	_, err = cmd.WriteTo(buf)
	if err != nil {
		return
	}
	c.wl.Lock()
	defer c.wl.Unlock()
	_, err = encoding.WriteUint32(c.w, (uint32)(buf.Len()))
	if err != nil {
		return
	}
	_, err = buf.WriteTo(c.w)
	return
}

func (c *Context)Listen(){
	if c.flag {
		panic("Already listening")
	}
	c.flag = true
	go c.listener()
}

func (c *Context)Wait()(err error){
	select{
	case <-c.ctx.Done():
		return c.err
	}
}

func (c *Context)ListenAndWait()(err error){
	c.Listen()
	return c.Wait()
}

func (c *Context)Close()(err error){
	err = c.r.Close()
	if wc, ok := c.w.(io.Closer); ok {
		wc.Close()
	}
	return
}

func (c *Context)listener(){
	var (
		l uint32
		err error
	)
	defer c.cancel()
	for {
		l, _, err = encoding.ReadUint32(c.r)
		if err != nil {
			break
		}
		buf := make([]byte, l)
		_, err = io.ReadFull(c.r, buf[:l])
		if err != nil {
			break
		}
		{
			id := (Cmd)(buf[0])
			newer, ok := c.cmds[id]
			if !ok {
				continue
			}
			cmd := newer()
			_, err := cmd.ReadFrom(bytes.NewReader(buf[1:]), c)
			if err != nil {
				println("error: " + err.Error(), id)
				continue
			}
		}
	}
	c.err = err
}
