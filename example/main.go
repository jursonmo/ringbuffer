package main

import (
	"fmt"
	"time"

	"github.com/jursonmo/ringbuffer"
)

func main() {
	// rb := ringbuffer.New(1024)
	// rb.Write([]byte("abcd"))
	// //fmt.Println(rb.Len())
	// //fmt.Println(rb.Free())
	// buf := make([]byte, 4)

	// rb.Read(buf)
	// fmt.Println(string(buf))
	// // Output: 4
	// // 1020
	// // abcd

	dataSize := 512
	myrb := ringbuffer.New(1024)
	wn := 1000
	r1 := uint32(0)
	r2 := uint32(0)
	go func() {
		buf := make([]byte, dataSize)
		for {
			n, _ := myrb.Read(buf)
			if n > 0 {
				if n != dataSize {
					panic("")
				}
				r1++
			}
		}
	}()
	go func() {
		buf := make([]byte, dataSize)
		for {
			n, _ := myrb.Read(buf)
			if n > 0 {
				if n != dataSize {
					panic("")
				}
				r2++
			}
		}
	}()

	data := make([]byte, dataSize)

	n := 0
	for i := 0; i < wn; i++ {
		//binary.BigEndian.PutUint32(data[:4], uint32(i))
		for {
			n, _ = myrb.Write(data)
			if n > 0 {
				break
			}
		}
	}
	fmt.Println("write over")
	for {
		time.Sleep(time.Second)
		fmt.Printf("wn:%d, r1:%d, r2:%d\n", wn, r1, r2)
		if r1+r2 == uint32(wn) {
			break
		}
	}

}
