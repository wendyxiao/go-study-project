package api

import (
	"go-study-project/internal/model"
	"go-study-project/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterPostoutes(r *gin.RouterGroup, post service.PostService, logger *zap.Logger) {
	// 公开接口（无需认证）
	r.POST("/post/add", createPostHandler(post, logger)) // 用户注册
	r.GET("/post/:id", getPostsHandler(post, logger))
	r.GET("/posts", getPostMaxHandler(post, logger)) //无如参数
}
func getPostMaxHandler(post service.PostService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentedPost, i, err := post.GetMostCommentedPost()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"post": commentedPost,
			"i":    i,
		})
	}
}

func getPostsHandler(post service.PostService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取路径参数ID
		idStr := c.Param("id")
		logger.Info("idStr----------" + idStr)
		posts, err := post.GetAllPostsByUserId(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": posts})
	}
}

func createPostHandler(post service.PostService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.Post
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("注册请求参数错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误", "details": err.Error()})
			return
		}
		err := post.AddPost(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		// 3. 返回成功响应（HTTP 201 Created）
		c.JSON(http.StatusCreated, gin.H{
			"code":    0,
			"message": "新增成功",
			"data":    gin.H{"user_id": nil},
		})
	}
}
