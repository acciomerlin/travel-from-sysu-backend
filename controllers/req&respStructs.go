package controllers

// ErrorResponse 错误返回信息
type ErrorResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error"`
}
