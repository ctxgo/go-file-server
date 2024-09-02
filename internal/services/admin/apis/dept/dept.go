package dept

import (
	"go-file-server/internal/common/repository"
)

type DeptApi struct {
	deptRepo *repository.DeptRepository
}

func NewDeptApi(
	deptRepo *repository.DeptRepository,
) *DeptApi {
	return &DeptApi{
		deptRepo: deptRepo,
	}
}
