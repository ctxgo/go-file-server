package zlog

type Option func(*options)

type options struct {
	path       string
	level      string
	maxSize    int  //文件大小限制,单位MB
	maxBackups int  //最大保留日志文件数量
	maxAge     int  //日志文件保留天数
	compress   bool //是否压缩处理
}

func setDefault() options {
	return options{
		path:       "./logs",
		level:      "info",
		maxSize:    50,    //文件大小限制,单位MB
		maxBackups: 10,    //最大保留日志文件数量
		maxAge:     7,     //日志文件保留天数
		compress:   false, //是否压缩处理
	}
}
func WithPath(s string) Option {
	return func(o *options) {
		o.path = s
	}
}

func WithLevel(s string) Option {
	return func(o *options) {
		o.level = s
	}
}
func WithMaxSize(s int) Option {
	return func(o *options) {
		o.maxSize = s
	}
}
func WithMaxBackups(s int) Option {
	return func(o *options) {
		o.maxBackups = s
	}
}
func WithMaxAge(s int) Option {
	return func(o *options) {
		o.maxAge = s
	}
}

func WithCompress(s bool) Option {
	return func(o *options) {
		o.compress = s
	}
}
