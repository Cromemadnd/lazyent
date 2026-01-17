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

// WithEdgeInStrategy 设置 Edge 在输入层（Request/Input）的策略
func WithEdgeInStrategy(strategy types.EdgeStrategy) Annotation {
	return Annotation{
		EdgeInStrategy: strategy,
	}
}

// WithEdgeOutStrategy 设置 Edge 在输出层（Response）的策略
func WithEdgeOutStrategy(strategy types.EdgeStrategy) Annotation {
	return Annotation{
		EdgeOutStrategy: strategy,
	}
}

// WithFieldInStrategy 设置普通字段在输入层的策略
func WithFieldInStrategy(strategy types.FieldStrategy) Annotation {
	return Annotation{
		FieldInStrategy: strategy,
	}
}

// WithFieldOutStrategy 设置普通字段在输出层的策略
func WithFieldOutStrategy(strategy types.FieldStrategy) Annotation {
	return Annotation{
		FieldOutStrategy: strategy,
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
		if opt.EdgeInStrategy != 0 {
			merged.EdgeInStrategy = opt.EdgeInStrategy
		}
		if opt.EdgeOutStrategy != 0 {
			merged.EdgeOutStrategy = opt.EdgeOutStrategy
		}
		if opt.FieldInStrategy != 0 {
			merged.FieldInStrategy = opt.FieldInStrategy
		}
		if opt.FieldOutStrategy != 0 {
			merged.FieldOutStrategy = opt.FieldOutStrategy
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
		if opt.ProtoValidation != "" {
			merged.ProtoValidation = opt.ProtoValidation
		}
	}
	return merged
}

// WithProtoValidation 自定义 Proto 字段验证规则 (Buf Validate)
// 例如: "string.min_len = 1"
// 最终生成: [(buf.validate.field).string.min_len = 1]
func WithProtoValidation(rules string) Annotation {
	return Annotation{
		ProtoValidation: rules,
	}
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
