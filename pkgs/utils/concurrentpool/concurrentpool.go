package concurrentpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type AntsPool struct {
	*ants.Pool
	*sync.WaitGroup
	options    *AntsPoolOption
	isRoot     bool // 标记是否为最外层的父线程池
	mu         sync.Mutex
	ChildPools map[string]*ants.Pool
}

type AntsPoolOption struct {
	AntsOptions []ants.Option
	PoolSize    int
	ChildPoolId string
}

type AntsPoolOptionSetter func(*AntsPoolOption)

func WithAntsOptions(options ...ants.Option) AntsPoolOptionSetter {
	return func(opt *AntsPoolOption) {
		opt.AntsOptions = options
	}
}

func WithPoolSize(size int) AntsPoolOptionSetter {
	return func(opt *AntsPoolOption) {
		opt.PoolSize = size
	}
}

// 对于相同参数的ChildPoolId，将会共享一个基础ants.Pool池
// 用于在嵌套的协程池中限制子协程池并发
func WithChildPoolId(cpId string) AntsPoolOptionSetter {
	return func(opt *AntsPoolOption) {
		opt.ChildPoolId = cpId
	}
}

func NewAntsPool(setters ...AntsPoolOptionSetter) (*AntsPool, error) {

	options := &AntsPoolOption{}
	for _, setter := range setters {
		setter(options)
	}

	p, err := ants.NewPool(options.PoolSize, options.AntsOptions...)
	if err != nil {
		return nil, err
	}

	ap := &AntsPool{
		Pool:      p,
		WaitGroup: &sync.WaitGroup{},
		options:   options,
		isRoot:    true,
	}

	return ap, nil
}

func (ap *AntsPool) Submit(task func()) error {

	ap.Add(1)
	return ap.Pool.Submit(func() {
		defer ap.Done()
		task()
	})
}

func (ap *AntsPool) Release() {

	if ap.isRoot && ap.ChildPools != nil {
		for _, p := range ap.ChildPools {
			p.Release()
		}
	}
	ap.Pool.Release() // 然后释放当前线程池
}

// 基于现有的 AntsPool 实例分分裂出一个新的子协程池实例
// 如果不设置AntsPoolOption，则重用AntsPool的AntsPoolOption
// 如果设置 WithChildPoolId，对于相同参数的ChildPoolId，将会共享一个基础ants.Pool池
func (ap *AntsPool) ForkChildPool(setters ...AntsPoolOptionSetter) (*AntsPool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.ChildPools == nil {
		ap.ChildPools = make(map[string]*ants.Pool)
	}

	childOptions := &AntsPoolOption{
		PoolSize: ap.options.PoolSize, // 默认使用父线程池的大小
	}
	for _, setter := range setters {
		setter(childOptions)
	}

	pool, err := ap.retrieveOrCreatePool(childOptions)
	if err != nil {
		return nil, err
	}

	return &AntsPool{
		Pool:      pool,
		WaitGroup: new(sync.WaitGroup),
		options:   ap.options,
	}, nil
}

func (ap *AntsPool) retrieveOrCreatePool(options *AntsPoolOption) (*ants.Pool, error) {
	var pool *ants.Pool
	var ok bool
	// 当 ChildPoolId 为空时，总是创建新的 ants.Pool 实例
	if options.ChildPoolId != "" {
		pool, ok = ap.ChildPools[options.ChildPoolId]
	} else {
		options.ChildPoolId = fmt.Sprintf("pool-%v", time.Now().UnixNano())
	}
	if !ok {
		var err error
		pool, err = ants.NewPool(options.PoolSize, append(ap.options.AntsOptions, options.AntsOptions...)...)
		if err != nil {
			return nil, err
		}
		ap.ChildPools[options.ChildPoolId] = pool
	}
	return pool, nil
}
