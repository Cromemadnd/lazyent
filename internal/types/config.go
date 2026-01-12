package types

// ProtoValidator 定义 Proto 校验器类型
type ProtoValidator int

const (
	// ProtoValidatorNoValidator 不生成任何校验规则
	ProtoValidatorNoValidator ProtoValidator = iota
	// ProtoValidatorPGV 使用 PGV (protoc-gen-validate) 校验器 (默认)
	ProtoValidatorPGV
	// ProtoValidatorProtoValidate 使用 Buf ProtoValidate 校验器
	ProtoValidatorProtoValidate
)
