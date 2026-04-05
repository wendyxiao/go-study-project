package service

import (
	"go-study-project/internal/model"
	"go-study-project/internal/repository"

	"gorm.io/gorm"
)

type CommentService interface {
	Create(comment *model.Comment) error
	Delete(id string) (*model.Comment, error)
	findById(id uint) (*model.Comment, error)
}
type CommentServiceImpl struct {
	commentRepo repository.CommentRepository
}

func NewCommentService(db *gorm.DB) *CommentServiceImpl {
	commentRepo := repository.NewCommentRepo(db)
	return &CommentServiceImpl{
		commentRepo: commentRepo,
	}
}

func (c *CommentServiceImpl) Create(comment *model.Comment) error {
	err := c.commentRepo.AddComment(comment)
	if err != nil {
		return err
	}
	return nil
}

func (c *CommentServiceImpl) Delete(id string) (*model.Comment, error) {
	return c.commentRepo.DeleteComment(id)
}

func (c *CommentServiceImpl) findById(id uint) (*model.Comment, error) {
	return c.commentRepo.FindCommentById(id)
}
