package model

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	// 业务字段
	Content string `gorm:"type:varchar(500);not null;comment:评论内容" json:"content" binding:"required,min=1,max=500"`          // 评论内容（非空，长度限制500）
	PostId  uint   `gorm:"not null;index:idx_post_id;comment:关联帖子ID" json:"post_id" binding:"required"`                      // 帖子ID（外键，非空，索引优化）
	UserId  string `gorm:"type:varchar(36);not null;index:idx_user_id;comment:用户ID（UUID）" json:"user_id" binding:"required"` // 用户ID（UUID格式，非空，索引优化）

}

//为 Comment 模型添加一个钩子函数，在评论删除时检查文章的评论数量，如果评论数量为 0，
//则更新文章的评论状态为 "无评论"

func (Comment) TableName() string {
	return "comment"
}

// 删除后的钩子函数
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	reuslt := tx.Model(&Comment{}).Where("post_id = ?", c.PostId)
	if reuslt.Error != nil {
		return errors.Wrap(reuslt.Error, "查询失败")
	}
	if reuslt.RowsAffected == 0 {
		//更新post表的评论状态
		tx.Model(&Post{}).Where("id = ?", c.PostId).UpdateColumn("comment_state", "无评论状态")
	}

	return nil
}
