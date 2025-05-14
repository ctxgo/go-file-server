package casbin

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/persist/cache"
	gormadapter "github.com/casbin/gorm-adapter/v3" // 更新为 v3 版本
	"gorm.io/gorm"
)

// Initialize the model from a string.
const text = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, remark

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && \
 (keyMatch2(r.obj, p.obj) || keyMatch(r.obj, p.obj)) && \
(r.act == p.act || p.act == "*")
`

// CasbinConfig holds configuration for creating a new enforcer.
type CasbinConfig struct {
	Cache       cache.Cache
	DB          *gorm.DB
	ModelPath   string
	Watcher     persist.Watcher
	tablePrefix string
	tableName   string
}

// Option defines a function that configures CasbinConfig.
type Option func(*CasbinConfig)

// WithGormDB configures a Gorm DB connection for the Casbin enforcer.
func WithGormDB(db *gorm.DB) Option {
	return func(c *CasbinConfig) {
		c.DB = db
	}
}

// default ""
func WithCasbinTablePrefix(prefix string) Option {
	return func(c *CasbinConfig) {
		c.tablePrefix = prefix
	}
}

// default "casbin_rule"
func WithCasbinTableName(name string) Option {
	return func(c *CasbinConfig) {
		c.tableName = name
	}
}

// WithModelPath configures the path to the Casbin model configuration.
func WithModelPath(path string) Option {
	return func(c *CasbinConfig) {
		c.ModelPath = path
	}
}

// WithWatcher configures a Casbin watcher for real-time policy updates.
func WithWatcher(watcher persist.Watcher) Option {
	return func(c *CasbinConfig) {
		c.Watcher = watcher
	}
}

// WithCache configures a cache for Casbin decisions.
func WithCache(cache cache.Cache) Option {
	return func(c *CasbinConfig) {
		c.Cache = cache
	}
}

// NewEnforcer creates a new Casbin enforcer with provided options.
func NewEnforcer(opts ...Option) (*casbin.CachedEnforcer, error) {
	config := &CasbinConfig{
		tableName: "casbin_rule",
	}
	for _, opt := range opts {
		opt(config)
	}
	adapter, err := gormadapter.NewAdapterByDBUseTableName(
		config.DB,
		config.tablePrefix,
		config.tableName,
	)
	if err != nil {
		return nil, err
	}
	m, err := model.NewModelFromString(text)
	if err != nil {
		panic(err)
	}

	enforcer, err := casbin.NewCachedEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}
	if config.Watcher != nil {
		enforcer.SetWatcher(config.Watcher)
	}
	if config.Cache != nil {
		enforcer.SetCache(config.Cache)
	}
	return enforcer, nil
}
