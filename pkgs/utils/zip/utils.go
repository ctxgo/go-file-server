package zip

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func zipToWriter(ctx context.Context, writer io.Writer, option Option, inPaths ...string) error {
	bufferedWriter := bufio.NewWriterSize(writer, option.bufferSize)
	zipWriter := zip.NewWriter(bufferedWriter)
	for _, inPath := range inPaths {
		if err := addFileToZip(ctx, zipWriter, inPath, option); err != nil {
			return err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return err
	}
	return bufferedWriter.Flush()
}

func addFileToZip(ctx context.Context, zipWriter *zip.Writer, srcPath string, option Option) error {

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {

			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = getHeaderName(path, srcPath, option)
		if info.IsDir() {
			header.Name += "/"
		}
		if option.verbose {
			fmt.Printf("adding... %s\n", header.Name)
		}
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return addFileContent(ctx, writer, path)
	})
}

// getHeaderName 计算文件在 zip 中的路径
func getHeaderName(path, srcPath string, option Option) string {
	if option.baseDir == "" {
		return strings.TrimPrefix(path, filepath.Dir(srcPath)+"/")

	}
	if path != srcPath {
		relPath := strings.TrimPrefix(path, srcPath)
		baseDir := filepath.Base(srcPath)
		return filepath.Join(option.baseDir, baseDir, relPath)
	}
	return filepath.Join(option.baseDir, filepath.Base(path))

}

func addFileContent(ctx context.Context, writer io.Writer, path string) error {
	fileReader, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%w: %s", err, path)
	}
	defer fileReader.Close()

	ctxWriter := &ContextWriter{writer: writer, ctx: ctx}
	_, err = io.Copy(ctxWriter, fileReader)
	return err
}

// 带ctx的Writer
type ContextWriter struct {
	ctx    context.Context
	writer io.Writer
}

func (cw *ContextWriter) Write(p []byte) (int, error) {
	select {
	case <-cw.ctx.Done():
		return 0, cw.ctx.Err()
	default:
		return cw.writer.Write(p)
	}
}
