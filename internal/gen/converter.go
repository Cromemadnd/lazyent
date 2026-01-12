package gen

import (
	"fmt"
	"strings"

	entgen "entgo.io/ent/entc/gen"
)

func convertToProto(f *entgen.Field, nodeName string) string {
	if f == nil {
		return ""
	}
	if isSensitive(f) {
		return zeroValue(getProtoType(f))
	}
	if f.IsEnum() {
		if isExternalEnum(f) {
			return fmt.Sprintf("string(b.%s)", bizFieldName(f))
		}
		return fmt.Sprintf("%s(b.%s)", enumToProtoFuncName(f, nodeName), bizFieldName(f))
	}
	if f.Type.String() == "time.Time" {
		return fmt.Sprintf("timestamppb.New(b.%s)", bizFieldName(f))
	}
	if f.Type.String() == "uuid.UUID" {
		return fmt.Sprintf("b.%s", bizFieldName(f))
	}

	if strings.HasPrefix(f.Type.String(), "[]") && f.Type.String() != "[]byte" {
		return "b." + bizFieldName(f)
	}

	pt := getProtoType(f)
	goProtoType := pt
	if pt == "double" {
		goProtoType = "float64"
	}
	if pt == "float" {
		goProtoType = "float32"
	}
	if pt == "bytes" {
		goProtoType = "[]byte"
	}

	if f.Type.String() == "string" && goProtoType == "string" {
		return "b." + bizFieldName(f)
	}
	if f.Type.String() == "bool" && goProtoType == "bool" {
		return "b." + bizFieldName(f)
	}

	switch goProtoType {
	case "int32", "int64", "uint32", "uint64", "float32", "float64":
		return fmt.Sprintf("%s(b.%s)", goProtoType, bizFieldName(f))
	case "string":
		if f.Type.String() != "string" {
			return fmt.Sprintf("string(b.%s)", bizFieldName(f))
		}
	case "bool":
		if f.Type.String() != "bool" {
			return fmt.Sprintf("bool(b.%s)", bizFieldName(f))
		}
	}

	return "b." + bizFieldName(f)
}

func convertFromProto(f *entgen.Field, nodeName string) string {
	if f == nil {
		return ""
	}
	if isSensitive(f) {
		return zeroValue(bizFieldType(f))
	}
	// For safer generation, we should use convertFromProtoUsage after generating Setup code.
	// But for backward compatibility or simple fields, we return inline.
	// Panic-prone conversions (MustParse) should be avoided if possible.

	if f.IsEnum() {
		if isExternalEnum(f) {
			return fmt.Sprintf("%s(p.%s)", getExternalEnumName(f), protoGoName(f))
		}
		return fmt.Sprintf("%s(p.%s)", enumFromProtoFuncName(f, nodeName), protoGoName(f))
	}
	if f.Type.String() == "time.Time" {
		return fmt.Sprintf("p.%s.AsTime()", protoGoName(f))
	}

	targetType := bizFieldType(f)
	if targetType == "" {
		targetType = f.Type.String()
	}

	if f.Type.String() == "uuid.UUID" {
		if targetType == "string" {
			return "p." + protoGoName(f)
		}
		// Uses MustParse which is dangerous. The Template should now use convertFromProtoSetup/Usage.
		// If this is still called directly, we might Panic.
		return fmt.Sprintf("uuid.MustParse(p.%s)", protoGoName(f))
	}

	if targetType == "string" {
		return "p." + protoGoName(f)
	}
	if targetType == "bool" {
		return "p." + protoGoName(f)
	}

	switch targetType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return fmt.Sprintf("%s(p.%s)", targetType, protoGoName(f))
	}

	return "p." + protoGoName(f)
}

func convertEntToBiz(f *entgen.Field, nodeName string, expr string) string {
	if f.IsEnum() {
		if isExternalEnum(f) {
			return expr
		}
		return fmt.Sprintf("Ent%s%sToBiz(%s)", nodeName, f.StructField(), expr)
	}

	bizType := explicitBizType(f)
	if bizType == "" {
		if f.Type.String() == "uuid.UUID" {
			bizType = "string"
		} else {
			return expr
		}
	}

	isPtr := f.Nillable
	exprVal := expr
	if isPtr {
		exprVal = "*" + expr
	}

	var castExpr string
	entType := f.Type.String()

	if entType == "time.Time" {
		if bizType == "int64" {
			castExpr = fmt.Sprintf("%s.Unix()", exprVal)
		} else if bizType == "string" {
			castExpr = fmt.Sprintf("%s.Format(time.RFC3339)", exprVal)
		} else {
			castExpr = fmt.Sprintf("%s(%s)", bizType, exprVal)
		}
	} else if entType == "uuid.UUID" {
		if bizType == "string" {
			castExpr = fmt.Sprintf("%s.String()", expr)
		} else {
			castExpr = fmt.Sprintf("%s(%s)", bizType, exprVal)
		}
	} else {
		castExpr = fmt.Sprintf("%s(%s)", bizType, exprVal)
	}

	if isPtr {
		zero := "0"
		switch bizType {
		case "string":
			zero = `""`
		case "bool":
			zero = "false"
		}
		return fmt.Sprintf("func() %s { if %s != nil { return %s }; return %s }()", bizType, expr, castExpr, zero)
	}

	return castExpr
}

