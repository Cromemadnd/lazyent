package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/Cromemadnd/lazyent"
	"github.com/Cromemadnd/lazyent/internal/tests/testenv/pkg/auth"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.Int("age").Positive().Annotations(
			lazyent.WithProtoFieldID(2),
		),
		field.String("nickname").Optional(), // Optional String
		field.Int("score").Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithBizType("uint8"),
			lazyent.WithBizName("UserScore"),
			lazyent.WithProtoType("uint32"),
			lazyent.WithProtoName("user_score"),
		)), // Nillable Int
		field.Bool("is_verified").Default(false),                  // Bool
		field.JSON("tags", []string{}).Optional().Comment("用户标签"), // JSON
		field.String("password").Sensitive().Optional(),           // Sensitive
		field.UUID("test_uuid", uuid.UUID{}).Default(uuid.New).Comment("测试UUID"),
		field.UUID("test_nillable_uuid", uuid.UUID{}).Default(uuid.New).Nillable().Comment("测试UUID2"),
		field.Enum("status").
			Values("UNSPECIFIED", "ACTIVE", "INACTIVE", "BANNED"), // Status Enum
		field.Enum("role").
			GoType(auth.UserRole("")).
			Default(string(auth.RoleUser)).
			Comment("用户权限组"), // Custom Type Enum
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("posts", Post.Type).
			Annotations(lazyent.MergeAnnotations(
				lazyent.WithBizName("PostIDs"),
				lazyent.WithProtoName("post_ids"),
				lazyent.WithEdgeFieldStrategy(lazyent.BizIDWithProtoID), // Test BizIDOnly strategy
			)),
		edge.From("groups", Group.Type).
			Ref("users"),
		edge.To("friends", User.Type). // Test Self-Reference
						Annotations(
				lazyent.WithEdgeFieldStrategy(lazyent.BizPointerWithProtoExclude), // Test ProtoExclude
			),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}
