package gen

import (
	"reflect"
	"sort"
	"strings"

	entgen "entgo.io/ent/entc/gen"
	"github.com/Cromemadnd/lazyent/internal/types"
)

// --- Adapters for Interface Support ---

func asGenField(v interface{}) *GenField {
	if v == nil {
		return nil
	}
	if gf, ok := v.(*GenField); ok {
		return gf
	}
	if f, ok := v.(*entgen.Field); ok {
		// Just-in-time adaptation
		gf, _ := adaptField(f, "")
		return gf
	}
	return nil
}

func asGenEdge(v interface{}) *GenEdge {
	if v == nil {
		return nil
	}
	if ge, ok := v.(*GenEdge); ok {
		return ge
	}
	if e, ok := v.(*entgen.Edge); ok {
		// Just-in-time adaptation
		ge, _ := adaptEdge(e)
		return ge
	}
	return nil
}

// --- Helper Functions ---

func hasTimeNodes(nodes []interface{}) bool {
	for _, node := range nodes {
		// Try GenNode
		if gn, ok := node.(*GenNode); ok {
			for _, f := range gn.Fields {
				if f.Type.String() == "time.Time" {
					return true
				}
			}
			continue
		}
	}
	return false
}

func hasUUIDNodes(nodes []interface{}) bool {
	for _, node := range nodes {
		if gn, ok := node.(*GenNode); ok {
			if hasUUID(gn.Fields, gn.Edges) {
				return true
			}
			continue
		}
	}
	return false
}

// Need to handle []interface{} (which might be []*GenField or equivalent)
// But template usually passes .Fields which is specific slice type.
// If hasTime is called with .Fields from template, it's either []*GenField or []*entgen.Field.
// We use reflection to handle slice iteration.
func hasTime(fields interface{}) bool {
	v := reflect.ValueOf(fields)
	if v.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < v.Len(); i++ {
		f := asGenField(v.Index(i).Interface())
		if f != nil && f.Type.String() == "time.Time" {
			return true
		}
	}
	return false
}

func hasUUID(fields interface{}, edges interface{}) bool {
	vFields := reflect.ValueOf(fields)
	if vFields.Kind() == reflect.Slice {
		for i := 0; i < vFields.Len(); i++ {
			f := asGenField(vFields.Index(i).Interface())
			if f != nil && f.Type.String() == "uuid.UUID" {
				return true
			}
		}
	}

	vEdges := reflect.ValueOf(edges)
	if vEdges.Kind() == reflect.Slice {
		for i := 0; i < vEdges.Len(); i++ {
			e := asGenEdge(vEdges.Index(i).Interface())
			if e != nil && e.Type.ID.Type.String() == "uuid.UUID" {
				return true
			}
		}
	}
	return false
}

func edgeHasFK(v interface{}) bool {
	e := asGenEdge(v)
	if e == nil {
		return false
	}
	if !e.IsInverse() {
		return e.Unique
	}
	if e.Ref != nil && !e.Ref.Unique {
		return e.Unique
	}
	return false
}

func edgeField(v interface{}, in bool) string { return bizEdgeName(v, in) }

func hasField(fields interface{}, name string) bool {
	v := reflect.ValueOf(fields)
	if v.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < v.Len(); i++ {
		f := asGenField(v.Index(i).Interface())
		if f != nil && f.StructField() == name {
			return true
		}
	}
	return false
}

func edgeIDType(v interface{}) string {
	e := asGenEdge(v)
	if e == nil {
		return "string"
	}
	t := e.Type.ID.Type.String()
	if t == "uuid.UUID" {
		return "string"
	}
	return t
}

func entIDType(v interface{}) string {
	e := asGenEdge(v)
	if e == nil {
		return "string"
	}
	return e.Type.ID.Type.String()
}

func edgeProtoType(v interface{}, in bool) string {
	e := asGenEdge(v)
	if e == nil {
		return "string"
	}
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

func protoType(v interface{}, nodeName string) string {
	f := asGenField(v)
	if f == nil {
		return "string"
	}
	// Use pre-calculated logic
	// If adapter didn't have fallback for nodeName, we rely on JIT here which might fail nodeName logic for enums if passed empty.
	// But GenField has NodeName.
	if f.NodeName == "" && nodeName != "" {
		f.NodeName = nodeName
	}
	return ResolveProtoTypeString(f.Field, f.Annotation, f.NodeName)
}

func getProtoType(v interface{}) string {
	return protoType(v, "")
}

func getProtoTag(v interface{}, i int) int {
	f := asGenField(v)
	if f == nil || f.Annotation == nil || f.Annotation.ProtoFieldID == 0 {
		return i + 1
	}
	return int(f.Annotation.ProtoFieldID)
}

func protoFieldName(v interface{}) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	return f.ProtoName
}

func protoEdgeFieldName(v interface{}, in bool) string {
	e := asGenEdge(v)
	if e == nil {
		return ""
	}
	if e.Annotation != nil && e.Annotation.ProtoName != "" {
		return e.Annotation.ProtoName
	}
	if isProtoID(v, in) {
		return camelToSnake(e.Name) + "_id"
	}
	return camelToSnake(e.Name)
}

