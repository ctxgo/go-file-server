package ftpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"go-file-server/internal/common/middlewares"
	"go-file-server/internal/common/repository"
	fsApi "go-file-server/internal/services/admin/apis/fs"
	"go-file-server/internal/services/admin/apis/fs/utils"
	"go-file-server/internal/services/admin/apis/role"
	"go-file-server/pkgs/cache"
	"go-file-server/pkgs/utils/limiter"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	Read   = "GET"
	Write  = "POST"
	Update = "PUT"
	Delet  = "DELETE"
)

type FileServerFs struct {
	token          string
	user           string
	roleKey        string
	cache          cache.AdapterCache
	roleRepo       *repository.RoleRepository
	fsRepo         *repository.FsRepository
	casbinEnforcer *casbin.CachedEnforcer
	limiterManager *utils.LimiterManager
}

func (f *FileServerFs) VerifPath(name string, action string) (string, error) {
	if f.roleKey == "admin" {
		return utils.GetRealPath(name)
	}
	if name != "/" {
		name = strings.TrimLeft(name, "/")
	}
	homePath, err := decryptPath(name)
	if err != nil {
		return "", err
	}
	requestParh := "/api/v1/fs" + homePath
	err = middlewares.CasbinEnforce(f.casbinEnforcer, f.roleKey, requestParh, action)
	if err != nil {
		return "", err
	}
	return utils.GetRealPath(homePath)
}

func (f *FileServerFs) Chmod(name string, mode fs.FileMode) error {
	path, err := f.VerifPath(name, Update)
	if err != nil {
		return err
	}
	return os.Chmod(path, mode)
}

func (f *FileServerFs) Chown(name string, uid int, gid int) error {
	path, err := f.VerifPath(name, Update)
	if err != nil {
		return err
	}
	return os.Chown(path, uid, gid)

}

func (f *FileServerFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	path, err := f.VerifPath(name, Update)
	if err != nil {
		return err
	}
	return os.Chtimes(path, atime, mtime)

}

func (f *FileServerFs) Create(name string) (afero.File, error) {
	path, err := f.VerifPath(name, Write)
	if err != nil {
		return nil, err
	}
	return f.fsRepo.Create(path)
}

// f.Mkdir 内部 CreateDir 会调用 os.MkdirAll
// 对于多层级路径创建，外部会逐层调用 f.Mkdir
func (f *FileServerFs) Mkdir(name string, perm fs.FileMode) error {
	path, err := f.VerifPath(name, Write)
	if err != nil {
		return err
	}
	return f.fsRepo.Mkdir(path, perm)
}

func (f *FileServerFs) MkdirAll(path string, perm fs.FileMode) error {
	path, err := f.VerifPath(path, Write)
	if err != nil {
		return err
	}
	return f.fsRepo.MkdirAll(path, perm)
}

func (f *FileServerFs) Name() string {
	return "FileServerFs"
}

