package repository

import (
	"go-study-project/internal/model"

	"gorm.io/gorm"
)

type CommentRepository interface {
	AddComment(comment *model.Comment) error
	DeleteComment(id string) (*model.Comment, error)
	FindCommentById(id uint) (*model.Comment, error)
}
type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{
		db: db,
	}
}
func (c *CommentRepo) AddComment(comment *model.Comment) error {
	return c.db.Create(comment).Error
}

// 删除
func (c *CommentRepo) DeleteComment(id string) (*model.Comment, error) {
	comment := &model.Comment{}
	c.db.Unscoped().Delete(comment, id)
	return comment, nil
}

// 查询
func (c *CommentRepo) FindCommentById(id uint) (*model.Comment, error) {
	comment := &model.Comment{}
	err := c.db.Where("id = ?", id).Find(comment).Error
	return comment, err
}
