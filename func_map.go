package lazyent

import (
	"fmt"
	"strings"
	"text/template"

	"entgo.io/ent/entc/gen"
)

var funcMap = template.FuncMap{
	"getEnumValues":        getEnumValues,
	"protoType":            protoType,
	"getValidateRules":     getValidateRules,
	"getProtoTag":          getProtoTag,
	"convertToProto":       convertToProto,
	"convertFromProto":     convertFromProto,
	"add":                  func(a, b int) int { return a + b },
	"lower":                strings.ToLower,
	"hasTime":              hasTime,
	"hasUUID":              hasUUID,
	"hasTimeNodes":         hasTimeNodes,
	"hasUUIDNodes":         hasUUIDNodes,
	"edgeHasFK":            edgeHasFK,
	"edgeField":            edgeField,
	"hasField":             hasField,
	"edgeIDType":           edgeIDType,
	"edgeProtoType":        edgeProtoType,
	"edgeConvertToProto":   edgeConvertToProto,
	"edgeConvertFromProto": edgeConvertFromProto,
	"zeroValue":            zeroValue,
	"validateConflict":     validateConflict,
	"bizStrategy":          getBizStrategy,
	"protoStrategy":        getProtoStrategy,
	"isBizIDOnly":          isBizIDOnly,
	"isBizExclude":         isBizExclude,
	"isBizPointer":         isBizPointer,
	"isBizValue":           isBizValue,
	"isProtoID":            isProtoID,
	"isProtoMessage":       isProtoMessage,
	"isProtoExclude":       isProtoExclude,
	"enumToProtoFunc":      enumToProtoFuncName,
	"enumFromProtoFunc":    enumFromProtoFuncName,
	"getAllEnums":          getAllEnums,
}

// ... hasTimeNodes, hasUUIDNodes, hasTime, hasUUID, edgeHasFK, edgeField, hasField, edgeIDType ...
// (Omitting unchanged simple helpers for brevity in thought, but must include in file)
// To save context, I'll include all.

func hasTimeNodes(nodes []interface{}) bool {
	for _, node := range nodes {
		n, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		fields, ok := n["Fields"].([]*gen.Field)
		if !ok {
			continue
		}
		for _, f := range fields {
			if f.Type.String() == "time.Time" {
				return true
			}
		}
	}
	return false
}

func hasUUIDNodes(nodes []interface{}) bool {
	for _, node := range nodes {
		n, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		fields, _ := n["Fields"].([]*gen.Field)
		edges, _ := n["Edges"].([]*gen.Edge)
		if hasUUID(fields, edges) {
			return true
		}
	}
	return false
}

func hasTime(fields []*gen.Field) bool {
	for _, f := range fields {
		if f.Type.String() == "time.Time" {
			return true
		}
	}
	return false
}

func hasUUID(fields []*gen.Field, edges []*gen.Edge) bool {
	for _, f := range fields {
		if f.Type.String() == "uuid.UUID" {
			return true
		}
	}
	for _, e := range edges {
		if e.Type.ID.Type.String() == "uuid.UUID" {
			return true
		}
	}
	return false
}

func edgeHasFK(e *gen.Edge) bool {
	if !e.IsInverse() {
		return e.Unique
	}
	if e.Ref != nil && !e.Ref.Unique {
		return e.Unique
	}
	return false
}

func edgeField(e *gen.Edge) string { return e.StructField() + "ID" }

func hasField(fields []*gen.Field, name string) bool {
	for _, f := range fields {
		if f.StructField() == name {
			return true
		}
	}
	return false
}

func edgeIDType(e *gen.Edge) string { return e.Type.ID.Type.String() }

func edgeProtoType(e *gen.Edge) string {
	if isProtoMessage(e) {
		return e.Type.Name
	}
	t := e.Type.ID.Type.String()
	switch t {
	case "int", "int32":
		return "int32"
	case "int64", "uint64":
		return "int64"
	case "string":
		return "string"
	default:
		return "string"
	}
}

