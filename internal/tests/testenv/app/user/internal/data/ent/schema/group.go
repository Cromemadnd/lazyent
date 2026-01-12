package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	lazyent "github.com/Cromemadnd/lazyent/internal/types"
)

// Group holds the schema definition for the Group entity.
type Group struct {
	ent.Schema
}

// Fields of the Group.
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique().Annotations(lazyent.Annotation{
			ProtoValidation: "min_len:0",
		}),
	}
}

// Edges of the Group.
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type).Annotations(
			lazyent.Annotation{
				EdgeFieldStrategy: lazyent.BizPointerWithProtoMessage, // Default explicit
			},
		),
		edge.To("admins", User.Type).
			Annotations(lazyent.Annotation{
				EdgeFieldStrategy: lazyent.BizExcludeWithProtoExclude, // Completely hidden
			}),
	}
}

func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}
