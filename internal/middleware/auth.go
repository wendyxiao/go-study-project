package middleware

import (
	"go-study-project/internal/config"
	"go-study-project/internal/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware JWT认证中间件（验证令牌有效性）
func AuthMiddleware(cfg *config.JWT, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未提供令牌"})
			c.Abort()
			return
		}

		// 2. 提取Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "令牌格式错误"})
			c.Abort()
			return
		}
		tokenStr := parts[1]

		// 3. 解析JWT令牌
		claims := &model.JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil // 密钥从配置读取
		})

		// 4. 验证令牌有效性
		if err != nil || !token.Valid {
			logger.Warn("JWT验证失败", zap.Error(err), zap.String("token", tokenStr))
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "令牌无效或已过期"})
			c.Abort()
			return
		}

		// 5. 将用户信息存入上下文（供后续接口使用）
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminOnly 管理员权限中间件（需配合AuthMiddleware使用）
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "权限不足（仅管理员可访问）"})
			c.Abort()
			return
		}
		c.Next()
	}
}
