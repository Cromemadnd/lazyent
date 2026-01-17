package gen

import (
	entgen "entgo.io/ent/entc/gen"
	"github.com/Cromemadnd/lazyent/internal/types"
)

// GenNode wraps entgen.Type with additional metadata.
// Fields and Edges shadow the embedded entgen.Type fields to correctly return wrapped types in templates.
type GenNode struct {
	*entgen.Type
	Fields []*GenField
	Edges  []*GenEdge
	Enums  []*GenField // Subset of Fields that are enums
}

// GenField wraps entgen.Field with additional metadata.
type GenField struct {
	*entgen.Field
	NodeName string

	// Annotations
	Annotation *types.Annotation

	// Pre-calculated Strategies
	StrategyIn  types.FieldStrategy
	StrategyOut types.FieldStrategy

	// Pre-calculated Names
	BizName   string
	ProtoName string

	// Pre-calculated Types
	ProtoType string
	BizType   string

	// Enum Helpers
	IsExternalEnum bool
	EnumValues     map[string]int32
}

// GenEdge wraps entgen.Edge with additional metadata.
type GenEdge struct {
	*entgen.Edge

	// Annotations
	Annotation *types.Annotation

	// Pre-calculated Strategies
	StrategyIn  types.EdgeStrategy
	StrategyOut types.EdgeStrategy

	// Pre-calculated Names
	BizName   string
	ProtoName string
}
