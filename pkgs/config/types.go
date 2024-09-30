package config

// Config 配置集合
type Config struct {
	Application *Application `mapstructure:"application"`
	Logger      *Logger      `mapstructure:"logger"`
	Jwt         *Jwt         `mapstructure:"jwt"`
	Database    *Database    `mapstructure:"database"`
	Cache       *Cache       `mapstructure:"cache"`
	OAuth       *OAuth       `mapstructure:"oauth"`
	Ftp         *Ftp         `mapstructure:"ftp"`
}

type Application struct {
	Host    string
	Port    string
	Basedir string
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
	Addr     string `mapstructure:"addr" json:"addr"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}

type OAuthGrpc struct {
	Addr    string `mapstructure:"addr"`
	TlsCert string `mapstructure:"tlsCert"`
	TlsKey  string `mapstructure:"tlsKey"`
	TlsCA   string `mapstructure:"tlsCA"`
}

type OAuth struct {
	Enable       bool      `mapstructure:"enable"`
	ClientID     string    `mapstructure:"clientID"`
	ClientSecret string    `mapstructure:"clientSecret"`
	RedirectUrl  string    `mapstructure:"redirectUrl"`
	IssuerUrl    string    `mapstructure:"issuerUrl"`
	State        string    `mapstructure:"state"`
	Scopes       []string  `mapstructure:"scopes"`
	Grpc         OAuthGrpc `mapstructure:"grpc"`
}

type Ftp struct {
	Enable           bool   `mapstructure:"enable"`
	Addr             string `mapstructure:"addr"`
	PublicHost       string `mapstructure:"publicHost"`
	PassivePortStart int    `mapstructure:"passivePortStart"`
	PassivePortEnd   int    `mapstructure:"passivePortEnd"`
}
