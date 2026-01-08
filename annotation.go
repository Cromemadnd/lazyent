package lazyent

type BizStrategy int

const (
	BizPointer BizStrategy = iota // *Group (默认)
	BizValue                      // Group (值类型)
	BizIDOnly                     // 仅ID
	BizExclude                    // 不生成
)

type ProtoStrategy int

const (
	ProtoMessage ProtoStrategy = iota // message Group
	ProtoID                           // group_id
	ProtoExclude                      // 不生成
)

// Annotation 定义 LazyEnt 的配置注解
type Annotation struct {
	// EnumValues 定义 Proto 枚举数值映射: "ENUM_VAL": 1
	EnumValues map[string]int32 `json:"enum_values"`
	// BizStrategy 外键 Biz 策略
	BizStrategy BizStrategy `json:"biz_strategy"`
	// ProtoStrategy 外键 Proto 策略
	ProtoStrategy ProtoStrategy `json:"proto_strategy"`
	// FieldID 指定 Proto 字段 ID
	FieldID int32 `json:"field_id"`
	// Validate 指定 Proto 校验规则 (pgv)
	Validate string `json:"validate"`
}

// Name 实现 ent.Annotation 接口
func (Annotation) Name() string {
	return "LazyEnt"
}
