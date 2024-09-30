package repository

import (
	"go-file-server/pkgs/pathtool"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/pkg/errors"
)

type FileDocument struct {
	Name  string
	Path  string
	IsDir bool
}

var ErrDocumentNotFound = errors.New("document not found")

type FsScope func(*bleve.SearchRequest)

type FsRepository struct {
	Indexer *pathtool.FileIndexer
	sync.RWMutex
}

func NewFsRepository(indexer *pathtool.FileIndexer) *FsRepository {
	return &FsRepository{Indexer: indexer}
}

// WithPagination 配置查询的分页
func WithPagination(index, size int) FsScope {
	return func(sr *bleve.SearchRequest) {
		offset := (index - 1) * size
		if offset < 0 {
			offset = 0
		}
		if size <= 0 {
			size = 10
		}
		sr.From = offset
		sr.Size = size
	}
}

// WithPathPrefix 查询特定路径前缀
func WithPrefixPath(path string) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewPrefixQuery(path)
		q.SetField("Path")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// WithParentPathPrefix 查询上级路径路径前缀
func WithParentPathPrefix(path string) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewPrefixQuery(path)
		q.SetField("ParentPath")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// WithParentPath 查询上级路径
func WithTermParentPath(path string) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewTermQuery(path)
		q.SetField("ParentPath")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// WithMatchName 匹配包含name(分词模式)
func WithMatchName(name string) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewMatchQuery(name)
		q.SetField("Name")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// WithRegexpName 正则匹配，简单正则比WithWildcardQuery效率更高
func WithRegexpName(name string) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewRegexpQuery(name)
		q.SetField("Name")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// WithMatchName 通配符
func WithWildcardQuery(name string) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewWildcardQuery(name)
		q.SetField("Name")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// WithMatchPhraseName 精确匹配
func WithMatchPhraseName(name string) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewMatchPhraseQuery(name)
		q.SetField("Name")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// WithIsDir 查询文件夹或文件
func WithIsDir(isDir bool) FsScope {
	return func(sr *bleve.SearchRequest) {
		q := bleve.NewBoolFieldQuery(isDir)
		q.SetField("IsDir")
		sr.Query = combineQueries(sr.Query, q)
	}
}

// combineQueries helps to combine multiple queries into a single conjunction query
func combineQueries(existing query.Query, newQuery query.Query) query.Query {
	if existing == nil {
		return newQuery
	}
	conjunctionQuery, ok := existing.(*query.ConjunctionQuery)
	if !ok {
		conjunctionQuery = query.NewConjunctionQuery([]query.Query{existing})
	}
	conjunctionQuery.AddQuery(newQuery)
	return conjunctionQuery
}

func (r *FsRepository) GetCount(scopes ...FsScope) (uint64, error) {
	r.RLock()
	defer r.RUnlock()
	if len(scopes) == 0 {
		return r.Indexer.Index.DocCount()
	}
	searchRequest := makeSearchRequest(scopes...)
	searchRequest.Size = 0
	results, err := r.Indexer.Index.Search(searchRequest)
	if err != nil {
		return 0, err
	}
	return results.Total, nil
}

func makeSearchRequest(scopes ...FsScope) *bleve.SearchRequest {

	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery())
	for _, scope := range scopes {
		scope(searchRequest)
	}
	return searchRequest
}

func (r *FsRepository) find(scopes ...FsScope) ([]FileDocument, uint64, error) {
	if len(scopes) == 0 {
		return []FileDocument{}, 0, nil
	}

	searchRequest := makeSearchRequest(scopes...)
	searchRequest.Fields = []string{"*"} // 请求加载所有字段

	results, err := r.Indexer.Search(searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var docs []FileDocument
	for _, hit := range results.Hits {
		doc := FileDocument{
			Name:  hit.Fields["Name"].(string),
			Path:  hit.Fields["Path"].(string),
			IsDir: hit.Fields["IsDir"].(bool),
		}
		docs = append(docs, doc)

	}
	return docs, results.Total, nil
}

func (r *FsRepository) Find(scopes ...FsScope) ([]FileDocument, uint64, error) {
	r.RLock()
	defer r.RUnlock()
	return r.find(scopes...)
}

func (r *FsRepository) FindOne(scopes ...FsScope) (FileDocument, error) {
	r.RLock()
	defer r.RUnlock()
	docs, _, err := r.find(append(scopes, WithPagination(0, 1))...)
	if err != nil {
		return FileDocument{}, err
	}
	if len(docs) > 0 {
		return docs[0], nil
	}
	return FileDocument{}, ErrDocumentNotFound
}

func (r *FsRepository) RemoveAll(path string) error {
	r.Lock()
	defer r.Unlock()
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	doc, _, err := r.find(WithPrefixPath(path))
	if err != nil {
		return err
	}
	for _, f := range doc {
		err = r.Indexer.DelResource(f.Path)
		if err != nil {
			return errors.Errorf(
				"removeAll index faild, path: %s , Subpath: %s , err: %v", path, f.Path, err)
		}
	}
	return nil
}

func (r *FsRepository) Remove(path string) error {
	r.Lock()
	defer r.Unlock()

	err := os.Remove(path)
	if err != nil {
		return err
	}

	return r.Indexer.DelResource(path)
}

func (r *FsRepository) Rename(src, des string) error {
	r.Lock()
	defer r.Unlock()
	if src == des {
		return nil
	}
	exist, err := pathtool.NewFiletool(des).IsExist()
	if err != nil {
		return err
	}
	if exist {
		return os.ErrExist
	}
	err = os.Rename(src, des)
	if err != nil {
		return err
	}
	err = r.Indexer.AddResource(des)
	if err != nil {
		return err
	}
	return r.Indexer.DelResource(src)

}

func (r *FsRepository) Mkdir(path string, perm fs.FileMode) error {
	r.Lock()
	defer r.Unlock()
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		return err
	}
	return r.Indexer.AddResource(path)
}

func findFirstNonExistentDir(path string) string {
	dir := filepath.Dir(path)
	if dir == path {
		return path
	}

	if _, err := os.Stat(dir); err == nil {
		return path
	}

	return findFirstNonExistentDir(dir)
}

func (r *FsRepository) MkdirAll(path string, perm fs.FileMode) error {
	r.Lock()
	defer r.Unlock()
	topPath := findFirstNonExistentDir(path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return r.Indexer.AddResource(topPath)
}

func (r *FsRepository) Create(path string) (*os.File, error) {
	r.Lock()
	defer r.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return f, r.Indexer.AddResource(path)
}

func (r *FsRepository) OpenFile(path string, flag int, perm os.FileMode) (f *os.File, err error) {
	r.Lock()
	defer r.Unlock()
	_, err = os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		defer func() {
			if err != nil {
				return
			}
			err = r.Indexer.AddResource(path)
		}()
	}

	f, err = os.OpenFile(path, flag, perm)
	return
}

func (r *FsRepository) AddResource(path string) error {
	r.Lock()
	defer r.Unlock()
	return r.Indexer.AddResource(path)
}

func (r *FsRepository) ResetIndex() error {
	r.Lock()
	defer r.Unlock()
	return r.Indexer.IndexInit()
}
