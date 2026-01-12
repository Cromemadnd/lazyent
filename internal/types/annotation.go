package types

type EdgeFieldStrategy int // EdgeFieldStrategy 外键字段策略

const (
	// BizPointerWithProtoMessage (Default)
	//  - Biz:   *Group / []*Group
	//  - Proto: Group group / repeated Group groups
	BizPointerWithProtoMessage EdgeFieldStrategy = iota

	// BizPointerWithProtoID
	//  - Biz:   *Group / []*Group
	//  - Proto: string(group_id) / repeated string(group_ids)
	BizPointerWithProtoID

	// BizPointerWithProtoExclude
	//  - Biz:   *Group / []*Group
	//  - Proto: (Excluded)
	BizPointerWithProtoExclude

	// BizIDWithProtoID
	//  - Biz:   GroupID / []GroupID
	//  - Proto: string(group_id) / repeated string(group_ids)
	BizIDWithProtoID

	// BizIDWithProtoExclude
	//  - Biz:   GroupID / []GroupID
	//  - Proto: (Excluded)
	BizIDWithProtoExclude

	// BizExcludeWithProtoExclude
	//  - Biz:   (Excluded)
	//  - Proto: (Excluded)
	BizExcludeWithProtoExclude
)

// Annotation 定义 LazyEnt 的配置注解
type Annotation struct {
	EnumValues        map[string]int32  `json:"enum_values"`         // EnumValues 定义枚举数值映射: "ENUM_VAL": 1 (仅 Enum Fields有效)
	EdgeFieldStrategy EdgeFieldStrategy `json:"edge_field_strategy"` // 仅 Edge 字段有效，默认为 BizPointerWithProtoMessage
	BizName           string            `json:"biz_name"`            // Biz Field 名称
	BizType           string            `json:"biz_type"`            // Biz Field 自定义类型
	ProtoName         string            `json:"proto_name"`          // Proto Field 名称
	ProtoType         string            `json:"proto_type"`          // Proto Field 自定义类型
	ProtoFieldID      int32             `json:"proto_field_id"`      // ProtoFieldID 指定 Proto 字段 ID
	ProtoValidation   string            `json:"proto_validation"`    // ProtoValidation 指定 Proto 校验规则 (pgv)
	Validation        *ValidationRules  `json:"validation"`          // Validation 指定结构化校验规则
}

// Name 实现 ent.Annotation 接口
func (Annotation) Name() string {
	return "LazyEnt"
}
