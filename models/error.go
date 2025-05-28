package models

import "fmt"

// ApiError 是一个自定义的错误类型，用于包含更多关于API调用失败的信息
type ApiError struct {
	Operation  string // 执行的操作
	Message    string // 错误信息
	HTTPStatus int    // HTTP 状态码 (例如 404, 500)。如果不是HTTP层面的错误，则可能为0或2xx。
	ApiCode    int    // API业务错误码 (DataHeaders.ResultCode)。如果不是业务逻辑错误，则为0。
	Err        error  // 原始错误 (可选)
}

// Error 实现 error 接口
func (e *ApiError) Error() string {
	errMsg := fmt.Sprintf("operation '%s' failed: %s", e.Operation, e.Message)
	if e.HTTPStatus != 0 && !(e.HTTPStatus >= 200 && e.HTTPStatus < 300 && e.ApiCode != 0) {
		// 仅当HTTPStatus本身表示错误，或者没有业务错误码时强调HTTPStatus
		errMsg += fmt.Sprintf(" (HTTP status: %d)", e.HTTPStatus)
	}
	if e.ApiCode != 0 && e.ApiCode != 1 { // 假设1是成功的业务码
		errMsg += fmt.Sprintf(" (Api code: %d)", e.ApiCode)
	}
	if e.Err != nil {
		errMsg += fmt.Sprintf(" (caused by: %v)", e.Err)
	}
	return errMsg
}

// Unwrap 返回包装的错误，以便与 errors.Is 和 errors.As 一起使用
func (e *ApiError) Unwrap() error {
	return e.Err
}
