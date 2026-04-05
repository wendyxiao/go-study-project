package service

import (
	"go-study-project/internal/model"
	"go-study-project/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PostService interface {
	GetMostCommentedPost() (*model.Post, int64, error)
	AddPost(post *model.Post) error
	GetAllPostsByUserId(userId string) ([]model.PostVo, error)
}

type PostServiceImpl struct {
	postRepo repository.PostRepository // 仓库层接口（依赖注入）
	logger   *zap.Logger               // 日志组件（依赖注入）
}

func NewPostService(db *gorm.DB, logger *zap.Logger) *PostServiceImpl {
	postRepo := repository.NewPostRepo(db)
	return &PostServiceImpl{
		postRepo: postRepo,
		logger:   logger,
	}
}

// GetMostCommentedPost 获取评论最多的帖子
func (s *PostServiceImpl) GetMostCommentedPost() (*model.Post, int64, error) {
	post, count, err := s.postRepo.GetMaxCommentPost()
	if err != nil {
		s.logger.Error("查询评论最多帖子失败", zap.Error(err))
		return nil, 0, err
	}
	return post, count, nil
}

// GetMostCommentedPost 获取评论最多的帖子
func (s *PostServiceImpl) AddPost(post *model.Post) error {
	err := s.postRepo.AddPost(post)
	if err != nil {
		s.logger.Error("查询评论最多帖子失败", zap.Error(err))
		return err
	}
	return nil
}

func (s *PostServiceImpl) GetAllPostsByUserId(userId string) ([]model.PostVo, error) {
	var postvos []model.PostVo
	postvos, err := s.postRepo.GetAllPostsByUserId(userId)
	if err != nil {
		return nil, err
	}
	return postvos, nil
}

//为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。
