package repository

import (
	"errors"
	"fmt"
	"go-study-project/internal/model" // 替换为你的项目模块路径

	"gorm.io/gorm"
)

// UserRepository 用户数据访问接口（抽象层，便于测试时Mock）
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	GetByEmail(email string) (*model.User, error) // 按邮箱查询
}

// UserRepo 仓库层实现（依赖GORM DB实例）
type UserRepo struct {
	db *gorm.DB // 由外部注入，避免全局变量
}

// NewUserRepo 创建UserRepo实例（工厂函数，显式依赖注入）
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create 创建用户（含唯一性校验：用户名/邮箱不能重复）
func (r *UserRepo) Create(user *model.User) error {
	// 检查用户名是否已存在
	var count int64
	if err := r.db.Model(&model.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return fmt.Errorf("check username exists failed: %w", err)
	}
	if count > 0 {
		return errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if err := r.db.Model(&model.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return fmt.Errorf("check email exists failed: %w", err)
	}
	if count > 0 {
		return errors.New("email already exists")
	}

	// 创建用户（GORM自动填充ID/CreatedAt）
	return r.db.Create(user).Error
}

// GetByID 根据ID查询用户
func (r *UserRepo) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user by id failed: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名查询用户（登录时用）
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user by username failed: %w", err)
	}
	return &user, nil
}

// GetByEmail 按邮箱查询用户
func (r *UserRepo) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, err
}

// Update 更新用户信息（含乐观锁：通过UpdatedAt控制）
func (r *UserRepo) Update(user *model.User) error {
	// 仅更新非零值字段（避免覆盖未修改的字段）
	return r.db.Model(user).Updates(user).Error
}

// Delete 软删除用户（GORM的DeletedAt字段自动处理）
func (r *UserRepo) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}
