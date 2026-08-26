package model

import "errors"

// 领域错误：供 store/service/httpapi 逐层映射为 HTTP 状态码。
var (
	// ErrNotFound 表示实体不存在。
	ErrNotFound = errors.New("not found")
	// ErrConflict 表示资源已存在（重复导入、重复创建）。
	ErrConflict = errors.New("conflict")
	// ErrInvalid 表示输入非法（坐标越界、格式错误）。
	ErrInvalid = errors.New("invalid input")
	// ErrState 表示非法状态流转。
	ErrState = errors.New("illegal state transition")
	// ErrFrozen 表示试图修改已冻结/封存的不可变对象。
	ErrFrozen = errors.New("frozen object cannot be modified")
	// ErrCycle 表示关系图出现出处循环。
	ErrCycle = errors.New("provenance cycle detected")
	// ErrTechnique 表示技法数据缺失导致无法验证。
	ErrTechnique = errors.New("technique data missing")
)
