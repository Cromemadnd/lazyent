package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	lazyent "github.com/Cromemadnd/lazyent/internal/types"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().Annotations(lazyent.Annotation{}),
		field.Text("content"),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("posts").
			Required().
			Unique().
			Annotations(lazyent.Annotation{
				EdgeFieldStrategy: lazyent.BizPointerWithProtoID,
			}),
	}
}

func (Post) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Post) Indexes() []ent.Index {
	return []ent.Index{}
}
