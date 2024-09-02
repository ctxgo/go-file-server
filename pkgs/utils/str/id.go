// snowflake/snowflake.go

package str

import (
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

const machineID uint16 = 1

type Option func(*sonyflake.Settings)

var sf *sonyflake.Sonyflake

func WithStartTime(t time.Time) Option {
	return func(o *sonyflake.Settings) {
		o.StartTime = t
	}
}

func WithMachineID(machineID uint16) Option {
	return func(o *sonyflake.Settings) {
		o.MachineID = func() (uint16, error) {
			return machineID, nil
		}
	}
}

func InitSnowflake(opts ...Option) {
	var settings sonyflake.Settings = sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return machineID, nil
		},
	}

	for _, opt := range opts {
		opt(&settings)
	}

	sf = sonyflake.NewSonyflake(settings)
	if sf == nil {
		panic("nil")
	}
}

func NextID() (uint64, error) {
	return sf.NextID()
}

func NextStrID() (string, error) {
	id, err := NextID()
	if err != nil {
		return "", err
	}
	strID := strconv.FormatUint(id, 10)
	return strID, nil
}
