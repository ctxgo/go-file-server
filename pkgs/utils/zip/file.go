package zip

import (
	"context"
	"fmt"
	"os"
)

// fileZipper implements the Zipper interface, zipping to a file.
type fileZipper struct {
	outputPath string
	option     Option
}

func (s *fileZipper) Zip(inPaths ...string) error {
	return s.ZipWithCtx(context.Background(), inPaths...)
}

func (f *fileZipper) ZipWithCtx(ctx context.Context, inPaths ...string) error {
	file, err := os.Create(f.outputPath)
	if err != nil {
		return fmt.Errorf("%w: %s", err, f.outputPath)
	}
	defer file.Close()
	return zipToWriter(ctx, file, f.option, inPaths...)
}
