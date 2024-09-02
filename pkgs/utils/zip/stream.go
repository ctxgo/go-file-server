package zip

import (
	"context"
	"io"
)

// streamZipper implements the Zipper interface, zipping to an io.Writer.
type streamZipper struct {
	writer io.Writer
	option Option
}

func (s *streamZipper) Zip(inPaths ...string) error {
	return s.ZipWithCtx(context.Background(), inPaths...)
}

func (s *streamZipper) ZipWithCtx(ctx context.Context, inPaths ...string) error {
	return zipToWriter(ctx, s.writer, s.option, inPaths...)
}
