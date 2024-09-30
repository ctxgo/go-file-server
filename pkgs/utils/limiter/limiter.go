package limiter

import (
	"context"
	"io"

	"golang.org/x/time/rate"
)

type ILimiter interface {
	LimitReader(ctx context.Context, reader io.Reader, opts ...opt) io.Reader
	LimitWriter(ctx context.Context, writer io.Writer, opts ...opt) io.Writer
	LimitReadertWriter(ctx context.Context, readWriter io.ReadWriter, opts ...opt) io.ReadWriter
}

// Limiter 是一个带宽限制器结构体
type Limiter struct {
	limiter *rate.Limiter
}

// NewLimiter 创建一个新的带宽限制器，
// rateLimit 是每秒的速率限制（以字节为单位），
// burst 是突发容量（以字节为单位）
func NewLimiter(rateLimitBytes uint64, burstBytes uint64) *Limiter {
	if rateLimitBytes == 0 {
		return &Limiter{
			limiter: rate.NewLimiter(rate.Inf, 0),
		}
	}
	// 内部转换：将字节转换为千字节(KB)来减少令牌的生成频率
	rateLimitKB := rate.Limit(rateLimitBytes / 1024)
	burstKB := burstBytes / 1024
	return &Limiter{
		limiter: rate.NewLimiter(rateLimitKB, int(burstKB)),
	}
}

// LimitReader 返回一个带限速功能的 io.Reader
// 默认缓冲区是8kb，可通过WithChunkSize调整
func (l *Limiter) LimitReader(ctx context.Context, reader io.Reader, opts ...opt) io.Reader {
	r := l.defaultLimitReaderWriter(ctx, opts...)
	r.reader = reader
	return r
}

func (l *Limiter) LimitWriter(ctx context.Context, writer io.Writer, opts ...opt) io.Writer {
	w := l.defaultLimitReaderWriter(ctx, opts...)
	w.writer = writer
	return w
}

func (l *Limiter) LimitReadertWriter(ctx context.Context, readWriter io.ReadWriter, opts ...opt) io.ReadWriter {
	r := l.defaultLimitReaderWriter(ctx, opts...)
	r.reader = readWriter
	r.writer = readWriter
	return r
}

func (l *Limiter) defaultLimitReaderWriter(ctx context.Context, opts ...opt) *limitReaderWriter {
	r := &limitReaderWriter{
		ctx:       ctx,
		limiter:   l.limiter,
		chunkSize: 8192,
		bufioSize: 8192,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

type opt func(*limitReaderWriter)

// WithChunkSize 设置在读/写操作中使用的数据块大小。默认8kb
// 这个大小直接决定了每次操作从速率限制器中消耗的令牌数量，从而显著影响数据传输的频率和平滑性。
// 较小的块大小使得数据传输更为频繁，增加传输的平滑性，但同时增加了调用速率限制器的计算开销。
// 较大的块大小减少了数据传输的频次，提高了提升整体的吞吐量，但在速率控制上的平滑性较低，
func WithChunkSize(size int) opt {
	return func(l *limitReaderWriter) {
		l.chunkSize = size
	}
}

// limitReader 是实现了 io.Reader 的结构体，用于限速读取
type limitReaderWriter struct {
	ctx       context.Context
	limiter   *rate.Limiter
	reader    io.Reader
	writer    io.Writer
	bufioSize int
	chunkSize int
}

func (r *limitReaderWriter) Read(p []byte) (int, error) {
	return r.processIO(p, r.reader.Read)

}

func (w *limitReaderWriter) Write(p []byte) (int, error) {
	return w.processIO(p, w.writer.Write)

}

func (lrw *limitReaderWriter) processIO(p []byte, operation func([]byte) (int, error)) (int, error) {
	totalProcessed := 0
	length := len(p)

	for totalProcessed < length {
		chunkSize := min(lrw.chunkSize, length-totalProcessed)
		if err := lrw.limiter.WaitN(lrw.ctx, chunkSize/1024); err != nil {
			return totalProcessed, err
		}

		n, err := operation(p[totalProcessed : totalProcessed+chunkSize])
		if err != nil {
			return totalProcessed, err
		}
		totalProcessed += n
	}

	return totalProcessed, nil
}
