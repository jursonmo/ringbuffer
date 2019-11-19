package ringbuffer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/jursonmo/ringbuffer/internal/pmath"
)

type ringhead struct {
	head uint64
	pad  [128 - unsafe.Sizeof(uint64(0))]byte
}

type ringtail struct {
	tail uint64
	pad  [128 - unsafe.Sizeof(uint64(0))]byte
}

type ringrese struct {
	rese uint64
	pad  [128 - unsafe.Sizeof(uint64(0))]byte
}

type RingBuffer struct {
	ringhead // next position to write
	ringtail // next position to read
	ringrese
	size   int
	mask   int
	isFull bool
	mu     sync.Mutex //only for writer sync
	buf    []byte
}

func assert(b bool) {
	if !b {
		panic("")
	}
}

var test bool = false

func init() {
	//回绕也没问题
	max := uint64(1<<64 - 1)
	min := uint64(0)
	if min-max != 1 {
		panic("")
	}
	min = uint64(1 << 63)
	n := min - max
	if n != uint64(1<<63)+1 {
		panic("")
	}
	fmt.Println(n)
}

func New(size int) *RingBuffer {
	size = pmath.CeilToPowerOfTwo(size)
	assert(pmath.IsPowerOfTwo(size))

	return &RingBuffer{
		buf:  make([]byte, size),
		size: size,
		mask: size - 1,
	}
}

func (r *RingBuffer) Cap() int {
	return r.size
}

func (r *RingBuffer) Read(p []byte) (n int, err error) {
	rese := uint64(0)
	rn := len(p)
	if rn == 0 {
		return 0, nil
	}
	assert(len(r.buf) == r.size)

	for {
		rese = atomic.LoadUint64(&r.rese)
		//如果读完r.rese,另一个reader 更新了r.rese 和r.tail,writer 就可以继续更新r.head,这时就会发生head-rese >r.size的情况
		head := atomic.LoadUint64(&r.head)
		if head == rese {
			//empty
			return 0, nil
		}
		//idx = int64(head) - int64(rese)//rese 不可能超过head, 这种方式是比较大小的时候有用
		idx := head - rese //head回绕也没问题
		assert(idx > 0)
		if idx > uint64(r.size) {
			//means multi readers and there is much buf to read, so try again
			continue
			//why ? head=33360 , rese=33660, idx=18446744073709551316, uint64(r.size+1)=1025
			//get rese first and get head second: head=22422016 , rese=22411264, idx=10752, r.size=1024
			//fmt.Printf("head=%d , rese=%d, idx=%d, r.size=%d\n", head, rese, idx, r.size)
			//panic("")
		}

		//n is min(idx, rn)
		if int(idx) < rn {
			n = int(idx)
		} else {
			n = rn
		}

		if test && n != 512 {
			fmt.Printf("head=%d , rese=%d, idx=%d, r.size=%d\n", head, rese, idx, r.size)
			panic("")
		}

		if atomic.CompareAndSwapUint64(&r.rese, rese, uint64(n)+rese) {
			break
		}
	}

	assert(len(r.buf) == r.size)
	assert(n > 0)

	start := int(rese & uint64(r.mask))
	ncopy := 0
	if start+n <= r.size {
		ncopy = copy(p[0:n], r.buf[start:start+n])
	} else {
		ncopy = copy(p[:n], r.buf[start:r.size])
		ncopy += copy(p[ncopy:], r.buf[:n-ncopy])
	}

	if ncopy != n {
		fmt.Printf("ncopy=%d , n=%d, start=%d \n", ncopy, n, start)
		panic("")
	}
	for {
		if atomic.CompareAndSwapUint64(&r.tail, rese, uint64(n)+rese) {
			break
		}
	}
	return
}

func (r *RingBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	assert(len(r.buf) == r.size)

	head := atomic.LoadUint64(&r.head)
	tail := atomic.LoadUint64(&r.tail)
	//idx := int64(head) - int64(tail)
	idx := head - tail
	if idx < 0 {
		panic("idx < 0")
	}
	if int(idx) > r.size {
		fmt.Printf(" head:%d,  tail:%d, idx:%d\n", head, tail, idx)
		panic("")
	}
	if int(idx) == r.size {
		//full, no space to write
		return
	}
	windex := int(head & uint64(r.mask))
	rindex := int(tail & uint64(r.mask))
	if windex < rindex {
		n = copy(r.buf[windex:rindex], p)
		//testing
		if test && n != 512 {
			fmt.Printf(" head:%d,  tail:%d, idx:%d, ncopy:%d \n", head, tail, idx, n)
			panic("")
		}
		atomic.AddUint64(&r.head, uint64(n))
		return
	}
	// windex >= rindex
	ncopy := copy(r.buf[windex:r.size], p)
	if ncopy < len(p) {
		ncopy += copy(r.buf[:rindex], p[ncopy:])
	}
	n = ncopy
	//testing
	if test && n != 512 {
		fmt.Printf(" head:%d,  tail:%d, idx:%d, ncopy:%d \n", head, tail, idx, ncopy)
		panic("")
	}
	atomic.AddUint64(&r.head, uint64(n))
	return
}
