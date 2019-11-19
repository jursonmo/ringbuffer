package ringbuffer

import (
	"strings"
	"testing"
)

var dataSize int = 300

// func BenchmarkRingBuffer_Sync(b *testing.B) {
// 	rb := New(1024)
// 	data := []byte(strings.Repeat("a", dataSize))
// 	buf := make([]byte, 512)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		rb.Write(data)
// 		rb.Read(buf)
// 	}
// }

// func BenchmarkRingBuffer_AsyncRead(b *testing.B) {
// 	rb := New(1024)
// 	data := []byte(strings.Repeat("a", dataSize))
// 	buf := make([]byte, dataSize)

// 	go func() {
// 		for {
// 			rb.Read(buf)
// 		}
// 	}()

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		rb.Write(data)
// 	}
// }

func BenchmarkRingBuffer_MultiAsyncRead(b *testing.B) {
	rb := New(1024)
	data := []byte(strings.Repeat("a", dataSize))
	buf := make([]byte, dataSize)
	buf1 := make([]byte, dataSize)
	go func() {
		for {
			rb.Read(buf)
		}
	}()

	go func() {
		for {
			rb.Read(buf1)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for {
			if n, _ := rb.Write(data); n > 0 {
				break
			}
		}
	}
}

// func BenchmarkRingBuffer_AsyncWrite(b *testing.B) {
// 	rb := New(1024)
// 	data := []byte(strings.Repeat("a", dataSize))
// 	buf := make([]byte, dataSize)

// 	go func() {
// 		for {
// 			rb.Write(data)
// 		}
// 	}()

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		rb.Read(buf)
// 	}
// }