func edgeConvertToProto(e *gen.Edge) string {
	if err := validateConflict(e); err != nil {
		panic(err)
	}
	if isProtoMessage(e) {
		if isBizPointer(e) {
			return fmt.Sprintf("ToProto%s(b.%s)", e.Type.Name, e.StructField())
		}
		if isBizValue(e) {
			return fmt.Sprintf("ToProto%s(&b.%s)", e.Type.Name, e.StructField())
		}
	} else {
		if isBizIDOnly(e) {
			field := e.StructField() + "ID"
			typ := e.Type.ID.Type.String()
			if typ == "uuid.UUID" {
				return fmt.Sprintf("b.%s.String()", field)
			}
			if typ == "int" || typ == "int32" {
				return fmt.Sprintf("int32(b.%s)", field)
			}
			return "b." + field
		}
		// BizPtr/Value -> ProtoID
		access := fmt.Sprintf("b.%s.ID", e.StructField())
		typ := e.Type.ID.Type.String()
		if typ == "uuid.UUID" {
			return fmt.Sprintf("%s.String()", access)
		}
		if typ == "int" || typ == "int32" {
			return fmt.Sprintf("int32(%s)", access)
		}
		return access
	}
	return "nil"
}

func edgeConvertFromProto(e *gen.Edge) string {
	if isProtoMessage(e) {
		if isBizPointer(e) {
			return fmt.Sprintf("FromProto%s(p.%s)", e.Type.Name, e.Name)
		}
		if isBizValue(e) {
			return fmt.Sprintf("*FromProto%s(p.%s)", e.Type.Name, e.Name)
		}
	} else {
		if isBizIDOnly(e) {
			typ := e.Type.ID.Type.String()
			if typ == "uuid.UUID" {
				return fmt.Sprintf("uuid.MustParse(p.%s_id)", e.Name)
			}
			if typ == "int" || typ == "int32" {
				return fmt.Sprintf("int(p.%s_id)", e.Name)
			}
			return fmt.Sprintf("p.%s_id", e.Name)
		}
		return zeroValue("ptr")
	}
	return "nil"
}

func zeroValue(t string) string {
	switch t {
	case "int", "int32", "int64", "uint64", "float64", "float32":
		return "0"
	case "string":
		return `""`
	case "bool":
		return "false"
	case "uuid.UUID":
		return "uuid.Nil"
	case "time.Time":
		return "time.Time{}"
	case "ptr":
		return "nil"
	default:
		return "nil"
	}
}

// Updated to accept Node Name to ensure uniqueness of Enums
func protoType(f *gen.Field, nodeName string) string {
	if f.IsEnum() {
		return nodeName + f.StructField()
	}
	switch f.Type.String() {
	case "int", "int32":
		return "int32"
	case "int64", "uint64":
		return "int64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "time.Time":
		return "google.protobuf.Timestamp"
	case "float64":
		return "double"
	case "float32":
		return "float"
	case "uuid.UUID":
		return "string"
	default:
		return "string"
	}
}

func getValidateRules(f *gen.Field) string {
	a := getFieldAnnotation(f)
	if a != nil && a.Validate != "" {
		return a.Validate
	}
	return "{}"
}

func getProtoTag(f *gen.Field, i int) int {
	a := getFieldAnnotation(f)
	if a != nil && a.FieldID > 0 {
		return int(a.FieldID)
	}
	return i + 1
}

// Updated to accept Node Name
func convertToProto(f *gen.Field, nodeName string) string {
	if f.IsEnum() {
		return fmt.Sprintf("%s(b.%s)", enumToProtoFuncName(f, nodeName), f.StructField())
	}
	if f.Type.String() == "time.Time" {
		return fmt.Sprintf("timestamppb.New(b.%s)", f.StructField())
	}
	if f.Type.String() == "uuid.UUID" {
		return fmt.Sprintf("b.%s.String()", f.StructField())
	}
	return "b." + f.StructField()
}

// Updated to accept Node Name
func convertFromProto(f *gen.Field, nodeName string) string {
	if f.IsEnum() {
		return fmt.Sprintf("%s(p.%s)", enumFromProtoFuncName(f, nodeName), f.StructField())
	}
	if f.Type.String() == "time.Time" {
		return fmt.Sprintf("p.%s.AsTime()", f.StructField())
	}
	if f.Type.String() == "uuid.UUID" {
		return fmt.Sprintf("uuid.MustParse(p.%s)", f.StructField())
	}
	return "p." + f.StructField()
}

func getEnumValues(f *gen.Field) map[string]int32 {
	a := getFieldAnnotation(f)
	if a != nil {
		return a.EnumValues
	}
	return nil
}