func convertBizToEnt(f *entgen.Field, nodeName string, expr string) string {
	if f.IsEnum() {
		if isExternalEnum(f) {
			return expr
		}
		return fmt.Sprintf("Biz%s%sToEnt(%s)", nodeName, f.StructField(), expr)
	}

	bizType := explicitBizType(f)
	if bizType == "" {
		return expr
	}

	entType := f.Type.String()
	var entExpr string
	if entType == "time.Time" {
		if bizType == "int64" {
			entExpr = fmt.Sprintf("time.Unix(%s, 0)", expr)
		} else if bizType == "string" {
			entExpr = fmt.Sprintf("func() time.Time { t, _ := time.Parse(time.RFC3339, %s); return t }()", expr)
		} else {
			entExpr = fmt.Sprintf("%s(%s)", entType, expr)
		}
	} else if entType == "uuid.UUID" {
		if bizType == "string" {
			// Dangerous MustParse
			entExpr = fmt.Sprintf("uuid.MustParse(%s)", expr)
		} else {
			entExpr = fmt.Sprintf("%s(%s)", entType, expr)
		}
	} else {
		entExpr = fmt.Sprintf("%s(%s)", entType, expr)
	}

	isPtr := f.Nillable
	if isPtr {
		return fmt.Sprintf("func() *%s { x := %s; return &x }()", entType, entExpr)
	}

	return entExpr
}

func edgeConvertToProto(e *entgen.Edge) string {
	if err := validateConflict(e); err != nil {
		panic(err)
	}
	if isProtoMessage(e) {
		if isBizPointer(e) {
			return fmt.Sprintf("Biz%sToProto(b.%s)", e.Type.Name, bizEdgeName(e))
		}
	} else {
		if isBizIDOnly(e) {
			field := bizEdgeName(e)
			typ := e.Type.ID.Type.String()
			if typ == "uuid.UUID" {
				return fmt.Sprintf("b.%s", field)
			}
			if typ == "int" || typ == "int32" {
				return fmt.Sprintf("int32(b.%s)", field)
			}
			return "b." + field
		}
		access := fmt.Sprintf("b.%s.UUID", bizEdgeName(e))
		typ := e.Type.ID.Type.String()
		if typ == "uuid.UUID" {
			return access
		}
		if typ == "int" || typ == "int32" {
			return fmt.Sprintf("int32(%s)", access)
		}
		return access
	}
	return "nil"
}

func edgeConvertFromProto(e *entgen.Edge) string {
	if isProtoMessage(e) {
		if isBizPointer(e) {
			return fmt.Sprintf("Proto%sToBiz(p.%s)", e.Type.Name, protoStructField(e))
		}
	} else {
		if isBizPointer(e) && isProtoID(e) {
			if e.Unique {
				idAccess := fmt.Sprintf("p.%s", protoStructField(e))
				return fmt.Sprintf("&biz.%s{%sBase: biz.%sBase{UUID: %s}}",
					e.Type.Name, e.Type.Name, e.Type.Name, idAccess)
			}
			return "nil"
		}

		if isBizIDOnly(e) {
			typ := e.Type.ID.Type.String()
			fieldName := protoStructField(e)

			if typ == "uuid.UUID" {
				return fmt.Sprintf("p.%s", fieldName)
			}
			if typ == "int" || typ == "int32" {
				return fmt.Sprintf("int(p.%s)", fieldName)
			}
			return fmt.Sprintf("p.%s", fieldName)
		}
		return zeroValue("ptr")
	}
	return "nil"
}

func enumToProtoFuncName(f *entgen.Field, nodeName string) string {
	return fmt.Sprintf("Biz%s%sToProto", nodeName, f.StructField())
}

