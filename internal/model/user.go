package model

import "gorm.io/gorm"

// User 用户实体（对应MySQL表users）
type User struct {
	gorm.Model        // 内置ID/CreatedAt/UpdatedAt/DeletedAt字段
	Username   string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"` // 用户名（唯一索引）
	Email      string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`   // 邮箱（唯一索引）
	Password   string `gorm:"type:varchar(100);not null" json:"-"`                   // 密码（JSON序列化忽略）
	Role       string `gorm:"type:varchar(20);default:'user'" json:"role"`           // 角色（admin/user）
}
