
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
	fmt.Println("initing")
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
	p.Register("add_ptr", func(a, b int, c *int){
		fmt.Println("adding ptr:", a, b, c)
		*c = a + b
		return
	})
	p.Register("dev_ptr", func(a, b int, c **int){
		if b == 0 {
			*c = nil
			return
		}
		**c = a / b
		return
	})
	p.Register("error", func(args []int)(err error){
		for i, _ := range args {
			if i != 0 {
				args[i] += args[i - 1]
			}
		}
		err = fmt.Errorf("test error: %v", args)
		return
	})
	fmt.Println("start listening")
	err := p.ListenAndWait()
	if err != nil && err != io.EOF {
		panic(err)
	}
}
