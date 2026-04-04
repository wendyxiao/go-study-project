package repository

import (
	"fmt"
	"go-study-project/internal/config"
	"go-study-project/internal/model"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局数据库实例（实际项目中可通过依赖注入优化）

// InitDB 初始化数据库连接（基于Viper配置）
func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		SkipDefaultTransaction: true, // 高性能场景关闭默认事务
		PrepareStmt:            true, // 预编译SQL提升性能
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取原生SQL DB对象，配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(cfg.ConnMaxLifetime))

	// 自动迁移模型（生产环境建议手动执行SQL）
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	return db, nil
}
