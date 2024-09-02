package utils

import (
	"go-file-server/pkgs/config"
	"os"
	"path/filepath"
	"regexp"
	"syscall"

	"github.com/pkg/errors"
)

func CheckFsName(s string) error {
	// regexp.MustCompile(`^[a-zA-Z0-9-_\.` + `\x4e00-\x9fa5` + `]+$`).MatchString(s)
	if regexp.MustCompile(`^[a-zA-Z0-9-_\.\p{Han}]+$`).MatchString(s) {

		return nil
	}
	return errors.Errorf("文件名不符合规则，只能以字母数字中下划线点组合")
}

func GetTmpDir() string {
	return filepath.Join(config.ApplicationCfg.Basedir, ".tmp")
}

func GetRealPath(paths ...string) (string, error) {
	realPath := config.ApplicationCfg.Basedir
	paths = append([]string{realPath}, paths...)
	return SafeJoinPath(paths...)
}

func SafeJoinPath(paths ...string) (string, error) {
	p := ""
	for _, path := range paths {
		if regexp.MustCompile(`\s|\.\.`).MatchString(path) {
			return "", errors.Errorf("路径不符合规则，不能出现空格或者连续的点")
		}
		p = filepath.Join(p, filepath.Clean(path))
	}
	return p, nil
}

func ParsePathErr(err error) (bool, error) {

	switch {
	case errors.Is(err, syscall.ENOENT):
		return true, errors.Errorf("路径不存在")
	case errors.Is(err, syscall.ENOTEMPTY):
		return true, errors.Errorf("目录非空")
	case errors.Is(err, syscall.EACCES):
		return true, errors.Errorf("无权限访问")
	case errors.Is(err, os.ErrExist):
		return true, errors.Errorf("路径已经存在")
	default:
		return false, err
	}

}
