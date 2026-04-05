package config

import (
	"time"

	"github.com/spf13/viper"
)

// ServerConfig 服务配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug/release
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	Log             LogConfig     `mapstructure:"log"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`       // 日志级别（debug/info/warn/error）
	Path       string `mapstructure:"path"`        // 日志文件路径（如./logs/app.log）
	MaxSize    int    `mapstructure:"max_size"`    // 单文件最大尺寸（MB）
	MaxBackups int    `mapstructure:"max_backups"` // 最大备份文件数
	MaxAge     int    `mapstructure:"max_age"`     // 文件保留天数（天）
	Compress   bool   `mapstructure:"compress"`    // 是否压缩备份文件
}

// AppConfig 全局配置
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
}

// LoadConfig 加载配置（Viper实现）
func LoadConfig(path string) (config AppConfig, err error) {
	viper.AddConfigPath(path)   // 配置文件目录
	viper.SetConfigName("app")  // 配置文件名（不含后缀）
	viper.SetConfigType("yaml") // 配置文件类型
	viper.AutomaticEnv()        // 自动读取环境变量（覆盖配置文件）

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		return config, err
	}

	// 解析到结构体
	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
