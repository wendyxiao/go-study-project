package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	// 业务字段
	Title        string `gorm:"type:varchar(200);not null;comment:文章标题" json:"title" binding:"required,min=1,max=200"`            // 标题（非空，长度限制200）
	Content      string `gorm:"type:text;not null;comment:文章内容" json:"content" binding:"required,min=1"`                          // 内容（长文本，非空）
	UserId       string `gorm:"type:varchar(36);not null;index:idx_user_id;comment:作者ID（UUID）" json:"user_id" binding:"required"` // 作者ID（UUID格式，非空，索引优化）
	CommentState string `gorm:"type:varchar(36);not null;comment:评论状态" json:"comment_state" `                                     // 评论状态
}

func (Post) TableName() string {
	return "post"
}
func (p *Post) AfterCreate(tx *gorm.DB) error {
	fmt.Printf("after create post: %+v\n", p)
	//1. 原子操作更新用户数量，避免并发问题
	result := tx.Model(&User{}).Where("id=?", p.UserId).Update("article_count", gorm.Expr("article_count + 1")).Error
	//2. 处理错误
	if result.Error != nil {
		return fmt.Errorf("更新文章失败")
	}
	return nil
}
