
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/2gui/rpc"
)

func startServer()(cmd *exec.Cmd, r io.ReadCloser, w io.WriteCloser, err error){
	cmd = exec.Command("../example_server", "-reader", "3", "-writer", "4")
	cr, mw, err := os.Pipe()
	if err != nil {
		return
	}
	defer cr.Close()
	mr, cw, err := os.Pipe()
	if err != nil {
		mw.Close()
		return
	}
	defer cw.Close()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.ExtraFiles = append(cmd.ExtraFiles, cr, cw)
	err = cmd.Start()
	if err != nil {
		mw.Close()
		mr.Close()
		return
	}
	r, w = mr, mw
	return
}

func main(){
	cmd, r, w, err := startServer()
	if err != nil {
		panic(err)
	}
	defer w.Close()
	defer r.Close()

	exit := make(chan struct{}, 0)
	go func(){
		defer close(exit)
		err := cmd.Wait()
		if err != nil {
			panic(err)
		}
	}()
	p := rpc.NewPoint(w, r)
	p.Listen()
	err = p.Ping()
	if err != nil {
		panic(err)
	}
	res, err := p.Call("add", 1, 2)
	fmt.Println(res, err)
	w.Close()
	<-exit
}
