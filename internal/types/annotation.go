package types

import "fmt"

type EdgeStrategy uint32
type FieldStrategy uint32

const (
	// --- Proto 表现层 (Bits 0-7) ---
	EdgeProtoMessage  EdgeStrategy = 1 << 0 // Entity entity
	EdgeProtoID       EdgeStrategy = 1 << 1 // string entity_id
	EdgeProtoExcluded EdgeStrategy = 1 << 2 // Proto 中隐藏
	EdgeProtoMask     EdgeStrategy = 0xFF

	// --- Biz 表现层 (Bits 8-15) ---
	EdgeBizPointer  EdgeStrategy = 1 << 8  // *Entity
	EdgeBizID       EdgeStrategy = 1 << 9  // EntityID (string)
	EdgeBizExcluded EdgeStrategy = 1 << 10 // Biz 中忽略
	EdgeBizMask     EdgeStrategy = 0xFF00

	// --- Proto 表现层 (Bits 0-7) ---
	FieldProtoRequired FieldStrategy = 1 << 0 // type field
	FieldProtoOptional FieldStrategy = 1 << 1 // optional type field
	FieldProtoExcluded FieldStrategy = 1 << 2 // Proto 中隐藏
	FieldProtoMask     FieldStrategy = 0xFF

	// --- Biz 表现层 (Bits 8-15) ---
	FieldBizValue    FieldStrategy = 1 << 8  // type field
	FieldBizPointer  FieldStrategy = 1 << 9  // type* field (可以适配 optional)
	FieldBizExcluded FieldStrategy = 1 << 10 // Biz 中隐藏
	FieldBizMask     FieldStrategy = 0xFF00
)

// Annotation 定义 LazyEnt 的配置注解
type Annotation struct {
	EnumValues   map[string]int32 `json:"enum_values"`    // EnumValues 定义枚举数值映射: "ENUM_VAL": 1 (仅 Enum Fields有效)
	BizName      string           `json:"biz_name"`       // Biz Field 名称
	BizType      string           `json:"biz_type"`       // Biz Field 自定义类型
	ProtoName    string           `json:"proto_name"`     // Proto Field 名称
	ProtoType    string           `json:"proto_type"`     // Proto Field 自定义类型
	ProtoFieldID int32            `json:"proto_field_id"` // ProtoFieldID 指定 Proto 字段 ID
	Virtual      bool             `json:"virtual"`        // 是否为虚拟字段 (不存入数据库)

	// 新的四大策略
	EdgeInStrategy   EdgeStrategy  `json:"edge_in_strategy"`   // 关联输入策略
	EdgeOutStrategy  EdgeStrategy  `json:"edge_out_strategy"`  // 关联输出策略
	FieldInStrategy  FieldStrategy `json:"field_in_strategy"`  // 字段输入策略
	FieldOutStrategy FieldStrategy `json:"field_out_strategy"` // 字段输出策略
}

// Name 实现 ent.Annotation 接口
func (Annotation) Name() string {
	return "LazyEnt"
}

func (s EdgeStrategy) Validate() error {
	// Check Proto Mask
	protoBits := s & EdgeProtoMask
	if protoBits != 0 && (protoBits&(protoBits-1)) != 0 {
		return fmt.Errorf("multiple bits set in EdgeProtoMask: %d", protoBits)
	}
	// Check Biz Mask
	bizBits := s & EdgeBizMask
	if bizBits != 0 && (bizBits&(bizBits-1)) != 0 {
		return fmt.Errorf("multiple bits set in EdgeBizMask: %d", bizBits)
	}
	return nil
}

func (s FieldStrategy) Validate() error {
	// Check Proto Mask
	protoBits := s & FieldProtoMask
	if protoBits != 0 && (protoBits&(protoBits-1)) != 0 {
		return fmt.Errorf("multiple bits set in FieldProtoMask: %d", protoBits)
	}
	// Check Biz Mask
	bizBits := s & FieldBizMask
	if bizBits != 0 && (bizBits&(bizBits-1)) != 0 {
		return fmt.Errorf("multiple bits set in FieldBizMask: %d", bizBits)
	}
	return nil
}
