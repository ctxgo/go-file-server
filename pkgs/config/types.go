package config

// Config 配置集合
type Config struct {
	Application *Application `yaml:"application"`
	Logger      *Logger      `yaml:"logger"`
	Jwt         *Jwt         `yaml:"jwt"`
	Database    *Database    `yaml:"database"`
	Cache       *Cache       `yaml:"cache"`
}

type Application struct {
	Host      string
	Port      string
	JwtSecret string
	Basedir   string
}

type Logger struct {
	Path  string
	Level string
}

type Jwt struct {
	Secret  string
	Timeout int64
}

type Database struct {
	Driver string
	Source string
}
type Cache struct {
	Redis  *RedisConnectOptions
	Memory interface{}
}
type RedisConnectOptions struct {
	Addr     string `yaml:"addr" json:"addr"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}
