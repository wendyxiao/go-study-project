package api

import (
	"go-study-project/internal/model"
	"go-study-project/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterCommentoutes(r *gin.RouterGroup, c service.CommentService, logger *zap.Logger) {
	// 公开接口（无需认证）
	r.POST("/comment/add", createCommentHandler(c, logger)) // 用户注册
	r.DELETE("/comment/:id", deleteCommentHandler(c, logger))
}
func createCommentHandler(comment service.CommentService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		//获取参数
		var req model.Comment
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("注册请求参数错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误", "details": err.Error()})
			return
		}
		err := comment.Create(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusCreated, gin.H{"comment": "add success"})
	}
}

func deleteCommentHandler(comment service.CommentService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		comment, err := comment.Delete(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"comment": comment})
	}
}
