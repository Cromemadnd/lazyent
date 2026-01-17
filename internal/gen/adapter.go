package gen

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	entgen "entgo.io/ent/entc/gen"
	"github.com/Cromemadnd/lazyent/internal/types"
)

// AdaptGraph converts entgen.Graph into a list of GenNode.
func AdaptGraph(g *entgen.Graph) ([]*GenNode, error) {
	var nodes []*GenNode
	for _, n := range g.Nodes {
		gn, err := adaptNode(n)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, gn)
	}
	// Sort by name for deterministic output
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	return nodes, nil
}

func adaptNode(n *entgen.Type) (*GenNode, error) {
	gn := &GenNode{
		Type: n,
	}

	// Fields
	for _, f := range n.Fields {
		gf, err := adaptField(f, n.Name)
		if err != nil {
			return nil, fmt.Errorf("node %s: %w", n.Name, err)
		}
		gn.Fields = append(gn.Fields, gf)
		if f.IsEnum() {
			gn.Enums = append(gn.Enums, gf)
		}
	}

	// Edges
	for _, e := range n.Edges {
		ge, err := adaptEdge(e)
		if err != nil {
			return nil, fmt.Errorf("node %s: %w", n.Name, err)
		}
		gn.Edges = append(gn.Edges, ge)
	}

	return gn, nil
}

func adaptField(f *entgen.Field, nodeName string) (*GenField, error) {
	gf := &GenField{
		Field:    f,
		NodeName: nodeName,
	}

	// 1. Annotation
	gf.Annotation = parseFieldAnnotation(f)

	// 2. Strategies
	gf.StrategyIn = resolveFieldInStrategy(f, gf.Annotation)
	gf.StrategyOut = resolveFieldOutStrategy(f, gf.Annotation)

	// 3. Names
	gf.BizName = f.StructField()
	if gf.Annotation != nil && gf.Annotation.BizName != "" {
		gf.BizName = gf.Annotation.BizName
	}

	gf.ProtoName = camelToSnake(f.Name)
	if gf.Annotation != nil && gf.Annotation.ProtoName != "" {
		gf.ProtoName = gf.Annotation.ProtoName
	}

	// 4. Types
	gf.IsExternalEnum = isExternalEnumInternal(f)
	gf.BizType = resolveBizType(f, gf.Annotation)
	gf.ProtoType = ResolveProtoTypeString(f, gf.Annotation, nodeName)

	// 5. Enum Values
	if f.IsEnum() {
		gf.EnumValues = resolveEnumValues(f, gf.Annotation)
	}

	return gf, nil
}

func adaptEdge(e *entgen.Edge) (*GenEdge, error) {
	ge := &GenEdge{
		Edge: e,
	}

	// 1. Annotation
	ge.Annotation = parseEdgeAnnotation(e)

	// 2. Strategies
	ge.StrategyIn = resolveEdgeInStrategy(e, ge.Annotation)
	ge.StrategyOut = resolveEdgeOutStrategy(e, ge.Annotation)

	// 3. Names
	ge.BizName = e.StructField() // Default
	if ge.Annotation != nil && ge.Annotation.BizName != "" {
		ge.BizName = ge.Annotation.BizName
	} else if isBizIDOnlyStrategy(ge.StrategyIn) {
		ge.BizName = e.StructField() + "ID"
	}

	ge.ProtoName = camelToSnake(e.Name)
	if ge.Annotation != nil && ge.Annotation.ProtoName != "" {
		ge.ProtoName = ge.Annotation.ProtoName
	} else if isProtoIDStrategy(ge.StrategyIn) {
		ge.ProtoName = camelToSnake(e.Name) + "_id"
	}

	return ge, nil
}

// --- Resolvers ---

func resolveFieldInStrategy(f *entgen.Field, a *types.Annotation) types.FieldStrategy {
	if a != nil && a.FieldInStrategy != 0 {
		return a.FieldInStrategy
	}
	if f.Sensitive() {
		return types.FieldBizValue | types.FieldProtoRequired
	}
	if f.Optional {
		return types.FieldBizPointer | types.FieldProtoOptional
	}
	return types.FieldBizValue | types.FieldProtoRequired
}

func resolveFieldOutStrategy(f *entgen.Field, a *types.Annotation) types.FieldStrategy {
	if a != nil && a.FieldOutStrategy != 0 {
		return a.FieldOutStrategy
	}
	if f.Sensitive() {
		return types.FieldBizExcluded | types.FieldProtoExcluded
	}
	if f.Optional {
		return types.FieldBizPointer | types.FieldProtoOptional
	}
	return types.FieldBizValue | types.FieldProtoRequired
}

func resolveEdgeInStrategy(e *entgen.Edge, a *types.Annotation) types.EdgeStrategy {
	if a != nil && a.EdgeInStrategy != 0 {
		return a.EdgeInStrategy
	}
	return types.EdgeProtoMessage | types.EdgeBizPointer
}

func resolveEdgeOutStrategy(e *entgen.Edge, a *types.Annotation) types.EdgeStrategy {
	if a != nil && a.EdgeOutStrategy != 0 {
		return a.EdgeOutStrategy
	}
	return types.EdgeProtoMessage | types.EdgeBizPointer
}

func resolveBizType(f *entgen.Field, a *types.Annotation) string {
	if a != nil && a.BizType != "" {
		return a.BizType
	}
	if f.IsEnum() {
		if isExternalEnumInternal(f) {
			return f.Type.String()
		}
		return f.StructField()
	}
	if f.Type.String() == "uuid.UUID" {
		return "string"
	}
	return f.Type.String()
}

func ResolveProtoTypeString(f *entgen.Field, a *types.Annotation, nodeName string) string {
	if a != nil && a.ProtoType != "" {
		return a.ProtoType
	}
	if f.IsEnum() && !isExternalEnumInternal(f) {
		return nodeName + f.StructField()
	}
	t := f.Type.String()
	switch t {
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
	case "[]byte":
		return "bytes"
	default:
		if strings.HasPrefix(t, "[]") {
			return "string"
		}
		return "string"
	}
}

func resolveEnumValues(f *entgen.Field, a *types.Annotation) map[string]int32 {
	if a != nil && a.EnumValues != nil {
		return a.EnumValues
	}
	m := make(map[string]int32)
	for i, e := range f.Enums {
		m[e.Value] = int32(i)
	}
	return m
}

// --- Utilities ---

func parseFieldAnnotation(f *entgen.Field) *types.Annotation {
	if f == nil || f.Annotations == nil {
		return nil
	}
	return extractAnnotation(f.Annotations)
}

func parseEdgeAnnotation(e *entgen.Edge) *types.Annotation {
	if e == nil || e.Annotations == nil {
		return nil
	}
	return extractAnnotation(e.Annotations)
}

func extractAnnotation(ants map[string]interface{}) *types.Annotation {
	if v, ok := ants["LazyEnt"]; ok {
		if a, ok := v.(types.Annotation); ok {
			return &a
		}
		if m, ok := v.(map[string]interface{}); ok {
			return decodeAnnotationMap(m)
		}
	}
	if v, ok := ants["lazyent"]; ok {
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

	return a
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

func isBizIDOnlyStrategy(s types.EdgeStrategy) bool {
	return (s & types.EdgeBizMask) == types.EdgeBizID
}

func isProtoIDStrategy(s types.EdgeStrategy) bool {
	return (s & types.EdgeProtoMask) == types.EdgeProtoID
}

// Duplicated from helpers for now
func isExternalEnumInternal(f *entgen.Field) bool {
	if !f.IsEnum() {
		return false
	}
	return f.Type.PkgPath != ""
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
