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
	"getProtoTag":          func(f *gen.Field, i int) int { return i + 1 },
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
}

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

func edgeField(e *gen.Edge) string {
	return e.StructField() + "ID"
}

func hasField(fields []*gen.Field, name string) bool {
	for _, f := range fields {
		if f.StructField() == name {
			return true
		}
	}
	return false
}

func edgeIDType(e *gen.Edge) string {
	return e.Type.ID.Type.String()
}

func edgeProtoType(e *gen.Edge) string {
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

func edgeConvertFromProto(e *gen.Edge) string {
	pbField := e.StructField() + "Id"
	typ := e.Type.ID.Type.String()
	if typ == "uuid.UUID" {
		return fmt.Sprintf("uuid.MustParse(p.%s)", pbField)
	}
	if typ == "int" || typ == "int32" {
		return fmt.Sprintf("int(p.%s)", pbField)
	}
	return "p." + pbField
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
	default:
		return "nil"
	}
}

func getEnumValues(f *gen.Field) map[string]int32 {
	if f.Annotations == nil {
		return nil
	}
	if v, ok := f.Annotations["LazyEnt"]; ok {
		if a, ok := v.(Annotation); ok {
			return a.EnumValues
		}
		if m, ok := v.(map[string]interface{}); ok {
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
					return res
				}
			}
		}
	}
	return nil
}

func protoType(f *gen.Field) string {
	if f.IsEnum() {
		return f.StructField()
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
	default:
		return "string"
	}
}

func getValidateRules(f *gen.Field) string {
	return "{}"
}

func convertToProto(f *gen.Field) string {
	if f.IsEnum() {
		return fmt.Sprintf("pb.%s(b.%s)", f.StructField(), f.StructField())
	}
	if f.Type.String() == "time.Time" {
		return fmt.Sprintf("timestamppb.New(b.%s)", f.StructField())
	}
	return "b." + f.StructField()
}

func convertFromProto(f *gen.Field) string {
	if f.IsEnum() {
		return fmt.Sprintf("biz.%s(p.%s)", f.StructField(), f.StructField())
	}
	if f.Type.String() == "time.Time" {
		return fmt.Sprintf("p.%s.AsTime()", f.StructField())
	}
	return "p." + f.StructField()
}
