package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/Cromemadnd/lazyent/internal/tests/testenv/pkg/auth"
	lazyent "github.com/Cromemadnd/lazyent/internal/types"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.Int("age").Positive().Annotations(lazyent.Annotation{
			ProtoFieldID:    2,
			ProtoValidation: "gte:0",
		}),
		field.String("nickname").Optional().Annotations(lazyent.Annotation{
			ProtoValidation: "min_len:2,max_len:20,ignore_empty:true",
		}), // Optional String
		field.Int("score").Optional().Annotations(lazyent.Annotation{
			BizType:   "uint8",
			BizName:   "UserScore",
			ProtoType: "uint32",
			ProtoName: "user_score",
		}), // Nillable Int
		field.Bool("is_verified").Default(false),                  // Bool
		field.JSON("tags", []string{}).Optional().Comment("用户标签"), // JSON
		field.String("password").Sensitive().Optional(),           // Sensitive
		field.Enum("status").
			Values("ACTIVE", "INACTIVE", "BANNED").
			Annotations(lazyent.Annotation{
				EnumValues: map[string]int32{
					"ACTIVE":   1,
					"INACTIVE": 4,
					"BANNED":   3,
				},
			}),
		field.Enum("role").
			GoType(auth.UserRole("")).
			Default(string(auth.RoleUser)).
			Comment("用户权限组").
			Annotations(lazyent.Annotation{
				EnumValues: map[string]int32{
					"public":  1,
					"user":    2,
					"tech":    3,
					"dev":     4,
					"leader":  5,
					"manager": 6,
				},
			}),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("posts", Post.Type).
			Annotations(lazyent.Annotation{
				BizName:           "PostIDs",
				ProtoName:         "post_ids",
				EdgeFieldStrategy: lazyent.BizIDWithProtoID, // Test BizIDOnly strategy
			}),
		edge.From("groups", Group.Type).
			Ref("users"),
		edge.To("friends", User.Type). // Test Self-Reference
						Annotations(lazyent.Annotation{
				EdgeFieldStrategy: lazyent.BizPointerWithProtoExclude, // Test ProtoExclude
			}),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}
