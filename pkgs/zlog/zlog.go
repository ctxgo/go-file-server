package zlog

import (
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var SugLog *zap.SugaredLogger
var Log *zap.Logger

func Init(_opts ...Option) {
	//创建核心对象
	var coreArr []zapcore.Core

	options := getOpts(_opts...)

	zapEncoder := getLogEncoder()

	level, lowPriority, highPriority := getLogLevel(options.level)

	//info文件writeSyncer
	infoFileCore := zapcore.NewCore(zapEncoder,
		zapcore.NewMultiWriteSyncer(getLogWriter(options, "log"),
			zapcore.AddSync(os.Stdout)), lowPriority)
	//error文件writeSyncer
	errorFileCore := zapcore.NewCore(zapEncoder,
		zapcore.NewMultiWriteSyncer(getLogWriter(options, "err"),
			zapcore.AddSync(os.Stdout)), highPriority)

	coreArr = append(coreArr, infoFileCore, errorFileCore)
	//zap.AddCaller()为显示文件名和行号，可省略
	Log = zap.New(zapcore.NewTee(coreArr...), zap.AddCaller())

	SugLog = Log.Sugar()
	//日志
	SugLog.Infof("  **********日志初始化完成 输出级别=[%v]**********", level)

}

func getOpts(opts ...Option) options {
	defaultOpts := setDefault()
	for _, f := range opts {
		f(&defaultOpts)
	}
	return defaultOpts
}

func getLogEncoder() zapcore.Encoder {
	//获取编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = timeEncoder
	//encoderConfig.CallerKey=""
	//按级别显示不同颜色，不需要的话取值zapcore.CapitalLevelEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	//显示完整文件路径
	//encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	//NewJSONEncoder输出json格式，NewConsoleEncoder输出普通文本格式
	return zapcore.NewConsoleEncoder(encoderConfig)

}

// 格式获取当前日志级别
func getLogLevel(loglevel string) (zapcore.Level, zap.LevelEnablerFunc, zap.LevelEnablerFunc) {
	var level zapcore.Level = zap.InfoLevel
	var err error
	if loglevel != "" {
		level, err = zapcore.ParseLevel(loglevel)
		if err != nil {
			panic(err)
		}
	}
	//info和debug级别,debug级别是最低的
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= level && lev < zap.ErrorLevel
	})
	//error级别
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	return level, lowPriority, highPriority
}

// 自定义时间编码器
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func getLogWriter(opts options, prefix string) zapcore.WriteSyncer {
	// 检查日志目录是否存在，如果不存在则尝试创建
	if _, err := os.Stat(opts.path); os.IsNotExist(err) {
		// 尝试创建目录
		if err := os.MkdirAll(opts.path, 0755); err != nil {
			panic(err)
		}
	}

	//普通日志输出
	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(opts.path, prefix), //日志文件存放目录，如果文件夹不存在会自动创建
		MaxSize:    opts.maxSize,                     //文件大小限制,单位MB
		MaxBackups: opts.maxBackups,                  //最大保留日志文件数量
		MaxAge:     opts.maxAge,                      //日志文件保留天数
		Compress:   opts.compress,                    //是否压缩处理
	})
	//返回
	return writeSyncer
}
