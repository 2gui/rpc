
package main

import (
	"flag"
	"io"
	"os"

	"github.com/2gui/rpc"
)

var (
	readfd, writefd int = 3, 4
)

func init(){
	flag.IntVar(&readfd, "reader", readfd, "The file descriptor for read")
	flag.IntVar(&writefd, "writer", writefd, "The file descriptor for write")
	flag.Parse()
}

func main(){
	println("started")
	r, w := os.NewFile((uintptr)(readfd), "reader"), os.NewFile((uintptr)(writefd), "writer")
	p := rpc.NewPoint(w, r)
	p.Register("helloWorld", func()(string){
		println("hello world")
		return "hello world"
	})
	p.Register("add", func(a, b int)(int){
		println("adding", a, b)
		return a + b
	})
	println("listening")
	err := p.ListenAndWait()
	if err != nil && err != io.EOF {
		panic(err)
	}
}
