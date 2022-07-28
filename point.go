
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

type sessionT struct{
	ret chan <- any
}

type Point struct{
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
	sessions map[uint32]sessionT

	pingch chan struct{}
}

func NewPoint(w io.Writer, r io.ReadCloser)(p *Point){
	return NewPointWithCtx(w, r, context.Background())
}

func NewPointWithCtx(w io.Writer, r io.ReadCloser, ctx context.Context)(p *Point){
	ctx0, cancel := context.WithCancel(ctx)
	p = &Point{
		flag: false,
		w: w,
		r: r,
		ctx: ctx0,
		cancel: cancel,
		cmds: make(map[Cmd]CommandNewer),
		defs: make(map[string]uint32),
		sessions: make(map[uint32]sessionT),
		pingch: make(chan struct{}),
	}
	p.DefineCmd(CmdPing, func()(Command){ return &PingCmd{} })
	p.DefineCmd(CmdPong, func()(Command){ return &PongCmd{} })
	p.DefineCmd(CmdDef, func()(Command){ return &DefCmd{} })
	p.DefineCmd(CmdCall, func()(Command){ return &CallCmd{} })
	p.DefineCmd(CmdReturn, func()(Command){ return &ReturnCmd{} })
	p.DefineCmd(CmdError, func()(Command){ return &ErrorCmd{} })
	return
}

func (p *Point)Ping()(err error){
	if !p.flag {
		panic("Listener not start")
	}
	err = p.SendCommand(CmdPing, &PingCmd{
		Data: 0xff,
	})
	if err != nil {
		return
	}
	<-p.pingch
	return
}

func (p *Point)DefineCmd(id Cmd, newer CommandNewer){
	if _, ok := p.cmds[id]; ok {
		panic("Command id already exists")
	}
	if newer == nil {
		panic("Newer cannot be nil")
	}
	p.cmds[id] = newer
}

func (p *Point)Register(name string, fuc any)(err error){
	fv := reflect.ValueOf(fuc)
	if fv.Kind() != reflect.Func {
		panic("Except a method")
	}
	p.funcs = append(p.funcs, fv)
	i := (uint32)(len(p.funcs) - 1)
	return p.SendCommand(CmdDef, &DefCmd{
		Id: i,
		Name: name,
	})
}

func (p *Point)Call(name string, args ...any)(res any, err error){
	vals := make([]reflect.Value, len(args))
	for i, v := range args {
		vals[i] = reflect.ValueOf(v)
	}
	id, ok := p.defs[name]
	if !ok {
		return nil, fmt.Errorf("No method named '%s'", name)
	}
	ret := make(chan any, 1)

	p.lock.Lock()
	p.sesinc++
	sesid := p.sesinc
	for {
		if _, ok := p.sessions[sesid]; !ok {
			break
		}
		sesid++
	}
	p.sessions[sesid] = sessionT{
		ret: ret,
	}
	p.lock.Unlock()

	err = p.SendCommand(CmdCall, &CallCmd{
		Id: id,
		Session: sesid,
		Args: vals,
	})
	if err != nil {
		return
	}
	res0 := <- ret
	close(ret)
	if err0, ok := res0.(*RemoteErr); ok {
		err = err0
		return
	}
	res = res0
	return
}

func (p *Point)SendCommand(id Cmd, cmd Command)(err error){
	buf := bytes.NewBuffer(nil)
	_, err = buf.Write([]byte{id})
	if err != nil {
		return
	}
	_, err = cmd.WriteTo(buf)
	if err != nil {
		return
	}
	p.wl.Lock()
	defer p.wl.Unlock()
	_, err = encoding.WriteUint32(p.w, (uint32)(buf.Len()))
	if err != nil {
		return
	}
	_, err = buf.WriteTo(p.w)
	return
}

func (p *Point)Listen(){
	if p.flag {
		panic("Already listening")
	}
	p.flag = true
	go p.listener()
}

func (p *Point)Wait()(err error){
	select{
	case <-p.ctx.Done():
		return p.err
	}
}

func (p *Point)ListenAndWait()(err error){
	p.Listen()
	return p.Wait()
}

func (p *Point)Close()(err error){
	err = p.r.Close()
	if wc, ok := p.w.(io.Closer); ok {
		wc.Close()
	}
	return
}

func (p *Point)listener(){
	var (
		l uint32
		err error
	)
	defer p.cancel()
	for {
		l, _, err = encoding.ReadUint32(p.r)
		if err != nil {
			break
		}
		buf := make([]byte, l)
		_, err = io.ReadFull(p.r, buf[:l])
		if err != nil {
			break
		}
		{
			id := (Cmd)(buf[0])
			newer, ok := p.cmds[id]
			// println(">>> recv cmd:", id, ok)
			if !ok {
				continue
			}
			cmd := newer()
			_, err := cmd.ReadFrom(bytes.NewReader(buf[1:]), p)
			if err != nil {
				println("error: " + err.Error(), id)
				continue
			}
		}
	}
	p.err = err
}
