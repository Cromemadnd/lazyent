package types

// ValidationRules 定义通用的校验规则结构体
// 这些规则将被转换为 PGV 或 ProtoValidate 语法
type ValidationRules struct {
	String   *StringRules   `json:"string,omitempty"`
	Number   *NumberRules   `json:"number,omitempty"` // Int, Uint, Float
	Repeated *RepeatedRules `json:"repeated,omitempty"`
	Enum     *EnumRules     `json:"enum,omitempty"`
}

type StringRules struct {
	Const       *string  `json:"const,omitempty"`
	Len         *uint64  `json:"len,omitempty"`
	MinLen      *uint64  `json:"min_len,omitempty"`
	MaxLen      *uint64  `json:"max_len,omitempty"`
	LenBytes    *uint64  `json:"len_bytes,omitempty"`
	Pattern     *string  `json:"pattern,omitempty"`
	Prefix      *string  `json:"prefix,omitempty"`
	Suffix      *string  `json:"suffix,omitempty"`
	Contains    *string  `json:"contains,omitempty"`
	In          []string `json:"in,omitempty"`
	NotIn       []string `json:"not_in,omitempty"`
	Email       bool     `json:"email,omitempty"`
	Hostname    bool     `json:"hostname,omitempty"`
	IP          bool     `json:"ip,omitempty"`
	IPV4        bool     `json:"ipv4,omitempty"`
	IPV6        bool     `json:"ipv6,omitempty"`
	URI         bool     `json:"uri,omitempty"`
	URIRef      bool     `json:"uri_ref,omitempty"`
	Address     bool     `json:"address,omitempty"`
	UUID        bool     `json:"uuid,omitempty"`
	IgnoreEmpty bool     `json:"ignore_empty,omitempty"`
}

// NumberRules 涵盖 int, uint, float
type NumberRules struct {
	Const       *float64  `json:"const,omitempty"`
	LT          *float64  `json:"lt,omitempty"`
	LTE         *float64  `json:"lte,omitempty"`
	GT          *float64  `json:"gt,omitempty"`
	GTE         *float64  `json:"gte,omitempty"`
	In          []float64 `json:"in,omitempty"`
	NotIn       []float64 `json:"not_in,omitempty"`
	IgnoreEmpty bool      `json:"ignore_empty,omitempty"`
}

type RepeatedRules struct {
	Len      *uint64 `json:"len,omitempty"`
	MinItems *uint64 `json:"min_items,omitempty"`
	MaxItems *uint64 `json:"max_items,omitempty"`
	Unique   bool    `json:"unique,omitempty"`
	// Items    *ValidationRules `json:"items,omitempty"` // Nested validation not supported yet via simple API
	IgnoreEmpty bool `json:"ignore_empty,omitempty"`
}

type EnumRules struct {
	Const       *int32  `json:"const,omitempty"`
	DefinedOnly bool    `json:"defined_only,omitempty"`
	In          []int32 `json:"in,omitempty"`
	NotIn       []int32 `json:"not_in,omitempty"`
}
