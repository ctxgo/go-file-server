package pathtool

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var fields = []struct {
	Name     string
	Mapping  *mapping.FieldMapping
	Analyzer string
}{
	{Name: "Path", Mapping: bleve.NewTextFieldMapping(), Analyzer: "keyword"},
	{Name: "ParentPath", Mapping: bleve.NewTextFieldMapping(), Analyzer: "keyword"},
	{Name: "Name", Mapping: bleve.NewTextFieldMapping(), Analyzer: "keyword"},
	{Name: "IsDir", Mapping: bleve.NewBooleanFieldMapping()},
}

type UpdateCallback func(*FileIndexer)

type storageType int

const (
	UseMem storageType = iota
	UseDisk
)

type FileIndexer struct {
	storageType    storageType
	IndexPath      string
	Index          bleve.Index
	WatchedRootDir string
	Logger         *zap.SugaredLogger
	mutex          sync.RWMutex
	enableWatch    bool
	watcher        *fsnotify.Watcher
	updateCallback UpdateCallback
}
type FileDocument struct {
	Name       string
	Path       string
	ParentPath string
	IsDir      bool
}

type Opt func(*FileIndexer)

func WithStorageType(t storageType) Opt {
	return func(fi *FileIndexer) {
		fi.storageType = t
	}
}

func WithEnableWatch(b bool) Opt {
	return func(fi *FileIndexer) {
		fi.enableWatch = b
	}
}

func WithLog(log *zap.SugaredLogger) Opt {
	return func(fi *FileIndexer) {
		fi.Logger = log
	}
}

func WithIndexPath(s string) Opt {
	return func(fi *FileIndexer) {
		fi.IndexPath = s
	}
}

func WithUpdateCallback(up UpdateCallback) Opt {

	return func(fi *FileIndexer) {
		fi.updateCallback = up
	}

}

func NewFileIndexer(path string, opts ...Opt) (*FileIndexer, error) {

	fileIndexer := &FileIndexer{
		storageType:    UseMem,
		WatchedRootDir: path,
	}
	for _, o := range opts {
		o(fileIndexer)
	}
	if fileIndexer.Logger == nil {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
		fileIndexer.Logger = logger.Sugar()
	}
	if fileIndexer.IndexPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		fileIndexer.IndexPath = homeDir
	}
	fileIndexer.IndexPath = filepath.Join(fileIndexer.IndexPath, ".bleve.index")

	//初始化索引文档
	if err := fileIndexer.IndexInit(); err != nil {
		return nil, errors.Wrap(err, "索引初始化失败")
	}
	if fileIndexer.enableWatch {
		if err := fileIndexer.initWatch(); err != nil {
			return nil, err
		}
	}
	return fileIndexer, nil
}

func (fi *FileIndexer) IndexInit() error {
	fi.mutex.Lock()
	defer fi.mutex.Unlock()
	ok, err := NewFiletool(fi.WatchedRootDir).AssertDir()
	if err != nil {
		return errors.Wrap(err, "检查目录失败")
	}
	if !ok {
		return errors.New("目录断言失败：" + fi.WatchedRootDir)
	}

	// 创建索引
	index, err := fi.createIndex()
	if err != nil {
		return err
	}
	fi.Index = index
	//添加条目
	if err := fi.addResource(fi.WatchedRootDir); err != nil {
		return err
	}
	go fi.printDocCount()
	return nil
}

func (fi *FileIndexer) createIndex() (bleve.Index, error) {
	// 创建文档映射
	fileMapping := bleve.NewDocumentMapping()

	for _, field := range fields {
		if field.Analyzer != "" {
			field.Mapping.Analyzer = field.Analyzer
		}
		fileMapping.AddFieldMappingsAt(field.Name, field.Mapping)
	}

	// 设置默认文档映射
	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = fileMapping

	if fi.storageType == UseMem {
		return bleve.NewMemOnly(indexMapping)

	}

	//如果索引文件则备份
	exist, err := NewFiletool(fi.IndexPath).IsExist()
	if err != nil {
		return nil, err
	}
	if exist {
		indexPathBakDir := filepath.Join(
			os.TempDir(), "index_bak_dir",
		)
		err := os.MkdirAll(indexPathBakDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
		indexPath := filepath.Join(indexPathBakDir,
			time.Now().Format("bleve.index.2006.01.02.15.04.05.000"))
		os.Rename(fi.IndexPath, indexPath)
	}

	return bleve.New(fi.IndexPath, indexMapping)
}

func (fi *FileIndexer) Search(req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	fi.mutex.RLock()
	defer fi.mutex.RUnlock()
	return fi.Index.Search(req)
}

func (fi *FileIndexer) IsSkippePath(path string) bool {
	return fi.storageType == UseDisk && strings.HasPrefix(path, fi.IndexPath)
}

func (fi *FileIndexer) DelResource(path string) error {
	fi.mutex.Lock()
	defer fi.mutex.Unlock()
	return fi.delResource(path)
}

func (fi *FileIndexer) delResource(path string) error {
	return fi.Index.Delete(path)
}

func (fi *FileIndexer) AddResource(path string) error {
	fi.mutex.Lock()
	defer fi.mutex.Unlock()
	return fi.addResource(path)
}

func (fi *FileIndexer) addResource(path string) error {
	if fi.IsSkippePath(path) {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fi.Index.Index(path, buildDoc(path, info))
	}
	fi.addDirResource(path)
	if fi.updateCallback != nil {
		go fi.updateCallback(fi)
	}
	return nil
}

func (fi *FileIndexer) addDirResource(path string) {

	batch := fi.Index.NewBatch()

	filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fi.Logger.Error(err)
			return nil
		}

		if fi.IsSkippePath(path) {
			return nil
		}

		if fi.enableWatch && info.IsDir() {
			fi.watchDir(path)
		}
		if path == fi.WatchedRootDir {
			return nil
		}
		err = batch.Index(path, buildDoc(path, info))
		if err != nil {
			fi.Logger.Error(err)
		}
		return nil
	})
	fi.Index.Batch(batch)
}

func buildDoc(path string, info os.FileInfo) FileDocument {
	doc := FileDocument{
		Name:       info.Name(),
		Path:       path,
		ParentPath: filepath.Dir(path),
		IsDir:      info.IsDir(),
	}
	return doc
}

func (fi *FileIndexer) printDocCount() (uint64, error) {
	docCount, err := fi.Index.DocCount()
	if err != nil {
		fi.Logger.Error(err)
		return 0, nil
	}
	fi.Logger.Debugf("当前文档总数：%v", docCount)

	return docCount, err
}

func (fi *FileIndexer) watchDir(path string) {
	if err := fi.watcher.Add(path); err != nil {
		fi.Logger.Error(err)
	}
}

func (fi *FileIndexer) processEvent(event fsnotify.Event) error {
	fi.mutex.Lock()
	defer fi.mutex.Unlock()
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		fi.addResource(event.Name)

	case event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename:
		fi.delResource(event.Name)

	default:
		return nil
	}
	go fi.printDocCount()
	return nil
}

func (fi *FileIndexer) StartWatching() {
	defer fi.watcher.Close()
	for {
		select {
		case event, ok := <-fi.watcher.Events:
			if !ok {
				return
			}
			if err := fi.processEvent(event); err != nil {
				fi.Logger.Error(err)
			}
		case err, ok := <-fi.watcher.Errors:
			if !ok {
				return
			}
			fi.Logger.Error(err)
		}
	}
}

func (fi *FileIndexer) initWatch() error {

	//初始化监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	fi.watcher = watcher

	//开启监听
	go fi.StartWatching()
	return nil
}
