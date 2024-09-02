package system

// SystemDetails 代表系统的整体信息
type SystemDetails struct {
	OS         string        `json:"os"`          // 操作系统类型
	Uptime     uint64        `json:"uptime"`      // 系统运行时间（秒）
	CPUInfo    CPUDetails    `json:"cpu_info"`    // CPU详情
	MemoryInfo MemoryDetails `json:"memory_info"` // 内存详情
	DiskInfo   DiskDetails   `json:"disk_info"`   // 磁盘详情
}

// CPUDetails 代表CPU的相关信息
type CPUDetails struct {
	UsagePercent float64 `json:"usage_percent"` // 使用百分比
}

// MemoryDetails 代表内存的相关信息
type MemoryDetails struct {
	Total uint64 `json:"total"` // 总内存（字节）
	Used  uint64 `json:"used"`  // 已用内存（字节）
	Free  uint64 `json:"free"`  // 空闲内存（字节）
}

// DiskDetails 代表磁盘的相关信息
type DiskDetails struct {
	TotalSize    uint64  `json:"total_size"`    // 总大小（字节）
	UsedSize     uint64  `json:"used_size"`     // 已用空间（字节）
	FreeSize     uint64  `json:"free_size"`     // 可用空间（字节）
	UsagePercent float64 `json:"usage_percent"` // 使用百分比
}
