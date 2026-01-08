package lazyent

// Annotation 定义 LazyEnt 的配置注解
type Annotation struct {
	// EnumValues 定义 Proto 枚举数值映射: "ENUM_VAL": 1
	EnumValues map[string]int32 `json:"enum_values"`
}

// Name 实现 ent.Annotation 接口
func (Annotation) Name() string {
	return "LazyEnt"
}
