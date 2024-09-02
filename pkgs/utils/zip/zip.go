package zip

import (
	"context"
	"io"
)

type Zipper interface {
	Zip(...string) error
	ZipWithCtx(context.Context, ...string) error
}

type Option struct {
	baseDir    string
	verbose    bool
	bufferSize int
}

type Options func(*Option)

// WithBaseDir 设置压缩文件中的基础目录
func WithBaseDir(baseDir string) Options {
	return func(o *Option) {
		o.baseDir = baseDir
	}
}

// WithBufferSize 设置writer缓冲区,默认16kb
func WithBufferSize(size int) Options {
	return func(o *Option) {
		o.bufferSize = size
	}
}

// WithVerbose 设置是否打印详细信息
func WithVerbose(verbose bool) Options {
	return func(o *Option) {
		o.verbose = verbose
	}
}

// NewZipToFile creates a Zipper that zips content directly to a file, using provided options.
func NewFileZip(outputPath string, opts ...Options) Zipper {
	return &fileZipper{outputPath: outputPath, option: configureOptions(opts...)}
}

// NewZipToWriter creates a Zipper that zips content to an arbitrary io.Writer, using provided options.
func NewStreamZip(writer io.Writer, opts ...Options) Zipper {
	return &streamZipper{writer: writer, option: configureOptions(opts...)}
}

func configureOptions(opts ...Options) Option {
	option := Option{bufferSize: 16 * 1024}
	for _, opt := range opts {
		opt(&option)
	}
	return option
}
