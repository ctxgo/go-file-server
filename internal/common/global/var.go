package global

import (
	syncpool "go-file-server/pkgs/utils/sync-pool"
)

var (
	BufferPool = syncpool.NewBufferPool()
)
