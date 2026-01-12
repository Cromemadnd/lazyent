package lazyent

import (
	"github.com/Cromemadnd/lazyent/internal/types"
)

// Exported types
type Annotation = types.Annotation
type ProtoValidator = types.ProtoValidator

const (
	// ProtoValidatorNoValidator 不生成任何校验规则
	ProtoValidatorNoValidator = types.ProtoValidatorNoValidator
	// ProtoValidatorPGV 使用 PGV (protoc-gen-validate) 校验器 (默认)
	ProtoValidatorPGV = types.ProtoValidatorPGV
	// ProtoValidatorProtoValidate 使用 Buf ProtoValidate 校验器
	ProtoValidatorProtoValidate = types.ProtoValidatorProtoValidate
)

const (
	// BizPointerWithProtoMessage 使用 Biz 指针和 Proto 消息
	BizPointerWithProtoMessage = types.BizPointerWithProtoMessage
	// BizPointerWithProtoID 使用 Biz 指针和 Proto ID
	BizPointerWithProtoID = types.BizPointerWithProtoID
	// BizPointerWithProtoExclude 使用 Biz 指针和 Proto 排除
	BizPointerWithProtoExclude = types.BizPointerWithProtoExclude
	// BizIDWithProtoID 使用 Biz ID 和 Proto ID
	BizIDWithProtoID = types.BizIDWithProtoID
	// BizIDWithProtoExclude 使用 Biz ID 和 Proto 排除
	BizIDWithProtoExclude = types.BizIDWithProtoExclude
	// BizExcludeWithProtoExclude 使用 Biz 排除和 Proto 排除
	BizExcludeWithProtoExclude = types.BizExcludeWithProtoExclude
)
