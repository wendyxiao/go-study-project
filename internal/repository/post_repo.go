package repository

import (
	"fmt"
	"go-study-project/internal/model"

	"gorm.io/gorm"
)

// 博客数据访问接口
type PostRepository interface {
	//新增博客
	AddPost(post *model.Post) error
	GetAllPostsByUserId(userId string) ([]model.PostVo, error)
	GetMaxCommentPost() (*model.Post, int64, error)
}

// DB
type PostRepo struct {
	db *gorm.DB
}

func NewPostRepo(db *gorm.DB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) AddPost(post *model.Post) error {
	return r.db.Create(post).Error
}

// getAllPostsByUserId
// 编写Go代码，使用Gorm查询某个用户发布的所有文章及其对应的评论信息。
func (r *PostRepo) GetAllPostsByUserId(userId string) ([]model.PostVo, error) {
	var postvos []model.PostVo
	var posts []model.Post
	var comments []model.Comment
	if err := r.db.Model(&model.Post{}).Where("user_id = ?", userId).Find(&posts, userId).Error; err != nil {
		return nil, fmt.Errorf("get post by user_id failed: %w", err)
	}
	for _, post := range posts {

		if err := r.db.Model(&model.Comment{}).Where("post_id = ?", post.ID).Find(&comments, userId).Error; err != nil {
			return nil, fmt.Errorf("get comment by post_id failed: %w", err)
		}
		postvo := model.PostVo{
			P: post,
			C: comments,
		}
		postvos = append(postvos, postvo)
	}
	return postvos, nil
}

// 编写Go代码，使用Gorm查询评论数量最多的文章信息。
func (r *PostRepo) GetMaxCommentPost() (*model.Post, int64, error) { // 步骤1：子查询计算每个帖子的评论数（过滤软删除评论）
	subQuery := r.db.Model(&model.Comment{}).
		Select("post_id, COUNT(id) AS comment_count").
		Where("deleted_at IS NULL"). // 排除已删除评论
		Group("post_id")

	// 步骤2：关联帖子表，按评论数降序取第一条
	type Result struct {
		PostID       uint  // 帖子ID
		CommentCount int64 // 评论数
	}
	var result Result
	err := r.db.Model(&model.Post{}).
		Select("posts.id AS post_id, COALESCE(sub.comment_count, 0) AS comment_count").
		Joins("LEFT JOIN (?) AS sub ON sub.post_id = posts.id", subQuery). // 关联子查询
		Where("posts.deleted_at IS NULL").                                 // 排除已删除帖子
		Order("comment_count DESC").                                       // 按评论数降序
		Limit(1).                                                          // 取最多评论的1篇
		Scan(&result).Error                                                // 扫描结果到结构体

	if err != nil {
		return nil, 0, err
	}
	if result.PostID == 0 { // 无数据时返回错误
		return nil, 0, gorm.ErrRecordNotFound
	}

	// 步骤3：查询帖子详情（预加载作者信息）
	var post model.Post
	err = r.db.Preload("users").First(&post, result.PostID).Error // Preload 加载关联作者
	if err != nil {
		return nil, 0, err
	}

	return &post, result.CommentCount, nil

}
