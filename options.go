package lazyent

import "github.com/Cromemadnd/lazyent/internal/types"

// Convenience aliases

// WithEnumValues 设置枚举数值映射
// key: 枚举名称 (例如 "ACTIVE"), value: 枚举值 (例如 1)
func WithEnumValues(values map[string]int32) Annotation {
	return Annotation{
		EnumValues: values,
	}
}

// WithEdgeFieldStrategy 设置 Edge 字段的生成策略
// 默认策略为 BizPointerWithProtoMessage
func WithEdgeFieldStrategy(strategy types.EdgeFieldStrategy) Annotation {
	return Annotation{
		EdgeFieldStrategy: strategy,
	}
}

// WithBizName 自定义生成的 Biz 结构体字段名称
func WithBizName(name string) Annotation {
	return Annotation{
		BizName: name,
	}
}

// WithBizType 自定义生成的 Biz 结构体字段类型
// 例如: "json.RawMessage", "map[string]interface{}"
func WithBizType(typeName string) Annotation {
	return Annotation{
		BizType: typeName,
	}
}

// WithProtoName 自定义生成的 Proto message 字段名称
// 建议遵循 snake_case
func WithProtoName(name string) Annotation {
	return Annotation{
		ProtoName: name,
	}
}

// WithProtoType 自定义生成的 Proto message 字段类型
func WithProtoType(typeName string) Annotation {
	return Annotation{
		ProtoType: typeName,
	}
}

// WithProtoFieldID 手动指定 Proto 字段的 ID (Tag)
// 如果不指定，将自动生成
func WithProtoFieldID(id int32) Annotation {
	return Annotation{
		ProtoFieldID: id,
	}
}

// Virtual 标记字段为虚拟字段
// 虚拟字段不会映射到数据库，仅在 Biz 和 Proto 层存在
func Virtual() Annotation {
	return Annotation{
		Virtual: true,
	}
}

// MergeAnnotations 合并多个 Annotation 选项
// 后面的选项会覆盖前面的选项
func MergeAnnotations(opts ...Annotation) Annotation {
	merged := Annotation{}
	for _, opt := range opts {
		if opt.EnumValues != nil {
			merged.EnumValues = opt.EnumValues
		}
		if opt.EdgeFieldStrategy != 0 {
			merged.EdgeFieldStrategy = opt.EdgeFieldStrategy
		}
		if opt.BizName != "" {
			merged.BizName = opt.BizName
		}
		if opt.BizType != "" {
			merged.BizType = opt.BizType
		}
		if opt.ProtoName != "" {
			merged.ProtoName = opt.ProtoName
		}
		if opt.ProtoType != "" {
			merged.ProtoType = opt.ProtoType
		}
		if opt.ProtoFieldID != 0 {
			merged.ProtoFieldID = opt.ProtoFieldID
		}
		if opt.Virtual {
			merged.Virtual = true
		}
	}
	return merged
}

// Primitive Ptr Helpers

func Int(v int) *int       { return &v }
func Int8(v int8) *int8    { return &v }
func Int16(v int16) *int16 { return &v }
func Int32(v int32) *int32 { return &v }
func Int64(v int64) *int64 { return &v }

func Uint(v uint) *uint       { return &v }
func Uint8(v uint8) *uint8    { return &v }
func Uint16(v uint16) *uint16 { return &v }
func Uint32(v uint32) *uint32 { return &v }
func Uint64(v uint64) *uint64 { return &v }

func Float32(v float32) *float32 { return &v }
func Float64(v float64) *float64 { return &v }

func String(v string) *string { return &v }
func Bool(v bool) *bool       { return &v }
