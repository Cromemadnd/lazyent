package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/Cromemadnd/lazyent"
)

// Group holds the schema definition for the Group entity.
type Group struct {
	ent.Schema
}

// Fields of the Group.
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique().Annotations(lazyent.Annotation{}),
	}
}

// Edges of the Group.
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type).Annotations(
			lazyent.Annotation{
				EdgeInStrategy:  lazyent.EdgeProtoMessage | lazyent.EdgeBizPointer,
				EdgeOutStrategy: lazyent.EdgeProtoMessage | lazyent.EdgeBizPointer,
			},
		),
		edge.To("admins", User.Type).
			Annotations(lazyent.Annotation{
				EdgeInStrategy:  lazyent.EdgeProtoExcluded | lazyent.EdgeBizExcluded,
				EdgeOutStrategy: lazyent.EdgeProtoExcluded | lazyent.EdgeBizExcluded,
			}),
	}
}

func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}