func protoStructField(v interface{}, in bool) string {
	e := asGenEdge(v)
	if e == nil {
		return ""
	}
	if e.Annotation != nil && e.Annotation.ProtoName != "" {
		return pascal(e.Annotation.ProtoName)
	}
	if isProtoID(v, in) {
		return pascal(e.Name) + "ID"
	}
	return pascal(e.Name)
}

func protoGoName(v interface{}) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	if f.Annotation != nil && f.Annotation.ProtoName != "" {
		return pascal(f.Annotation.ProtoName)
	}
	return pascal(f.Name)
}

func getEnumValues(v interface{}) map[string]int32 {
	f := asGenField(v)
	if f == nil {
		return nil
	}
	return f.EnumValues
}

type EnumPair struct {
	Key   string
	Value int32
}

func getEnumPairs(v interface{}) []EnumPair {
	f := asGenField(v)
	var pairs []EnumPair
	if f == nil {
		return pairs
	}
	vals := f.EnumValues
	// To preserve generic order if no annotation, JIT
	if vals == nil {
		// Should not happen if adaptField worked
		return pairs
	}
	// Sort by value to be deterministic or check original enums
	// If original Enums slice exists, use that order unless overridden values
	if len(f.Enums) > 0 {
		for i, e := range f.Enums {
			if val, ok := vals[e.Value]; ok {
				pairs = append(pairs, EnumPair{Key: e.Value, Value: val})
			} else {
				// fallback
				pairs = append(pairs, EnumPair{Key: e.Value, Value: int32(i)})
			}
		}
	} else {
		// Just map
		for k, v := range vals {
			pairs = append(pairs, EnumPair{Key: k, Value: v})
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].Value < pairs[j].Value })
	}
	return pairs
}

// Re-export Annotation getter for internal use
func getFieldAnnotation(f *entgen.Field) *types.Annotation {
	// Delegated to adapter.go via exported or duplicated logic?
	// helpers.go originally had this. We can keep it or use the one from adapter.go (if exported).
	// Since I duplicated/moved logic to adapter.go, I should probably remove this or make it call adapter.
	// But adapter.go functions are private mostly.
	// I'll keep the logic I copied into adapter.go, and since I'm overwriting helpers.go, I can rely on asGenField().Annotation
	return asGenField(f).Annotation
}

// Strategies

func isBizIDOnly(v interface{}, in bool) bool {
	e := asGenEdge(v)
	if e == nil {
		return false
	}
	return isBizIDOnlyStrategy(ifStrategy(in, e.StrategyIn, e.StrategyOut))
}

func isBizExclude(v interface{}, in bool) bool {
	e := asGenEdge(v)
	if e == nil {
		return false
	}
	s := ifStrategy(in, e.StrategyIn, e.StrategyOut)
	return (s & types.EdgeBizMask) == types.EdgeBizExcluded
}

func shouldGenerateBizEdge(v interface{}) bool {
	return !isBizExclude(v, true) || !isBizExclude(v, false)
}

func isBizPointer(v interface{}, in bool) bool {
	e := asGenEdge(v)
	if e == nil {
		return false
	}
	s := ifStrategy(in, e.StrategyIn, e.StrategyOut)
	return (s & types.EdgeBizMask) == types.EdgeBizPointer
}

func shouldBizPointer(v interface{}) bool {
	return isBizPointer(v, true) || isBizPointer(v, false)
}

func isSensitive(v interface{}) bool {
	f := asGenField(v)
	return f != nil && f.Sensitive()
}

func isVirtual(v interface{}) bool {
	f := asGenField(v)
	return f != nil && f.Annotation != nil && f.Annotation.Virtual
}

func isProtoID(v interface{}, in bool) bool {
	e := asGenEdge(v)
	if e == nil {
		return false
	}
	return isProtoIDStrategy(ifStrategy(in, e.StrategyIn, e.StrategyOut))
}

func isProtoMessage(v interface{}, in bool) bool {
	e := asGenEdge(v)
	if e == nil {
		return false
	}
	s := ifStrategy(in, e.StrategyIn, e.StrategyOut)
	return (s & types.EdgeProtoMask) == types.EdgeProtoMessage
}

func isProtoExclude(v interface{}, in bool) bool {
	e := asGenEdge(v)
	if e == nil {
		return false
	}
	s := ifStrategy(in, e.StrategyIn, e.StrategyOut)
	return (s & types.EdgeProtoMask) == types.EdgeProtoExcluded
}

func isFieldProtoExclude(v interface{}, in bool) bool {
	f := asGenField(v)
	if f == nil {
		return false
	}
	s := ifFStrategy(in, f.StrategyIn, f.StrategyOut)
	return (s & types.FieldProtoMask) == types.FieldProtoExcluded
}

func isFieldProtoOptional(v interface{}, in bool) bool {
	f := asGenField(v)
	if f == nil {
		return false
	}
	s := ifFStrategy(in, f.StrategyIn, f.StrategyOut)
	return (s & types.FieldProtoMask) == types.FieldProtoOptional
}

