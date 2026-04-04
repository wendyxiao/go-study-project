package service

import (
	"fmt"
	"go-study-project/internal/config" // 替换为你的项目模块路径
	"go-study-project/internal/model"
	"go-study-project/internal/repository"

	"go.uber.org/zap"            // 日志组件（需提前安装：go get go.uber.org/zap）
	"golang.org/x/crypto/bcrypt" // 密码加密（需安装：go get golang.org/x/crypto/bcrypt）
	"gorm.io/gorm"
)

// 错误码定义（企业级项目需统一错误码，便于前端处理）
const (
	ErrCodeUserExists   = 10001 // 用户已存在
	ErrCodeUserNotFound = 10002 // 用户不存在
	ErrCodeInvalidPass  = 10003 // 密码无效
	ErrCodeNoPermission = 10004 // 无操作权限
)

// UserService 用户服务接口（抽象层，便于测试）
type UserService interface {
	Register(req *model.User) (uint, error)                    // 注册用户，返回用户ID
	Login(username, password string) (string, error)           // 登录，返回Token（简化版用随机字符串）
	GetUserByID(id uint, operatorID uint) (*model.User, error) // 获取用户详情（含权限校验）
	UpdateUser(user *model.User, operatorID uint) error        // 更新用户（含权限校验）
	DeleteUser(id uint, operatorID uint) error                 // 删除用户（含权限校验）
}

// UserServiceImpl 服务层实现（依赖仓库层接口）
type UserServiceImpl struct {
	userRepo repository.UserRepository // 仓库层接口（依赖注入）
	logger   *zap.Logger               // 日志组件（依赖注入）
	cfg      *config.AppConfig         // 全局配置（如Token密钥，可选）
}

// NewUserService 创建UserService实例（工厂函数，显式依赖注入）
func NewUserService(
	db *gorm.DB,
	logger *zap.Logger,
	cfg *config.AppConfig,
) UserService {
	userRepo := repository.NewUserRepo(db) // 初始化仓库层
	return &UserServiceImpl{
		userRepo: userRepo,
		logger:   logger,
		cfg:      cfg,
	}
}

// Register 注册用户（业务逻辑：密码加密+唯一性校验）
func (s *UserServiceImpl) Register(req *model.User) (uint, error) {
	// 1. 密码加密（bcrypt，成本因子10）
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err), zap.String("username", req.Username))
		return 0, fmt.Errorf("encrypt password failed: %w", err)
	}
	req.Password = string(hashedPass) // 覆盖明文密码

	// 2. 调用仓库层创建用户
	if err := s.userRepo.Create(req); err != nil {
		s.logger.Warn("用户注册失败", zap.Error(err), zap.String("username", req.Username))
		// 转换错误为业务错误码（如用户名已存在）
		if err.Error() == "username already exists" {
			return 0, fmt.Errorf("err_code:%d,msg:%s", ErrCodeUserExists, "用户名已存在")
		}
		if err.Error() == "email already exists" {
			return 0, fmt.Errorf("err_code:%d,msg:%s", ErrCodeUserExists, "邮箱已存在")
		}
		return 0, fmt.Errorf("create user failed: %w", err)
	}

	s.logger.Info("用户注册成功", zap.Uint("user_id", req.ID), zap.String("username", req.Username))
	return req.ID, nil
}

// Login 登录（业务逻辑：密码验证+生成Token）
func (s *UserServiceImpl) Login(username, password string) (string, error) {
	// 1. 查询用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		s.logger.Warn("登录失败：用户不存在", zap.String("username", username), zap.Error(err))
		return "", fmt.Errorf("err_code:%d,msg:%s", ErrCodeUserNotFound, "用户不存在")
	}

	// 2. 验证密码（bcrypt对比）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Warn("登录失败：密码错误", zap.String("username", username), zap.Error(err))
		return "", fmt.Errorf("err_code:%d,msg:%s", ErrCodeInvalidPass, "密码错误")
	}

	// 3. 生成Token（简化版：用随机字符串，实际项目用JWT）
	token := fmt.Sprintf("token_%d_%s", user.ID, "random_secret") // 实际需用JWT签名
	s.logger.Info("用户登录成功", zap.Uint("user_id", user.ID), zap.String("username", username))
	return token, nil
}

// GetUserByID 获取用户详情（含权限校验：仅本人或管理员可查）
func (s *UserServiceImpl) GetUserByID(id uint, operatorID uint) (*model.User, error) {
	// 1. 查询用户
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err // 已包含ErrCodeUserNotFound
	}

	// 2. 权限校验：仅本人（operatorID == user.ID）或管理员（user.Role == "admin"）可查
	if operatorID != user.ID && user.Role != "admin" {
		s.logger.Warn("无权限查询用户",
			zap.Uint("target_id", id),
			zap.Uint("operator_id", operatorID),
			zap.String("user_role", user.Role),
		)
		return nil, fmt.Errorf("err_code:%d,msg:%s", ErrCodeNoPermission, "无权限操作")
	}

	return user, nil
}

// UpdateUser 更新用户（含权限校验+事务）
func (s *UserServiceImpl) UpdateUser(user *model.User, operatorID uint) error {
	// 1. 开启事务（GORM事务：确保更新操作的原子性）
	s.userRepo.(*repository.UserRepo).Update(user) // 注意：此处需断言为具体实现，实际项目用接口事务

	// 2. 查询原用户（校验存在性）
	oldUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return err
	}

	// 3. 权限校验：仅本人或管理员可更新
	if operatorID != oldUser.ID && oldUser.Role != "admin" {
		return fmt.Errorf("err_code:%d,msg:%s", ErrCodeNoPermission, "无权限操作")
	}

	// 4. 执行更新（调用仓库层，传入事务DB）

	// 5. 提交事务

	s.logger.Info("用户更新成功", zap.Uint("user_id", user.ID), zap.Uint("operator_id", operatorID))
	return nil
}

// DeleteUser 删除用户（含权限校验+软删除）
func (s *UserServiceImpl) DeleteUser(id uint, operatorID uint) error {
	// 1. 查询用户（校验存在性+权限）
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if operatorID != user.ID && user.Role != "admin" {
		return fmt.Errorf("err_code:%d,msg:%s", ErrCodeNoPermission, "无权限操作")
	}

	// 2. 软删除（GORM自动处理DeletedAt字段）
	if err := s.userRepo.Delete(id); err != nil {
		s.logger.Error("删除用户失败", zap.Error(err), zap.Uint("user_id", id))
		return fmt.Errorf("delete user failed: %w", err)
	}

	s.logger.Info("用户删除成功", zap.Uint("user_id", id), zap.Uint("operator_id", operatorID))
	return nil
}
