package lazyent

import (
	"github.com/Cromemadnd/lazyent/internal/types"
)

// Exported types & Consts
type Annotation = types.Annotation
type EdgeStrategy = types.EdgeStrategy
type FieldStrategy = types.FieldStrategy

const (
	// --- Edge Strategies ---
	EdgeProtoMessage  = types.EdgeProtoMessage
	EdgeProtoID       = types.EdgeProtoID
	EdgeProtoExcluded = types.EdgeProtoExcluded

	EdgeBizPointer  = types.EdgeBizPointer
	EdgeBizID       = types.EdgeBizID
	EdgeBizExcluded = types.EdgeBizExcluded

	// --- Field Strategies ---
	FieldProtoRequired = types.FieldProtoRequired
	FieldProtoOptional = types.FieldProtoOptional
	FieldProtoExcluded = types.FieldProtoExcluded

	FieldBizValue    = types.FieldBizValue
	FieldBizPointer  = types.FieldBizPointer
	FieldBizExcluded = types.FieldBizExcluded
)
