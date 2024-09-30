package config

import (
	"log"
	"runtime"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Viper struct{ *viper.Viper }

// viperConfig 全局配置变量
var (
	ApplicationCfg = new(Application)
	DatabaseCfg    = new(Database)
	LoggerCfg      = new(Logger)
	JwtCfg         = new(Jwt)
	CacheCfg       = new(Cache)
	OAuthCfg       = new(OAuth)
	FptCfg         = new(Ftp)
)

func init() {
	//完整的日志记录器在config之后才能进行初始化
	//这里我们先用log来记录日志
	log.SetFlags(log.LstdFlags) //设置日志行号
}

func Init(opts ...func(*Viper)) {
	config := &Viper{viper.GetViper()}
	config.setup()
	for _, opt := range opts {
		opt(config)
	}
	err := config.ReadInConfig()
	checkError(err)

	err = config.Unmarshal(getSettings())
	checkError(err)
}

func getSettings() *Config {
	return &Config{
		Application: ApplicationCfg,
		Logger:      LoggerCfg,
		Jwt:         JwtCfg,
		Database:    DatabaseCfg,
		Cache:       CacheCfg,
		OAuth:       OAuthCfg,
		Ftp:         FptCfg,
	}

}

func (c *Viper) setup() {
	c.SetConfigFile("./config.yaml")
	c.SetConfigType("yaml")
}

func SetFile(f string) func(viper *Viper) {
	return func(viper *Viper) {
		viper.SetConfigFile(f)
	}
}
func SetConfigType(t string) func(viper *Viper) {
	return func(viper *Viper) {
		viper.SetConfigType(t)
	}
}

func SetEnvPrefix(Prefix string) func(viper *Viper) {
	return func(viper *Viper) {
		viper.SetEnvPrefix(Prefix)
	}
}

func SetAutomaticEnv() func(viper *Viper) {
	return func(viper *Viper) {
		viper.AutomaticEnv()
	}

}

func Setflags(flags *pflag.FlagSet) func(viper *Viper) {
	return func(viper *Viper) {
		err := viper.BindPFlags(flags)
		checkError(err)
	}
}

func checkError(err error) {
	_, file, line, _ := runtime.Caller(1)
	if err != nil {
		log.Fatalf("%v: %v, %v", file, line, err)
	}
}
