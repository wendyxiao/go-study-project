package logger

import (
	"fmt"
	"go-study-project/internal/config" // 替换为你的项目模块路径
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewZapLogger 初始化Zap日志实例（配置驱动）
func NewZapLogger(cfg config.LogConfig) (*zap.Logger, error) {
	// 1. 解析日志级别（如"info" -> zapcore.InfoLevel）
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level failed: %w", err)
	}

	// 2. 配置日志编码器（JSON格式，结构化输出）
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 级别小写（如"info"）
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // 时间格式（ISO8601）
		EncodeDuration: zapcore.SecondsDurationEncoder, // 耗时单位（秒）
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短路径调用者（如"main.go:123"）
	}

	// 3. 配置输出目标（控制台+文件，按环境切换）
	var cores []zapcore.Core
	// 控制台输出（开发环境默认开启）
	if cfg.Path != "" || isDevEnv() { // isDevEnv()判断是否为开发环境（如cfg.Server.Mode=="debug"）
		consoleCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig), // JSON编码器
			zapcore.AddSync(os.Stdout),            // 输出到控制台
			level,                                 // 日志级别
		)
		cores = append(cores, consoleCore)
	}
	// 文件输出（生产环境必选）
	if cfg.Path != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Path,       // 日志文件路径
			MaxSize:    cfg.MaxSize,    // 单文件最大尺寸（MB）
			MaxBackups: cfg.MaxBackups, // 最大备份文件数
			MaxAge:     cfg.MaxAge,     // 文件保留天数（天）
			Compress:   cfg.Compress,   // 是否压缩备份文件（gzip）
		}
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter), // 输出到文件
			level,
		)
		cores = append(cores, fileCore)
	}

	// 4. 合并核心（多输出目标）
	core := zapcore.NewTee(cores...)

	// 5. 创建Zap实例（添加调用者信息、堆栈跟踪）
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}

// parseLogLevel 解析日志级别（字符串->zapcore.Level）
func parseLogLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "dpanic":
		return zapcore.DPanicLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// isDevEnv 判断是否为开发环境（根据配置文件的server.mode）
func isDevEnv() bool {
	// 实际项目中，可通过Viper获取配置文件的server.mode（如"debug"）
	// 这里简化处理，默认返回true（开发环境）
	return true
}

// Sync 刷新日志缓冲区（程序退出时调用）
func Sync(logger *zap.Logger) {
	if logger != nil {
		_ = logger.Sync() // 忽略错误（如文件已关闭）
	}
}
