package api

import (
	"go-study-project/internal/config"
	"go-study-project/internal/middleware"
	"go-study-project/internal/model"
	"go-study-project/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterUserRoutes 注册用户相关路由（路由分组管理）
func RegisterUserRoutes(r *gin.RouterGroup, userSvc service.UserService, logger *zap.Logger, jwt *config.JWT) {
	// 公开接口（无需认证）
	r.POST("/users/register", createUserHandler(userSvc, logger)) // 用户注册
	r.POST("/users/login", loginHandler(userSvc, logger))         // 用户登录

	// 需要认证的接口（需从Token中解析用户ID作为operatorID）
	authGroup := r.Group("/users")
	authGroup.Use(middleware.AuthMiddleware(jwt, logger)) // 认证中间件（示例，需自行实现）
	{
		authGroup.GET("/:id", getUserHandler(userSvc, logger))       // 获取用户详情
		authGroup.PUT("/:id", updateUserHandler(userSvc, logger))    // 更新用户
		authGroup.DELETE("/:id", deleteUserHandler(userSvc, logger)) // 删除用户
	}

}

// createUserHandler 创建用户（注册）
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.User true "用户信息（不含ID）"
// @Success 201 {object} map[string]interface{} "注册成功，返回用户ID"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /users/register [post]
func createUserHandler(svc service.UserService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 绑定请求参数（自动校验JSON格式）
		var req model.User
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("注册请求参数错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误", "details": err.Error()})
			return
		}

		// 2. 调用服务层注册逻辑
		userID, err := svc.Register(&req)
		if err != nil {
			logger.Error("注册失败", zap.Error(err), zap.String("username", req.Username))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 3. 返回成功响应（HTTP 201 Created）
		c.JSON(http.StatusCreated, gin.H{
			"code":    0,
			"message": "注册成功",
			"data":    gin.H{"user_id": userID},
		})
	}
}

// loginHandler 用户登录
// @Summary 用户登录
// @Description 验证用户名密码，返回访问令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param credentials body struct{Username string `json:"username"`; Password string `json:"password"`} true "登录凭证"
// @Success 200 {object} map[string]interface{} "登录成功，返回Token"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "用户名或密码错误"
// @Router /users/login [post]
func loginHandler(svc service.UserService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 绑定登录凭证（自定义结构体，避免绑定到model.User）
		var cred struct {
			Username string `json:"username" binding:"required,min=3,max=50"`
			Password string `json:"password" binding:"required,min=6"`
		}
		if err := c.ShouldBindJSON(&cred); err != nil {
			logger.Warn("登录请求参数错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
			return
		}

		// 2. 调用服务层登录逻辑
		token, err := svc.Login(cred.Username, cred.Password)
		if err != nil {
			logger.Warn("登录失败", zap.String("username", cred.Username), zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// 3. 返回Token响应
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "登录成功",
			"data":    gin.H{"token": token},
		})
	}
}

// getUserHandler 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID查询用户信息（需本人或管理员权限）
// @Tags 用户管理
// @Produce json
// @Param id path uint true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "用户信息"
// @Failure 403 {object} map[string]string "无权限访问"
// @Failure 404 {object} map[string]string "用户不存在"
// @Router /users/{id} [get]
func getUserHandler(svc service.UserService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取路径参数ID
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logger.Warn("无效的用户ID", zap.String("id", idStr))
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}

		// 2. 从上下文中获取操作者ID（认证中间件设置）
		operatorID, exists := c.Get("user_id")
		if !exists {
			logger.Error("无法获取操作者ID")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			return
		}

		// 3. 调用服务层查询逻辑
		user, err := svc.GetUserByID(uint(id), operatorID.(uint))
		if err != nil {
			logger.Warn("获取用户失败", zap.Uint64("id", id), zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// 4. 过滤敏感字段（如密码）
		user.Password = ""
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data":    user,
		})
	}
}

// updateUserHandler 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息（需本人或管理员权限）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户ID"
// @Param user body model.User true "更新后的用户信息"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "更新成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 403 {object} map[string]string "无权限操作"
// @Router /users/{id} [put]
func updateUserHandler(svc service.UserService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取路径参数ID
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logger.Warn("无效的用户ID", zap.String("id", idStr))
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}

		// 2. 绑定更新数据
		var updateData model.User
		if err := c.ShouldBindJSON(&updateData); err != nil {
			logger.Warn("更新请求参数错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误"})
			return
		}
		updateData.ID = uint(id) // 确保ID与路径参数一致

		// 3. 获取操作者ID（认证中间件设置）
		operatorID, exists := c.Get("userID")
		if !exists {
			logger.Error("无法获取操作者ID")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			return
		}

		// 4. 调用服务层更新逻辑
		if err := svc.UpdateUser(&updateData, operatorID.(uint)); err != nil {
			logger.Error("更新用户失败", zap.Uint("id", uint(id)), zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
	}
}

// deleteUserHandler 删除用户（软删除）
// @Summary 删除用户
// @Description 删除指定用户（需本人或管理员权限）
// @Tags 用户管理
// @Produce json
// @Param id path uint true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "删除成功"
// @Failure 403 {object} map[string]string "无权限操作"
// @Failure 404 {object} map[string]string "用户不存在"
// @Router /users/{id} [delete]
func deleteUserHandler(svc service.UserService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取路径参数ID
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			logger.Warn("无效的用户ID", zap.String("id", idStr))
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}

		// 2. 获取操作者ID（认证中间件设置）
		operatorID, exists := c.Get("userID")
		if !exists {
			logger.Error("无法获取操作者ID")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			return
		}

		// 3. 调用服务层删除逻辑
		if err := svc.DeleteUser(uint(id), operatorID.(uint)); err != nil {
			logger.Error("删除用户失败", zap.Uint("id", uint(id)), zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	}
}
