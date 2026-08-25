package domain

import "errors"

var (
	ErrNotFound          = errors.New("资源不存在")
	ErrConflict          = errors.New("修订号冲突")
	ErrInvalidTransition = errors.New("不允许的状态迁移")
	ErrValidation        = errors.New("数据校验失败")
	ErrStaleEvidence     = errors.New("整改证据不是基于最新内容版本")
	ErrNotReady          = errors.New("验收案尚不满足操作条件")
	ErrIntegrity         = errors.New("数据完整性校验失败")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []FieldError

func (e ValidationErrors) Error() string { return ErrValidation.Error() }

func (e ValidationErrors) Is(target error) bool { return target == ErrValidation }

func (e ValidationErrors) Empty() bool { return len(e) == 0 }
