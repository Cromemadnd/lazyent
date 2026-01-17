package lazyent

import (
	"github.com/Cromemadnd/lazyent/internal/types"
)

// Exported types & Consts
type Annotation = types.Annotation

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