func enumFromProtoFuncName(f *entgen.Field, nodeName string) string {
	return fmt.Sprintf("Proto%s%sToBiz", nodeName, f.StructField())
}

// --- NEW Safe Mapping Helper Functions ---

func requiresErrorCheck(f *entgen.Field, mode string) bool {
	// mode: "ProtoToBiz" or "BizToEnt"
	if mode == "ProtoToBiz" {
		targetType := bizFieldType(f)
		if targetType == "" {
			targetType = f.Type.String()
		}
		if f.Type.String() == "uuid.UUID" && targetType != "string" {
			return true
		}
	}
	if mode == "BizToEnt" {
		entType := f.Type.String()
		// Implicit Biz Type for UUID is string, so we need conversion if Ent is UUID
		if entType == "uuid.UUID" {
			// If implicit biz type is used, it returns "" from explicitBizType
			bizType := explicitBizType(f)
			if bizType == "" || bizType == "string" {
				return true
			}
		}
		// Time Parsing Check
		if entType == "time.Time" {
			bizType := explicitBizType(f)
			if bizType == "string" {
				return true
			}
		}
	}
	return false
}

// Generates code block for conversion with error check
func convertFromProtoSetup(f *entgen.Field, nodeName string) string {
	if isSensitive(f) {
		return ""
	}
	if !requiresErrorCheck(f, "ProtoToBiz") {
		return ""
	}

	// Gen safe uuid parse
	// var <Name>Val uuid.UUID
	// var err error
	// <Name>Val, err = uuid.Parse(p.<ProtoName>)
	varName := camel(f.StructField()) + "Val"
	return fmt.Sprintf("%s, err := uuid.Parse(p.%s)\nif err != nil {\n\treturn nil, fmt.Errorf(\"invalid UUID for %s: %%w\", err)\n}", varName, protoGoName(f), f.Name)
}

func convertFromProtoUsage(f *entgen.Field, nodeName string) string {
	if isSensitive(f) {
		return zeroValue(bizFieldType(f))
	}
	if requiresErrorCheck(f, "ProtoToBiz") {
		return camel(f.StructField()) + "Val"
	}
	return convertFromProto(f, nodeName)
}

func convertBizToEntSetup(f *entgen.Field, nodeName string) string {
	if !requiresErrorCheck(f, "BizToEnt") {
		return ""
	}

	// Code block
	varName := camel(f.StructField()) + "EntVal"
	bizExpr := fmt.Sprintf("b.%s", bizFieldName(f))

	// Special handling for Nillable/Optional UUID (implied string in Biz)
	if f.Nillable && f.Type.String() == "uuid.UUID" {
		return fmt.Sprintf("var %s *uuid.UUID\nif %s != \"\" {\n\tparsed, err := uuid.Parse(%s)\n\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"invalid UUID for %s: %%w\", err)\n\t}\n\t%s = &parsed\n}", varName, bizExpr, bizExpr, f.Name, varName)
	}

	if f.Type.String() == "uuid.UUID" {
		return fmt.Sprintf("%s, err := uuid.Parse(%s)\nif err != nil {\n\treturn nil, fmt.Errorf(\"invalid UUID for %s: %%w\", err)\n}", varName, bizExpr, f.Name)
	}

	if f.Type.String() == "time.Time" && explicitBizType(f) == "string" {
		if f.Nillable {
			// Optional Time
			return fmt.Sprintf("var %s *time.Time\nif %s != \"\" {\n\tparsed, err := time.Parse(time.RFC3339, %s)\n\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"invalid Time format for %s: %%w\", err)\n\t}\n\t%s = &parsed\n}", varName, bizExpr, bizExpr, f.Name, varName)
		}
		return fmt.Sprintf("%s, err := time.Parse(time.RFC3339, %s)\nif err != nil {\n\treturn nil, fmt.Errorf(\"invalid Time format for %s: %%w\", err)\n}", varName, bizExpr, f.Name)
	}

	return ""
}

func convertBizToEntUsage(f *entgen.Field, nodeName string) string {
	if requiresErrorCheck(f, "BizToEnt") {
		varName := camel(f.StructField()) + "EntVal"
		// If Nillable, Setup created a *UUID/*Time var, so just use it.
		if f.Nillable {
			// For both UUID and Time (string)
			return varName
		}

		isPtr := f.Nillable
		if isPtr {
			return "&" + varName
		}
		return varName
	}
	return convertBizToEnt(f, nodeName, fmt.Sprintf("b.%s", bizFieldName(f)))
}
