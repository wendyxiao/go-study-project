package main

import (
	"go-study-project/internal/api"
	"go-study-project/internal/config"
	"go-study-project/internal/repository"
	"go-study-project/internal/service"
	"go-study-project/pkg/logger"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("/Users/wendybookmac/go-study-project/configs")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 2. 初始化Zap日志（核心步骤）
	zapLogger, err := logger.NewZapLogger(cfg.Log)
	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
	defer logger.Sync(zapLogger) // 程序退出时刷新日志缓冲区

	// 2. 初始化数据库
	db, err := repository.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("init db failed: %v", err)
	}
	repository.DB = db // 赋值给全局变量（实际项目用依赖注入）

	// 3. 初始化Gin引擎
	gin.SetMode(cfg.Server.Mode) // 设置模式（debug/release）
	r := gin.New()

	// 4. 注册中间件（日志、恢复、跨域等）
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	// 1. 创建基础路由组（路径前缀 /api/v1，无中间件）
	apiV1 := r.Group("/api/v1")

	userSvc := service.NewUserService(db, zapLogger, &cfg)
	// 5. 注册API路由
	api.RegisterUserRoutes(apiV1, userSvc, zapLogger) // 传递db给API层

	postsrv := service.NewPostService(db, zapLogger)
	api.RegisterPostoutes(apiV1, postsrv, zapLogger)

	commentRepo := service.NewCommentService(db)
	api.RegisterCommentoutes(apiV1, commentRepo, zapLogger)

	// 6. 启动服务
	log.Printf("server starting on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// corsMiddleware 跨域中间件（企业级项目必备）
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