func isFieldProtoRequired(v interface{}, in bool) bool {
	f := asGenField(v)
	if f == nil {
		return false
	}
	s := ifFStrategy(in, f.StrategyIn, f.StrategyOut)
	return (s & types.FieldProtoMask) == types.FieldProtoRequired
}

func shouldGenerateBizField(v interface{}) bool {
	return !isFieldProtoExclude(v, true) || !isFieldProtoExclude(v, false) // Wait, logic in original was BizExclude
	// Original: 533: 	return !isFieldBizExclude(f, true) || !isFieldBizExclude(f, false)
	// I need isFieldBizExclude
}

func isFieldBizExclude(v interface{}, in bool) bool {
	f := asGenField(v)
	if f == nil {
		return false
	}
	s := ifFStrategy(in, f.StrategyIn, f.StrategyOut)
	return (s & types.FieldBizMask) == types.FieldBizExcluded

}

// Utils

func ifStrategy(in bool, sIn, sOut types.EdgeStrategy) types.EdgeStrategy {
	if in {
		return sIn
	}
	return sOut
}

func ifFStrategy(in bool, sIn, sOut types.FieldStrategy) types.FieldStrategy {
	if in {
		return sIn
	}
	return sOut
}

func isSlice(v interface{}) bool {
	f := asGenField(v)
	return f != nil && strings.HasPrefix(f.Type.String(), "[]") && f.Type.String() != "[]byte"
}

func getSliceElementType(v interface{}) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	t := f.Type.String()
	if strings.HasPrefix(t, "[]") {
		return strings.TrimPrefix(t, "[]")
	}
	return t
}

func getGoProtoType(v interface{}) string {
	f := asGenField(v)
	if f == nil {
		return "string"
	}
	// Delegate to pre-calculated ProtoType if matches basic types
	// But getGoProtoType mapping is slightly different from ProtoType (e.g. double -> float64)
	pt := f.ProtoType
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

func isSliceTypeMatch(v interface{}) bool {
	f := asGenField(v)
	if f == nil || !isSlice(f) {
		return false
	}
	bizType := getSliceElementType(f)
	protoType := getGoProtoType(f)
	return bizType == protoType
}

func getAllEnums(nodes interface{}) []interface{} {
	// Returning struct that template can use
	type EnumDef struct {
		NodeName string
		Field    *GenField // Use GenField
	}
	var enums []interface{}

	var processNode func(*GenNode)
	processNode = func(n *GenNode) {
		for _, f := range n.Fields {
			if f.IsEnum() && !f.IsExternalEnum {
				enums = append(enums, EnumDef{NodeName: n.Name, Field: f})
			}
		}
	}

	// Helper to iterate nodes generic list
	v := reflect.ValueOf(nodes)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			node := v.Index(i).Interface()
			if gn, ok := node.(*GenNode); ok {
				processNode(gn)
			}
		}
	}
	return enums
}

func bizFieldName(v interface{}) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	return f.BizName
}

func bizFieldType(v interface{}) string {
	f := asGenField(v)
	if f == nil {
		return ""
	}
	return f.BizType
}

func explicitBizType(v interface{}) string {
	f := asGenField(v)
	if f != nil && f.Annotation != nil && f.Annotation.BizType != "" {
		return f.Annotation.BizType
	}
	return ""
}

func bizEdgeName(v interface{}, in bool) string {
	e := asGenEdge(v)
	if e == nil {
		return ""
	}
	if in {
		// Use StrategyIn logic (e.g. if ID only)
		// But AdaptEdge already calculated BizName. Wait, BizName is static name?
		// My AdaptEdge logic: if IDOnly, BizName = Field + ID.
		// Does this cover both In and Out?
		// If StrategyIn is IDOnly, but StrategyOut is Message, do we have different BizNames?
		// "BizName" usually refers to the field name in the Biz struct.
		// The Biz struct is usually one shared struct. So name should be consistent?
		// Actually input and output might be different structs (UpdateInput vs Entity).
		// For now, return e.BizName.
		return e.BizName
	}
	// For output, if defined differently... stick to e.BizName for now.
	return e.BizName
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

func collectExternalImports(nodes []interface{}) []string {
	// ... Simplified logic using GenNode
	m := make(map[string]bool)
	for _, node := range nodes {
		if gn, ok := node.(*GenNode); ok {
			for _, f := range gn.Fields {
				if f.IsExternalEnum {
					m[f.Type.PkgPath] = true
				}
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

func getEnumLiteralValues(v interface{}) []string {
	f := asGenField(v)
	var res []string
	if f == nil {
		return res
	}
	for _, e := range f.Enums {
		res = append(res, e.Value)
	}
	return res
}

func isExternalEnum(v interface{}) bool {
	f := asGenField(v)
	return f != nil && f.IsExternalEnum
}
func getExternalEnumPkg(v interface{}) string {
	f := asGenField(v)
	if f != nil {
		return f.Type.PkgPath
	}
	return ""
}
func getExternalEnumName(v interface{}) string {
	f := asGenField(v)
	if f != nil {
		return f.Type.String()
	}
	return ""
}
