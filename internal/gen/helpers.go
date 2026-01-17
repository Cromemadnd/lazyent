package gen

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	entgen "entgo.io/ent/entc/gen"
	types "github.com/Cromemadnd/lazyent/internal/types"
)

func hasTimeNodes(nodes []interface{}) bool {
	for _, node := range nodes {
		n, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		fields, ok := n["Fields"].([]*entgen.Field)
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
		fields, _ := n["Fields"].([]*entgen.Field)
		edges, _ := n["Edges"].([]*entgen.Edge)
		if hasUUID(fields, edges) {
			return true
		}
	}
	return false
}

func hasTime(fields []*entgen.Field) bool {
	for _, f := range fields {
		if f.Type.String() == "time.Time" {
			return true
		}
	}
	return false
}

func hasUUID(fields []*entgen.Field, edges []*entgen.Edge) bool {
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

func edgeHasFK(e *entgen.Edge) bool {
	if !e.IsInverse() {
		return e.Unique
	}
	if e.Ref != nil && !e.Ref.Unique {
		return e.Unique
	}
	return false
}

func edgeField(e *entgen.Edge, in bool) string { return bizEdgeName(e, in) }

func hasField(fields []*entgen.Field, name string) bool {
	for _, f := range fields {
		if f.StructField() == name {
			return true
		}
	}
	return false
}

func edgeIDType(e *entgen.Edge) string { return e.Type.ID.Type.String() }

func edgeProtoType(e *entgen.Edge, in bool) string {
	if isProtoMessage(e, in) {
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

func protoType(f *entgen.Field, nodeName string) string {
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

func getProtoType(f *entgen.Field) string {
	if f == nil {
		return "string"
	}
	a := getFieldAnnotation(f)
	if a != nil && a.ProtoType != "" {
		return a.ProtoType
	}
	t := f.Type.String()
	switch t {
	case "int", "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint64":
		return "uint64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "float64":
		return "double"
	case "float32":
		return "float"
	case "uuid.UUID":
		return "string"
	}
	if strings.HasPrefix(t, "[]") {
		return "string"
	}
	return "string"
}

func getProtoTag(f *entgen.Field, i int) int {
	a := getFieldAnnotation(f)
	if a != nil && a.ProtoFieldID > 0 {
		return int(a.ProtoFieldID)
	}
	return i + 1
}

func protoFieldName(f *entgen.Field) string {
	a := getFieldAnnotation(f)
	if a != nil && a.ProtoName != "" {
		return a.ProtoName
	}
	return camelToSnake(f.Name)
}

func protoEdgeFieldName(e *entgen.Edge, in bool) string {
	a := getAnnotation(e)
	if a != nil && a.ProtoName != "" {
		return a.ProtoName
	}
	if isProtoID(e, in) {
		return camelToSnake(e.Name) + "_id"
	}
	return camelToSnake(e.Name)
}

func camelToSnake(s string) string {
	var results []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			results = append(results, '_')
			results = append(results, r+'a'-'A')
		} else {
			results = append(results, r)
		}
	}
	return strings.ToLower(string(results))
}

func protoGoName(f *entgen.Field) string {
	if f == nil {
		return ""
	}
	a := getFieldAnnotation(f)
	if a != nil && a.ProtoName != "" {
		return pascal(a.ProtoName)
	}
	return pascal(f.Name)
}

func getEnumValues(f *entgen.Field) map[string]int32 {
	a := getFieldAnnotation(f)
	if a == nil || a.EnumValues == nil {
		if len(f.Enums) > 0 {
			m := make(map[string]int32)
			for i, e := range f.Enums {
				m[e.Value] = int32(i)
			}
			return m
		}
		return nil
	}
	if len(f.Enums) > 0 {
		return a.EnumValues
	}
	return a.EnumValues
}

type EnumPair struct {
	Key   string
	Value int32
}

func getEnumPairs(f *entgen.Field) []EnumPair {
	a := getFieldAnnotation(f)
	var pairs []EnumPair

	if a == nil || a.EnumValues == nil {
		for i, e := range f.Enums {
			pairs = append(pairs, EnumPair{Key: e.Value, Value: int32(i)})
		}
		return pairs
	}

	for _, e := range f.Enums {
		if v, ok := a.EnumValues[e.Value]; ok {
			pairs = append(pairs, EnumPair{Key: e.Value, Value: v})
		}
	}
	return pairs
}

func getFieldAnnotation(f *entgen.Field) *types.Annotation {
	if f == nil {
		return nil
	}
	if f.Annotations == nil {
		return nil
	}
	if v, ok := f.Annotations["LazyEnt"]; ok {
		if a, ok := v.(types.Annotation); ok {
			return &a
		}
		if m, ok := v.(map[string]interface{}); ok {
			return decodeAnnotationMap(m)
		}
	}
	if v, ok := f.Annotations["lazyent"]; ok {
		if a, ok := v.(types.Annotation); ok {
			return &a
		}
		if m, ok := v.(map[string]interface{}); ok {
			return decodeAnnotationMap(m)
		}
	}
	return nil
}

func decodeAnnotationMap(m map[string]interface{}) *types.Annotation {
	a := &types.Annotation{}
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
	if fid, ok := m["proto_field_id"]; ok {
		a.ProtoFieldID = int32(toInt(fid))
	} else if fid, ok := m["ProtoFieldID"]; ok {
		a.ProtoFieldID = int32(toInt(fid))
	}
	if fid, ok := m["field_id"]; ok {
		a.ProtoFieldID = int32(toInt(fid))
	}

	if v, ok := m["biz_name"]; ok {
		a.BizName, _ = v.(string)
	} else if v, ok := m["BizName"]; ok {
		a.BizName, _ = v.(string)
	}

	if v, ok := m["biz_type"]; ok {
		a.BizType, _ = v.(string)
	} else if v, ok := m["BizType"]; ok {
		a.BizType, _ = v.(string)
	}

	if v, ok := m["proto_name"]; ok {
		a.ProtoName, _ = v.(string)
	} else if v, ok := m["ProtoName"]; ok {
		a.ProtoName, _ = v.(string)
	}

	if v, ok := m["proto_type"]; ok {
		a.ProtoType, _ = v.(string)
	} else if v, ok := m["ProtoType"]; ok {
		a.ProtoType, _ = v.(string)
	}
	if v, ok := m["virtual"]; ok {
		a.Virtual, _ = v.(bool)
	} else if v, ok := m["Virtual"]; ok {
		a.Virtual, _ = v.(bool)
	}

	if v, ok := m["proto_validation"]; ok {
		a.ProtoValidation, _ = v.(string)
	} else if v, ok := m["ProtoValidation"]; ok {
		a.ProtoValidation, _ = v.(string)
	}

	// Strategies
	if v, ok := m["edge_in_strategy"]; ok {
		a.EdgeInStrategy = types.EdgeStrategy(toInt(v))
	} else if v, ok := m["EdgeInStrategy"]; ok {
		a.EdgeInStrategy = types.EdgeStrategy(toInt(v))
	}

	if v, ok := m["edge_out_strategy"]; ok {
		a.EdgeOutStrategy = types.EdgeStrategy(toInt(v))
	} else if v, ok := m["EdgeOutStrategy"]; ok {
		a.EdgeOutStrategy = types.EdgeStrategy(toInt(v))
	}

	if v, ok := m["field_in_strategy"]; ok {
		a.FieldInStrategy = types.FieldStrategy(toInt(v))
	} else if v, ok := m["FieldInStrategy"]; ok {
		a.FieldInStrategy = types.FieldStrategy(toInt(v))
	}

	if v, ok := m["field_out_strategy"]; ok {
		a.FieldOutStrategy = types.FieldStrategy(toInt(v))
	} else if v, ok := m["FieldOutStrategy"]; ok {
		a.FieldOutStrategy = types.FieldStrategy(toInt(v))
	}

	// Validation
	if err := a.EdgeInStrategy.Validate(); err != nil {
		panic(fmt.Sprintf("EdgeInStrategy conflict in field annotation: %v", err))
	}
	if err := a.EdgeOutStrategy.Validate(); err != nil {
		panic(fmt.Sprintf("EdgeOutStrategy conflict in field annotation: %v", err))
	}
	if err := a.FieldInStrategy.Validate(); err != nil {
		panic(fmt.Sprintf("FieldInStrategy conflict in field annotation: %v", err))
	}
	if err := a.FieldOutStrategy.Validate(); err != nil {
		panic(fmt.Sprintf("FieldOutStrategy conflict in field annotation: %v", err))
	}

	return a
}

func getAnnotation(e *entgen.Edge) *types.Annotation {
	if e.Annotations == nil {
		return nil
	}
	if v, ok := e.Annotations["LazyEnt"]; ok {
		if a, ok := v.(types.Annotation); ok {
			return &a
		}
		if m, ok := v.(map[string]interface{}); ok {
			return decodeEdgeAnnotationMap(m)
		}
	}
	if v, ok := e.Annotations["lazyent"]; ok {
		if a, ok := v.(types.Annotation); ok {
			return &a
		}
		if m, ok := v.(map[string]interface{}); ok {
			return decodeEdgeAnnotationMap(m)
		}
	}
	return nil
}

func decodeEdgeAnnotationMap(m map[string]interface{}) *types.Annotation {
	a := &types.Annotation{}
	if v, ok := m["biz_name"]; ok {
		a.BizName, _ = v.(string)
	} else if v, ok := m["BizName"]; ok {
		a.BizName, _ = v.(string)
	}

	if v, ok := m["proto_name"]; ok {
		a.ProtoName, _ = v.(string)
	} else if v, ok := m["ProtoName"]; ok {
		a.ProtoName, _ = v.(string)
	}

	if v, ok := m["edge_in_strategy"]; ok {
		a.EdgeInStrategy = types.EdgeStrategy(toInt(v))
	} else if v, ok := m["EdgeInStrategy"]; ok {
		a.EdgeInStrategy = types.EdgeStrategy(toInt(v))
	}

	if v, ok := m["edge_out_strategy"]; ok {
		a.EdgeOutStrategy = types.EdgeStrategy(toInt(v))
	} else if v, ok := m["EdgeOutStrategy"]; ok {
		a.EdgeOutStrategy = types.EdgeStrategy(toInt(v))
	}

	// Validation
	if err := a.EdgeInStrategy.Validate(); err != nil {
		panic(fmt.Sprintf("EdgeInStrategy conflict in edge annotation: %v", err))
	}
	if err := a.EdgeOutStrategy.Validate(); err != nil {
		panic(fmt.Sprintf("EdgeOutStrategy conflict in edge annotation: %v", err))
	}

	return a
}

func getEdgeInStrategy(e *entgen.Edge) types.EdgeStrategy {
	a := getAnnotation(e)
	if a != nil && a.EdgeInStrategy != 0 {
		return a.EdgeInStrategy
	}
	return types.EdgeProtoMessage | types.EdgeBizPointer
}

func getEdgeOutStrategy(e *entgen.Edge) types.EdgeStrategy {
	a := getAnnotation(e)
	if a != nil && a.EdgeOutStrategy != 0 {
		return a.EdgeOutStrategy
	}
	return types.EdgeProtoMessage | types.EdgeBizPointer
}

func getFieldInStrategy(f *entgen.Field) types.FieldStrategy {
	a := getFieldAnnotation(f)
	if a != nil && a.FieldInStrategy != 0 {
		return a.FieldInStrategy
	}
	if isSensitive(f) {
		return types.FieldBizValue | types.FieldProtoRequired
	}
	if f.Optional {
		return types.FieldBizPointer | types.FieldProtoOptional
	}
	return types.FieldBizValue | types.FieldProtoRequired
}

func getFieldOutStrategy(f *entgen.Field) types.FieldStrategy {
	a := getFieldAnnotation(f)
	if a != nil && a.FieldOutStrategy != 0 {
		return a.FieldOutStrategy
	}
	if isSensitive(f) {
		return types.FieldBizExcluded | types.FieldProtoExcluded
	}
	if f.Optional {
		return types.FieldBizPointer | types.FieldProtoOptional
	}
	return types.FieldBizValue | types.FieldProtoRequired
}

func isBizIDOnly(e *entgen.Edge, in bool) bool {
	var s types.EdgeStrategy
	if in {
		s = getEdgeInStrategy(e)
	} else {
		s = getEdgeOutStrategy(e)
	}
	return (s & types.EdgeBizMask) == types.EdgeBizID
}

func isFieldBizExclude(f *entgen.Field, in bool) bool {
	var s types.FieldStrategy
	if in {
		s = getFieldInStrategy(f)
	} else {
		s = getFieldOutStrategy(f)
	}
	return (s & types.FieldBizMask) == types.FieldBizExcluded
}

func shouldGenerateBizField(f *entgen.Field) bool {
	return !isFieldBizExclude(f, true) || !isFieldBizExclude(f, false)
}

func isBizExclude(e *entgen.Edge, in bool) bool {
	var s types.EdgeStrategy
	if in {
		s = getEdgeInStrategy(e)
	} else {
		s = getEdgeOutStrategy(e)
	}
	return (s & types.EdgeBizMask) == types.EdgeBizExcluded
}

func shouldGenerateBizEdge(e *entgen.Edge) bool {
	return !isBizExclude(e, true) || !isBizExclude(e, false)
}

func isBizPointer(e *entgen.Edge, in bool) bool {
	var s types.EdgeStrategy
	if in {
		s = getEdgeInStrategy(e)
	} else {
		s = getEdgeOutStrategy(e)
	}
	return (s & types.EdgeBizMask) == types.EdgeBizPointer
}

func shouldBizPointer(e *entgen.Edge) bool {
	return isBizPointer(e, true) || isBizPointer(e, false)
}

func isSensitive(f *entgen.Field) bool {
	if f == nil {
		return false
	}
	return f.Sensitive()
}

func isVirtual(f *entgen.Field) bool {
	if f == nil {
		return false
	}
	a := getFieldAnnotation(f)
	return a != nil && a.Virtual
}

func isProtoID(e *entgen.Edge, in bool) bool {
	var s types.EdgeStrategy
	if in {
		s = getEdgeInStrategy(e)
	} else {
		s = getEdgeOutStrategy(e)
	}
	return (s & types.EdgeProtoMask) == types.EdgeProtoID
}

func isProtoMessage(e *entgen.Edge, in bool) bool {
	var s types.EdgeStrategy
	if in {
		s = getEdgeInStrategy(e)
	} else {
		s = getEdgeOutStrategy(e)
	}
	return (s & types.EdgeProtoMask) == types.EdgeProtoMessage
}

func isProtoExclude(e *entgen.Edge, in bool) bool {
	var s types.EdgeStrategy
	if in {
		s = getEdgeInStrategy(e)
	} else {
		s = getEdgeOutStrategy(e)
	}
	return (s & types.EdgeProtoMask) == types.EdgeProtoExcluded
}

func isFieldProtoExclude(f *entgen.Field, in bool) bool {
	var s types.FieldStrategy
	if in {
		s = getFieldInStrategy(f)
	} else {
		s = getFieldOutStrategy(f)
	}
	return (s & types.FieldProtoMask) == types.FieldProtoExcluded
}

func isFieldProtoOptional(f *entgen.Field, in bool) bool {
	var s types.FieldStrategy
	if in {
		s = getFieldInStrategy(f)
	} else {
		s = getFieldOutStrategy(f)
	}
	return (s & types.FieldProtoMask) == types.FieldProtoOptional
}

func isFieldProtoRequired(f *entgen.Field, in bool) bool {
	var s types.FieldStrategy
	if in {
		s = getFieldInStrategy(f)
	} else {
		s = getFieldOutStrategy(f)
	}
	return (s & types.FieldProtoMask) == types.FieldProtoRequired
}

func isSlice(f *entgen.Field) bool {
	return strings.HasPrefix(f.Type.String(), "[]") && f.Type.String() != "[]byte"
}

func getSliceElementType(f *entgen.Field) string {
	t := f.Type.String()
	if strings.HasPrefix(t, "[]") {
		return strings.TrimPrefix(t, "[]")
	}
	return t
}

func getGoProtoType(f *entgen.Field) string {
	pt := getProtoType(f)
	switch pt {
	case "double":
		return "float64"
	case "float":
		return "float32"
	case "int32", "uint32", "int64", "uint64", "bool", "string":
		return pt
	case "bytes":
		return "[]byte"
	default:
		return "string"
	}
}

func isSliceTypeMatch(f *entgen.Field) bool {
	if !isSlice(f) {
		return false
	}
	bizType := getSliceElementType(f)
	protoType := getGoProtoType(f)

	return bizType == protoType
}

type EnumDef struct {
	NodeName string
	Field    *entgen.Field
}

func getAllEnums(nodes []interface{}) []EnumDef {
	var enums []EnumDef
	for _, node := range nodes {
		n, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		name := n["Name"].(string)
		fields, ok := n["Fields"].([]*entgen.Field)
		if !ok {
			continue
		}
		for _, f := range fields {
			if f.IsEnum() {
				if isExternalEnum(f) {
					continue
				}
				enums = append(enums, EnumDef{NodeName: name, Field: f})
			}
		}
	}
	return enums
}

func bizFieldName(f *entgen.Field) string {
	if f == nil {
		return ""
	}
	a := getFieldAnnotation(f)
	if a != nil && a.BizName != "" {
		return a.BizName
	}
	return f.StructField()
}

func bizFieldType(f *entgen.Field) string {
	if f == nil {
		return ""
	}
	a := getFieldAnnotation(f)
	if a != nil && a.BizType != "" {
		return a.BizType
	}
	if f.IsEnum() {
		if isExternalEnum(f) {
			// External enum: use the external type name
			// Use the package name alias + type name
			// But here we might just return the type string if imports are handled.
			// Ideally, we want "auth.UserRole"
			return f.Type.String()
		}
		// Internal enum: defined in Biz layer
		return f.StructField()
	}
	if f.Type.String() == "uuid.UUID" {
		return "string"
	}
	return f.Type.String()
}

func explicitBizType(f *entgen.Field) string {
	if f == nil {
		return ""
	}
	a := getFieldAnnotation(f)
	if a != nil && a.BizType != "" {
		return a.BizType
	}
	return ""
}

func bizEdgeName(e *entgen.Edge, in bool) string {
	a := getAnnotation(e)
	if a != nil && a.BizName != "" {
		return a.BizName
	}
	if isBizIDOnly(e, in) {
		return e.StructField() + "ID"
	}
	return e.StructField()
}

func pascal(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, "")
}

func camel(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = []rune(strings.ToLower(string(r[0])))[0]
	return string(r)
}

func protoStructField(e *entgen.Edge, in bool) string {
	a := getAnnotation(e)
	if a != nil && a.ProtoName != "" {
		return pascal(a.ProtoName)
	}
	if isProtoID(e, in) {
		return pascal(e.Name) + "ID"
	}
	return pascal(e.Name)
}

func isExternalEnum(f *entgen.Field) bool {
	if !f.IsEnum() {
		return false
	}
	return f.Type.PkgPath != ""
}

func getExternalEnumPkg(f *entgen.Field) string {
	if f == nil {
		return ""
	}
	return f.Type.PkgPath
}

func getExternalEnumName(f *entgen.Field) string {
	if f == nil {
		return ""
	}
	return f.Type.String()
}

func collectExternalImports(nodes []interface{}) []string {
	m := make(map[string]bool)
	for _, node := range nodes {
		n, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		fields, ok := n["Fields"].([]*entgen.Field)
		if !ok {
			continue
		}
		for _, f := range fields {
			if isExternalEnum(f) {
				m[getExternalEnumPkg(f)] = true
			}
		}
	}
	var res []string
	for k := range m {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func getEnumLiteralValues(f *entgen.Field) []string {
	var res []string
	for _, e := range f.Enums {
		res = append(res, e.Value)
	}
	return res
}

func toInt(v interface{}) uint32 {
	if v == nil {
		return 0
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint32(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint32(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return uint32(rv.Float())
	default:
		return 0
	}
}
