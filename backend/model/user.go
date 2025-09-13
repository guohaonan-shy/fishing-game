package model

// User 用户信息
type User struct {
	Username string `json:"username"` // 用户名
	WxID     string `json:"wx_id"`    // 微信ID
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	WxID     string `json:"wx_id" binding:"required"`    // 微信ID
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	Username string `json:"username"` // 用户名
	WxID     string `json:"wx_id"`    // 微信ID
	Success  bool   `json:"success"`  // 操作结果
}

// GetUserRequest 获取用户请求
type GetUserRequest struct {
	WxID string `json:"wx_id" binding:"required"` // 微信ID
}

// GetUserResponse 获取用户响应
type GetUserResponse struct {
	Username string `json:"username"` // 用户名
	WxID     string `json:"wx_id"`    // 微信ID
	Found    bool   `json:"found"`    // 是否找到用户
}

// BatchGetUsersRequest 批量获取用户请求
type BatchGetUsersRequest struct {
	WxIDs []string `json:"wx_ids" binding:"required"` // 微信ID列表
}

// BatchGetUsersResponse 批量获取用户响应
type BatchGetUsersResponse struct {
	Users map[string]*User `json:"users"` // wx_id -> User 映射
}
