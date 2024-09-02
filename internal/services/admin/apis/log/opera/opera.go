package opera

import "go-file-server/internal/common/repository"

type OperaAPI struct {
	operarepo *repository.OperaLogRepository
}

func NewOperaAPI(
	operarepo *repository.OperaLogRepository,

) *OperaAPI {
	return &OperaAPI{
		operarepo: operarepo,
	}
}
