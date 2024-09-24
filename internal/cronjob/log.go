package cronjob

import "go.uber.org/zap"

type CornLogger struct {
	logger *zap.SugaredLogger
}

func (zl *CornLogger) Printf(format string, args ...interface{}) {
	zl.logger.Infof(format, args...)
}
