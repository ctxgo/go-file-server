package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v4"
)

func HandleFile(ctx context.Context, f archiver.File, dest string) (err error) {
	fpath := filepath.Join(dest, f.NameInArchive)

	if f.IsDir() {
		return os.MkdirAll(fpath, os.ModePerm)
	}

	if err = os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %w", fpath, err)
	}

	inFile, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip file entry %s: %w", f.NameInArchive, err)
	}
	defer func() {
		if cerr := inFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close input file %s: %w", f.NameInArchive, cerr)
		}
	}()

	// 使用os.OpenFile和适当的标志来避免覆盖现有文件
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", fpath, err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close output file %s: %w", fpath, cerr)
		}
	}()
	ctxReader := &ContextReader{ctx: ctx, reader: inFile}
	if _, err = io.Copy(outFile, ctxReader); err != nil {
		return fmt.Errorf("failed to copy contents to %s: %w", fpath, err)
	}
	return nil
}

// 带ctx的Reader
type ContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *ContextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(p)
	}
}