func (f *FileServerFs) Open(name string) (afero.File, error) {
	path, err := f.VerifPath(name, Read)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (f *FileServerFs) ReadDir(name string) ([]os.FileInfo, error) {
	if strings.HasPrefix(name, "/.tmp") {
		_, err := f.ensureTempDir()
		if err != nil {
			return nil, err
		}
	}

	if name != "/" || f.roleKey == "admin" {
		realPath, err := f.VerifPath(name, Read)
		if err != nil {
			return nil, err
		}
		return f.listPath(realPath)
	}

	return f.listNormalPath()

}

func (f *FileServerFs) getLimiter() (*limiter.Limiter, error) {

	raleLimiteBytes, err := fsApi.GetRaleLimiteBytes(f.roleKey, f.cache, f.roleRepo)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s-%s", f.user, f.roleKey)
	raleLimiter := f.limiterManager.GetLimiter(key, raleLimiteBytes)
	return raleLimiter, nil

}

func (f *FileServerFs) OpenFile(name string, flag int, perm fs.FileMode) (afero.File, error) {

	path, err := f.VerifPath(name, Read)
	if err != nil {
		return nil, err
	}

	raleLimiter, err := f.getLimiter()
	if err != nil {
		return nil, err
	}

	file, err := f.fsRepo.OpenFile(path, flag, perm)
	if err != nil {
		return nil, err
	}

	nf := &File{
		File:       file,
		ReadWriter: raleLimiter.LimitReadertWriter(context.Background(), file),
	}

	return nf, nil

}

func (f *FileServerFs) Remove(path string) error {
	return f.execRemove(path, f.fsRepo.Remove)
}

func (f *FileServerFs) RemoveAll(path string) error {
	return f.execRemove(path, f.fsRepo.RemoveAll)
}

func (f *FileServerFs) execRemove(path string, removeFunc func(string) error) error {

	realPath, err := f.VerifPath(path, Delet)
	if err != nil {
		return err
	}

	// 直接删除
	if strings.HasPrefix(path, "/.tmp") {
		return removeFunc(realPath)
	}

	tmpDir, err := f.ensureTempDir()
	if err != nil {
		return err
	}

	// 转移到回收站
	desPath := filepath.Join(tmpDir, filepath.Base(path)+"_"+utils.GetTimeStr())

	return f.fsRepo.Rename(realPath, desPath)

}

func (f *FileServerFs) Rename(oldname string, newname string) error {
	oldname, err := f.VerifPath(oldname, Delet)
	if err != nil {
		return err
	}
	newname, err = f.VerifPath(newname, Write)
	if err != nil {
		return err
	}
	return f.fsRepo.Rename(oldname, newname)

}

func (f *FileServerFs) Stat(name string) (fs.FileInfo, error) {
	path, err := f.VerifPath(name, Read)
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}

func (f *FileServerFs) ensureTempDir() (string, error) {
	tempPath, err := fsApi.EnsureTempDir(f.roleKey)
	if err != nil {
		return "", err
	}
	return tempPath, f.fsRepo.MkdirAll(tempPath, os.ModePerm)
}

func (f *FileServerFs) listPath(path string) ([]os.FileInfo, error) {

	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (f *FileServerFs) listNormalPath() ([]os.FileInfo, error) {
	var files []os.FileInfo
	policies := f.casbinEnforcer.GetFilteredPolicy(0, f.roleKey, "", "GET", "fs")
	if len(policies) == 0 {
		return nil, errors.Errorf("无任何目录权限，请联系管理员赋权")
	}
	for _, p := range policies {

		fsPath := role.ParseFsRolepath(p[1])

		if strings.HasPrefix(fsPath, "/.tmp") {
			_, err := f.ensureTempDir()
			if err != nil {
				return nil, err
			}
		}

		realPath, err := utils.GetRealPath(fsPath)
		if err != nil {
			return nil, err
		}
		fileInfo, err := os.Stat(realPath)
		if err != nil {
			return nil, err
		}
		files = append(files, &FileInfo{
			FileInfo: fileInfo,
			fullName: encryptPath(fsPath),
		})
	}
	return files, nil
}

// 加密逻辑：Base64URL编码路径的前面部分，最后一段保留
func encryptPath(path string) string {
	path = filepath.Clean(path)
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}
	lastPart := parts[len(parts)-1]
	remainingPath := strings.Join(parts[:len(parts)-1], "/")
	encoded := base64.URLEncoding.EncodeToString([]byte(remainingPath))
	encoded = strings.TrimRight(encoded, "=")
	return encoded + "-" + lastPart
}

// 解密逻辑：还原加密路径
func decryptPath(encryptedPath string) (string, error) {
	// 根据 "-" 拆分出加密部分和最后一段
	parts := strings.Split(encryptedPath, "-")
	if len(parts) < 2 {
		return encryptedPath, nil
	}
	encodedPart := parts[0]
	lastPart := parts[1]

	// 还原 Base64URL 编码的部分，补充丢失的 "="
	missingPadding := len(encodedPart) % 4
	if missingPadding > 0 {
		encodedPart += strings.Repeat("=", 4-missingPadding)
	}

	decodedBytes, err := base64.URLEncoding.DecodeString(encodedPart)
	if err != nil {
		return "", fmt.Errorf("error decoding Base64URL: %v", err)
	}

	return string(decodedBytes) + "/" + lastPart, nil
}