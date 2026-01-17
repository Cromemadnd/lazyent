package gen

import (
	"fmt"
	"strings"
)

func convertToProto(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	if isFieldProtoExclude(f, false) {
		return zeroValue(getProtoType(f))
	}
	if f.IsEnum() {
		if isExternalEnum(f.Field) {
			return fmt.Sprintf("string(b.%s)", bizFieldName(f))
		}
		return fmt.Sprintf("%s(b.%s)", enumToProtoFuncName(f, nodeName), bizFieldName(f))
	}
	if f.Type.String() == "time.Time" {
		if getProtoType(f) == "int64" {
			return fmt.Sprintf("b.%s.UnixMilli()", bizFieldName(f))
		}
		if getProtoType(f) == "uint64" {
			return fmt.Sprintf("uint64(b.%s.UnixMilli())", bizFieldName(f))
		}
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
	} else if pt == "float" {
		goProtoType = "float32"
	} else if pt == "bytes" {
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

func convertFromProto(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	if isFieldProtoExclude(f, true) {
		return zeroValue(bizFieldType(f))
	}
	// For safer generation, we should use convertFromProtoUsage after generating Setup code.

	if f.IsEnum() {
		if isExternalEnum(f.Field) {
			return fmt.Sprintf("%s(p.%s)", getExternalEnumName(f), protoGoName(f))
		}
		return fmt.Sprintf("%s(p.%s)", enumFromProtoFuncName(f, nodeName), protoGoName(f))
	}
	if f.Type.String() == "time.Time" {
		// Use getProtoType via helper to check override
		pt := getProtoType(f)
		if pt == "int64" || pt == "uint64" {
			return fmt.Sprintf("time.UnixMilli(int64(p.%s))", protoGoName(f))
		}
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

func convertEntToBiz(v interface{}, nodeName string, expr string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}

	if f.IsEnum() {
		if isExternalEnum(f.Field) {
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
		castExpr = convertTimeEntToBiz(exprVal, bizType)
	} else if entType == "uuid.UUID" {
		castExpr = convertUUIDEntToBiz(expr, bizType)
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

func convertBizToEnt(v interface{}, nodeName string, expr string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}

	if f.IsEnum() {
		if isExternalEnum(f.Field) {
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
		entExpr = convertTimeBizToEnt(expr, bizType, entType)
	} else if entType == "uuid.UUID" {
		entExpr = convertUUIDBizToEnt(expr, bizType, entType)
	} else {
		entExpr = fmt.Sprintf("%s(%s)", entType, expr)
	}

	isPtr := f.Nillable
	if isPtr {
		return fmt.Sprintf("func() *%s { x := %s; return &x }()", entType, entExpr)
	}

	return entExpr
}

func edgeConvertToProto(v interface{}) string {
	e := asGenEdge(v)
	if e == nil {
		return "nil"
	}

	if isProtoMessage(e, false) {
		if isBizPointer(e, false) {
			return fmt.Sprintf("Biz%sToProto(b.%s)", e.Type.Name, bizEdgeName(e, false))
		}
	} else {
		if isBizIDOnly(e, false) {
			field := bizEdgeName(e, false)
			typ := e.Type.ID.Type.String()
			if typ == "uuid.UUID" {
				return fmt.Sprintf("b.%s", field)
			}
			if typ == "int" || typ == "int32" {
				return fmt.Sprintf("int32(b.%s)", field)
			}
			return "b." + field
		}
		access := fmt.Sprintf("b.%s.UUID", bizEdgeName(e, false))
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

func edgeConvertFromProto(v interface{}, in bool) string {
	e := asGenEdge(v)
	if e == nil {
		return "nil"
	}

	if isProtoMessage(e, in) {
		if isBizPointer(e, in) {
			return fmt.Sprintf("Proto%sToBiz(p.%s)", e.Type.Name, protoStructField(e, in))
		}
	} else {
		// Case: ProtoID -> BizPointer
		if isBizPointer(e, in) && isProtoID(e, in) {
			idField := bizFieldName(e.Type.ID)
			if e.Unique {
				idAccess := fmt.Sprintf("p.%s", protoStructField(e, in))
				return fmt.Sprintf("&biz.%s{%sBase: biz.%sBase{%s: %s}}",
					e.Type.Name, e.Type.Name, e.Type.Name, idField, idAccess)
			} else {
				// For list element (inside loop, variable is 'item')
				return fmt.Sprintf("&biz.%s{%sBase: biz.%sBase{%s: item}}",
					e.Type.Name, e.Type.Name, e.Type.Name, idField)
			}
		}

		if isBizIDOnly(e, in) {
			typ := e.Type.ID.Type.String()
			fieldName := protoStructField(e, in)
			if typ == "uuid.UUID" {
				return fmt.Sprintf("p.%s", fieldName)
			}
			if typ == "int" || typ == "int32" {
				return fmt.Sprintf("int(p.%s)", fieldName)
			}
			return fmt.Sprintf("p.%s", fieldName)
		}
	}
	return "nil"
}

func enumToProtoFuncName(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	return fmt.Sprintf("Biz%s%sToProto", nodeName, f.StructField())
}

func enumFromProtoFuncName(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	return fmt.Sprintf("Proto%s%sToBiz", nodeName, f.StructField())
}

// --- NEW Safe Mapping Helper Functions ---

func requiresErrorCheck(v interface{}, mode string) bool {
	f := asGenField(v)
	if f == nil {
		return false
	}
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
		if entType == "uuid.UUID" {
			bizType := explicitBizType(f)
			if bizType == "" || bizType == "string" {
				return true
			}
		}
		if entType == "time.Time" {
			bizType := explicitBizType(f)
			if bizType == "string" {
				return true
			}
		}
	}
	return false
}

func convertFromProtoSetup(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	if isFieldProtoExclude(f, true) {
		return ""
	}
	if !requiresErrorCheck(f, "ProtoToBiz") {
		return ""
	}

	varName := camel(f.StructField()) + "Val"
	return fmt.Sprintf("%s, err := uuid.Parse(p.%s)\nif err != nil {\n\treturn nil, fmt.Errorf(\"invalid UUID for %s: %%w\", err)\n}", varName, protoGoName(f), f.Name)
}

func convertFromProtoUsage(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	if isFieldProtoExclude(f, true) {
		return zeroValue(bizFieldType(f))
	}
	if requiresErrorCheck(f, "ProtoToBiz") {
		return camel(f.StructField()) + "Val"
	}
	return convertFromProto(f, nodeName)
}

func convertBizToEntSetup(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	if !requiresErrorCheck(f, "BizToEnt") {
		return ""
	}

	varName := camel(f.StructField()) + "EntVal"
	bizExpr := fmt.Sprintf("b.%s", bizFieldName(f))

	if f.Nillable && f.Type.String() == "uuid.UUID" {
		return fmt.Sprintf("var %s *uuid.UUID\nif %s != \"\" {\n\tparsed, err := uuid.Parse(%s)\n\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"invalid UUID for %s: %%w\", err)\n\t}\n\t%s = &parsed\n}", varName, bizExpr, bizExpr, f.Name, varName)
	}

	if f.Type.String() == "uuid.UUID" {
		return fmt.Sprintf("var %s uuid.UUID\nif %s != \"\" {\n\tval, err := uuid.Parse(%s)\n\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"invalid UUID for %s: %%w\", err)\n\t}\n\t%s = val\n}", varName, bizExpr, bizExpr, f.Name, varName)
	}

	if f.Type.String() == "time.Time" && explicitBizType(f) == "string" {
		if f.Nillable {
			return fmt.Sprintf("var %s *time.Time\nif %s != \"\" {\n\tparsed, err := time.Parse(time.RFC3339, %s)\n\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"invalid Time format for %s: %%w\", err)\n\t}\n\t%s = &parsed\n}", varName, bizExpr, bizExpr, f.Name, varName)
		}
		return fmt.Sprintf("%s, err := time.Parse(time.RFC3339, %s)\nif err != nil {\n\treturn nil, fmt.Errorf(\"invalid Time format for %s: %%w\", err)\n}", varName, bizExpr, f.Name)
	}

	return ""
}

func convertBizToEntUsage(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	if requiresErrorCheck(f, "BizToEnt") {
		varName := camel(f.StructField()) + "EntVal"
		if f.Nillable {
			return varName
		}
		if f.Nillable { // Logic error in original?
			// Original Line 393: isPtr := f.Nillable
			// Line 394: if isPtr { return "&" + varName }
			// But line 388 says: if f.Nillable { return varName }.
			// One branch must be redundant or specific.
			// Re-reading original:
			/*
				if f.Nillable { return varName }
				isPtr := f.Nillable (true)
				if isPtr { ... }
			*/
			// So line 394 was unreachable?
			// I will preserve my reading of it.
		}
		isPtr := f.Nillable
		if isPtr {
			return "&" + varName
		}
		return varName
	}
	return convertBizToEnt(f, nodeName, fmt.Sprintf("b.%s", bizFieldName(f)))
}

// --- Internal Helper Functions ---

func convertTimeEntToBiz(expr string, bizType string) string {
	if bizType == "int64" {
		return fmt.Sprintf("%s.Unix()", expr)
	} else if bizType == "string" {
		return fmt.Sprintf("%s.Format(time.RFC3339)", expr)
	}
	return fmt.Sprintf("%s(%s)", bizType, expr)
}

func convertUUIDEntToBiz(expr string, bizType string) string {
	if bizType == "string" {
		return fmt.Sprintf("%s.String()", expr)
	}
	return fmt.Sprintf("%s(%s)", bizType, expr)
}

func convertTimeBizToEnt(expr string, bizType string, entType string) string {
	if bizType == "int64" {
		return fmt.Sprintf("time.Unix(%s, 0)", expr)
	} else if bizType == "string" {
		return fmt.Sprintf("func() time.Time { t, _ := time.Parse(time.RFC3339, %s); return t }()", expr)
	}
	return fmt.Sprintf("%s(%s)", entType, expr)
}

func convertUUIDBizToEnt(expr string, bizType string, entType string) string {
	if bizType == "string" {
		// Dangerous MustParse: ensure validation is done before this point, or use setup/usage functions
		return fmt.Sprintf("uuid.MustParse(%s)", expr)
	}
	return fmt.Sprintf("%s(%s)", entType, expr)
}
