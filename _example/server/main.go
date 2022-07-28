
package main

import (
	"flag"
	"fmt"
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
	fmt.Println("started")
	r, w := os.NewFile((uintptr)(readfd), "reader"), os.NewFile((uintptr)(writefd), "writer")
	p := rpc.NewPoint(w, r)
	p.Register("helloWorld", func()(string){
		fmt.Println("calling: helloWorld")
		return "hello world"
	})
	p.Register("add", func(a, b int)(int){
		fmt.Println("adding:", a, b)
		return a + b
	})
	p.Register("error", func(args []int)(err error){
		err = fmt.Errorf("args error: %v", args)
		fmt.Println("gen error:", err)
		return
	})
	fmt.Println("listening")
	err := p.ListenAndWait()
	if err != nil && err != io.EOF {
		panic(err)
	}
}
