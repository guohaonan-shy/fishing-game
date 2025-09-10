package model

// APIResponse 统一API响应格式
type APIResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Code: 200,
		Data: data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse() *APIResponse {
	return &APIResponse{
		Code: 500,
	}
}