func getFieldAnnotation(f *gen.Field) *Annotation {
	if f.Annotations == nil {
		return nil
	}
	if v, ok := f.Annotations["lazyent"]; ok {
		if a, ok := v.(Annotation); ok {
			return &a
		}
		// Decode from map if needed
		if m, ok := v.(map[string]interface{}); ok {
			a := &Annotation{}
			if ev, ok := m["enum_values"]; ok {
				if evMap, ok := ev.(map[string]interface{}); ok {
					res := make(map[string]int32)
					for k, val := range evMap {
						if fVal, ok := val.(float64); ok {
							res[k] = int32(fVal)
						} else if iVal, ok := val.(int); ok {
							res[k] = int32(iVal)
						}
					}
					a.EnumValues = res
				}
			}
			// Parse FieldID
			if fid, ok := m["field_id"]; ok {
				if fVal, ok := fid.(float64); ok {
					a.FieldID = int32(fVal)
				} else if iVal, ok := fid.(int); ok {
					a.FieldID = int32(iVal)
				}
			}
			// Parse Validate
			if vRule, ok := m["validate"]; ok {
				if sVal, ok := vRule.(string); ok {
					a.Validate = sVal
				}
			}
			return a
		}
	}
	return nil
}

// Strategy Helpers (retained)
func getAnnotation(e *gen.Edge) *Annotation {
	if e.Annotations == nil {
		return nil
	}
	if v, ok := e.Annotations["lazyent"]; ok {
		if a, ok := v.(Annotation); ok {
			return &a
		}
		if m, ok := v.(map[string]interface{}); ok {
			a := &Annotation{}
			if bs, ok := m["biz_strategy"]; ok {
				if f, ok := bs.(float64); ok {
					a.BizStrategy = BizStrategy(int(f))
				} else if i, ok := bs.(int); ok {
					a.BizStrategy = BizStrategy(i)
				}
			}
			if ps, ok := m["proto_strategy"]; ok {
				if f, ok := ps.(float64); ok {
					a.ProtoStrategy = ProtoStrategy(int(f))
				} else if i, ok := ps.(int); ok {
					a.ProtoStrategy = ProtoStrategy(i)
				}
			}
			return a
		}
	}
	return nil
}

func getBizStrategy(e *gen.Edge) BizStrategy {
	a := getAnnotation(e)
	if a != nil {
		return a.BizStrategy
	}
	return BizPointer
}

func getProtoStrategy(e *gen.Edge) ProtoStrategy {
	a := getAnnotation(e)
	if a != nil {
		return a.ProtoStrategy
	}
	return ProtoID
}

func isBizIDOnly(e *gen.Edge) bool    { return getBizStrategy(e) == BizIDOnly }
func isBizExclude(e *gen.Edge) bool   { return getBizStrategy(e) == BizExclude }
func isBizPointer(e *gen.Edge) bool   { return getBizStrategy(e) == BizPointer }
func isBizValue(e *gen.Edge) bool     { return getBizStrategy(e) == BizValue }
func isProtoID(e *gen.Edge) bool      { return getProtoStrategy(e) == ProtoID }
func isProtoMessage(e *gen.Edge) bool { return getProtoStrategy(e) == ProtoMessage }
func isProtoExclude(e *gen.Edge) bool { return getProtoStrategy(e) == ProtoExclude }

func validateConflict(e *gen.Edge) error {
	if isBizIDOnly(e) && isProtoMessage(e) {
		return fmt.Errorf("conflict: Edge '%s' has BizIDOnly but ProtoMessage. Cannot generate mapper.", e.Name)
	}
	return nil
}

func enumToProtoFuncName(f *gen.Field, nodeName string) string {
	return fmt.Sprintf("%s%sToProto", nodeName, f.StructField())
}

func enumFromProtoFuncName(f *gen.Field, nodeName string) string {
	return fmt.Sprintf("%s%sFromProto", nodeName, f.StructField())
}

// Special struct for Enum to pass to template
type EnumDef struct {
	NodeName string
	Field    *gen.Field
}

func getAllEnums(nodes []interface{}) []EnumDef {
	var enums []EnumDef
	for _, node := range nodes {
		n, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		name := n["Name"].(string)
		fields, ok := n["Fields"].([]*gen.Field)
		if !ok {
			continue
		}
		for _, f := range fields {
			if f.IsEnum() {
				enums = append(enums, EnumDef{NodeName: name, Field: f})
			}
		}
	}
	return enums
}
