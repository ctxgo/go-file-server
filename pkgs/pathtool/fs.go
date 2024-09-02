package pathtool

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/filetype"
)

type Filetool struct {
	Path string
	fs.FileInfo
	Err error
}

type Options struct {
	OnlyDir          bool        //是否只显示目录
	RecursiveFind    bool        //遍历目录的时候是否递归查找子目录
	MkdirAll         bool        // 是否递归创建目录
	CreateIfNotExist bool        // 文件不存在时是否创建
	FileMode         os.FileMode // 文件/目录权限设置
	AppendMode       bool        // 当设置为 true 时，如果文件存在，则追加内容；如果为 false，则覆盖现有内容

}

type OptionSetter func(*Options)

type FilesDetails struct {
	Name    string
	Path    string
	Type    string
	ModTime time.Time
	Size    int64
	Err     error
}

func WithOnlyDir(b bool) OptionSetter {
	return func(opt *Options) {
		opt.OnlyDir = b
	}
}

func WithRecursive(recursive bool) OptionSetter {
	return func(opt *Options) {
		opt.RecursiveFind = recursive
	}
}

func WithCreateIfNotExist(create bool) OptionSetter {
	return func(opt *Options) {
		opt.CreateIfNotExist = create
	}
}

func WithFileMode(mode os.FileMode) OptionSetter {
	return func(opt *Options) {
		opt.FileMode = mode
	}
}

// WithAppendMode 设置文件写入模式。如果 append 为 true，则追加内容到现有文件。
// 如果 append 为 false，则覆盖现有文件内容。
func WithAppendMode(append bool) OptionSetter {
	return func(opt *Options) {
		opt.AppendMode = append
	}
}

// WithRecursive 设置是否递归创建目录
func WithMkDirAll(mkAll bool) OptionSetter {
	return func(opt *Options) {
		opt.MkdirAll = mkAll
	}
}

func NewFiletool(path string) *Filetool {
	info, err := os.Stat(path)

	return &Filetool{
		Path:     path,
		FileInfo: info,
		Err:      err,
	}
}

func (f *Filetool) IsExist() (bool, error) {
	if f.Err == nil {
		return true, nil
	}
	if os.IsNotExist(f.Err) {
		return false, nil
	}
	return false, f.Err
}

func (f *Filetool) AssertFile() (bool, error) {
	if f.Err != nil {
		return false, f.Err
	}
	if f.FileInfo.IsDir() {
		return false, nil
	}
	return true, nil
}

func (f *Filetool) AssertDir() (bool, error) {
	ok, err := f.AssertFile()
	if err != nil {
		return false, err
	}
	return !ok, err
}

func (f *Filetool) IsEmptyDir() (bool, error) {
	ok, err := f.AssertDir()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	dir, err := os.Open(f.Path)
	if err != nil {
		return false, err
	}
	defer dir.Close()
	_, err = dir.Readdirnames(1) // 试图读取目录中的一个条目
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func (f *Filetool) GetFsDetails() FilesDetails {
	if f.Err != nil {
		return FilesDetails{Path: f.Path, Err: f.Err}
	}
	return FilesDetails{
		Name:    f.Name(),
		Path:    f.Path,
		Err:     f.Err,
		Type:    f.getfiletype(),
		Size:    f.Size(),
		ModTime: f.ModTime(),
	}

}

func (f *Filetool) IterateFiles(setters ...OptionSetter) <-chan FilesDetails {
	options := &Options{}
	for _, setter := range setters {
		setter(options)
	}

	out := make(chan FilesDetails)

	go func() {
		defer close(out)

		walkFn := func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				out <- FilesDetails{Path: path, Err: err}
				return nil
			}

			if !options.RecursiveFind && path != f.Path && info.IsDir() {
				return filepath.SkipDir
			}

			if !info.IsDir() && options.OnlyDir {
				return nil
			}

			out <- (&Filetool{path, info, nil}).GetFsDetails()

			return nil
		}

		if err := filepath.Walk(f.Path, walkFn); err != nil {
			out <- FilesDetails{Path: f.Path, Err: err}
		}
	}()

	return out
}

func (f *Filetool) GetFiles(setters ...OptionSetter) []FilesDetails {
	var FilesLists []FilesDetails
	fileCh := f.IterateFiles(setters...)
	for f := range fileCh {
		FilesLists = append(FilesLists, f)
	}
	return FilesLists

}

// WriteToFileString writes a string to the file specified in filetool
func (f *Filetool) WriteToFileString(content string, setters ...OptionSetter) error {
	return f.WriteToFileByte([]byte(content), setters...)
}

func (f *Filetool) WriteToFileByte(content []byte, setters ...OptionSetter) error {
	// 初始化 Options 结构体
	options := Options{
		FileMode: 0644, // 默认文件模式
	}

	// 应用提供的设置函数
	for _, setter := range setters {
		setter(&options)
	}

	// 检查文件是否存在

	ok, err := f.AssertFile()
	if err != nil {
		if os.IsNotExist(err) && options.CreateIfNotExist {
			return os.WriteFile(f.Path, content, options.FileMode)
		}
		return err
	}
	if !ok {
		return errors.New("Unable to write to directory")
	}

	// 确定打开文件的模式：追加或覆盖
	flag := os.O_WRONLY
	if options.AppendMode {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	file, err := os.OpenFile(f.Path, flag, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入内容到文件
	_, err = file.Write(content)
	return err
}

func CreateDir(path string, setters ...OptionSetter) error {
	options := Options{
		FileMode: 0755, // 默认目录权限
	}

	for _, setter := range setters {
		setter(&options)
	}

	var err error
	if options.MkdirAll {
		// 递归创建目录
		err = os.MkdirAll(path, options.FileMode)
	} else {
		// 非递归创建单个目录
		err = os.Mkdir(path, options.FileMode)
	}

	return err
}

func (f *Filetool) getfiletype() string {
	if f.IsDir() {
		return "dir"
	}
	var f_buffer []byte = make([]byte, 261)
	_f, _ := os.Open(f.Path)
	defer _f.Close()
	n, _ := _f.Read(f_buffer)
	contentType, _ := filetype.Match(f_buffer[0:n])
	if contentType == filetype.Unknown {
		fext := strings.Split(f.Path, ".")
		if len(fext) > 1 {
			return fext[len(fext)-1]
		}
		return "file"
	}
	return contentType.Extension
}
