package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/Cromemadnd/lazyent/internal/tests/testenv/pkg/auth"
	lazyent "github.com/Cromemadnd/lazyent/internal/types"
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
