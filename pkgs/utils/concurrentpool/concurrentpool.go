package concurrentpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

// SharedPoolRef stores goroutine pool information and reference count
type SharedPoolRef struct {
	pool *ants.Pool
	refs int
}

// AntsPool wraps ants.Pool, providing parent-child pool management and wait group functionality
type AntsPool struct {
	*ants.Pool
	*sync.WaitGroup
	options    *AntsPoolOption
	mu         sync.Mutex
	ChildPools map[string]*SharedPoolRef // child pool mapping table
	parentPool *AntsPool                 // parent pool reference
	poolId     string                    // identifier in parent pool
}

// AntsPoolOption defines pool configuration options
type AntsPoolOption struct {
	AntsOptions []ants.Option
	PoolSize    int
	ChildPoolId string
}

// AntsPoolOptionSetter function type for setting pool options
type AntsPoolOptionSetter func(*AntsPoolOption)

// WithAntsOptions sets options for ants.Pool
func WithAntsOptions(options ...ants.Option) AntsPoolOptionSetter {
	return func(opt *AntsPoolOption) {
		opt.AntsOptions = options
	}
}

// WithPoolSize sets the pool size
func WithPoolSize(size int) AntsPoolOptionSetter {
	return func(opt *AntsPoolOption) {
		opt.PoolSize = size
	}
}

// WithChildPoolId sets the child pool ID
func WithChildPoolId(cpId string) AntsPoolOptionSetter {
	return func(opt *AntsPoolOption) {
		opt.ChildPoolId = cpId
	}
}

// NewAntsPool creates a new goroutine pool
func NewAntsPool(setters ...AntsPoolOptionSetter) (*AntsPool, error) {
	options := &AntsPoolOption{}
	for _, setter := range setters {
		setter(options)
	}

	p, err := ants.NewPool(options.PoolSize, options.AntsOptions...)
	if err != nil {
		return nil, err
	}

	return &AntsPool{
		Pool:      p,
		WaitGroup: &sync.WaitGroup{},
		options:   options,
	}, nil
}

// AntsSubmit submits task to goroutine pool and automatically handles WaitGroup
func (ap *AntsPool) AntsSubmit(task func()) error {
	return ap.Pool.Submit(func() {
		defer ap.Done()
		task()
	})
}

// Submit submits task and increments WaitGroup counter
func (ap *AntsPool) Submit(task func()) error {
	ap.Add(1)
	return ap.AntsSubmit(task)
}

// getOrCreatePool gets existing pool or creates new pool
func (ap *AntsPool) getOrCreatePool(options *AntsPoolOption) (*SharedPoolRef, error) {
	// reuse existing pool
	if sharedPool, exists := ap.ChildPools[options.ChildPoolId]; exists {
		sharedPool.refs++
		return sharedPool, nil
	}

	// create new pool
	pool, err := ants.NewPool(options.PoolSize,
		append(ap.options.AntsOptions, options.AntsOptions...)...)
	if err != nil {
		return nil, err
	}

	sharedPool := &SharedPoolRef{
		pool: pool,
		refs: 1,
	}
	ap.ChildPools[options.ChildPoolId] = sharedPool
	return sharedPool, nil
}

// ForkChildPool creates or reuses child pool
func (ap *AntsPool) ForkChildPool(setters ...AntsPoolOptionSetter) (*AntsPool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.ChildPools == nil {
		ap.ChildPools = make(map[string]*SharedPoolRef)
	}

	childOptions := &AntsPoolOption{
		PoolSize: ap.options.PoolSize,
	}
	for _, setter := range setters {
		setter(childOptions)
	}

	if childOptions.ChildPoolId == "" {
		childOptions.ChildPoolId = fmt.Sprintf("pool-%v", time.Now().UnixNano())
	}

	sharedPool, err := ap.getOrCreatePool(childOptions)
	if err != nil {
		return nil, err
	}

	return &AntsPool{
		Pool:       sharedPool.pool,
		WaitGroup:  new(sync.WaitGroup),
		options:    childOptions,
		parentPool: ap,
		poolId:     childOptions.ChildPoolId,
	}, nil
}

// Release releases pool resources
func (ap *AntsPool) Release() {
	// if current pool is root pool, release directly
	if ap.parentPool == nil {
		ap.Pool.Release()
		return
	}
	// remove self from parent pool
	ap.parentPool.mu.Lock()
	if sharedPool, exists := ap.parentPool.ChildPools[ap.poolId]; exists {
		sharedPool.refs--
		if sharedPool.refs <= 0 {
			sharedPool.pool.Release()
			delete(ap.parentPool.ChildPools, ap.poolId)
		}
	}
	ap.parentPool.mu.Unlock()
}
