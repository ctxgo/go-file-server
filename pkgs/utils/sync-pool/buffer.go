package syncpool

import (
	"bytes"
	"sync"
)

type BufferPool struct {
	*sync.Pool
}

func NewBufferPool() *BufferPool {
	return &BufferPool{&sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	},
	}
}

func (p *BufferPool) AcquireBuffer() *bytes.Buffer {
	return p.Get().(*bytes.Buffer)

}
func (p *BufferPool) ReleaseBuffer(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		p.Put(buf)
	}
}
