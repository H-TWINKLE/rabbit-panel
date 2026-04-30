package model

import "time"

// Session 会话信息
type Session struct {
	Username            string    `json:"username"`
	ExpiresAt           time.Time `json:"expires_at"`
	NeedChangePassword bool      `json:"need_change_password"`
}

// User 用户信息
type User struct {
	ID                  int       `json:"id"`
	Username            string    `json:"username"`
	PasswordHash        string    `json:"-"`
	NeedChangePassword bool      `json:"need_change_password"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token              string `json:"token"`
	NeedChangePassword bool   `json:"need_change_password"`
	Message            string `json:"message"`
}

// CurrentUserResponse 当前用户信息响应
type CurrentUserResponse struct {
	Username            string `json:"username"`
	NeedChangePassword bool   `json:"need_change_password"`
}
