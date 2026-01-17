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
		field.String("remote_token").
			Annotations(lazyent.Virtual()).
			Comment("虚拟令牌，不存数据库"),
		field.Any("ext_user").
			Annotations(lazyent.MergeAnnotations(
				lazyent.Virtual(),
				lazyent.WithBizType("*auth.User"),
				lazyent.WithProtoType("auth.v1.User"),
			)).
			Comment("外部用户对象，不存数据库"),

		field.Time("test_time").Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithBizType("time.Time"),
			lazyent.WithProtoType("int64"),
		)),

		// 补全字段策略测试覆盖
		field.String("last_login_ip").Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithFieldInStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
			lazyent.WithFieldOutStrategy(lazyent.FieldProtoOptional|lazyent.FieldBizValue),
		)).Comment("仅回包包含"),
		field.String("verification_code").Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithFieldInStrategy(lazyent.FieldProtoRequired|lazyent.FieldBizValue),
			lazyent.WithFieldOutStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
		)).Comment("仅入参包含"),
		field.Int64("internal_id").Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithFieldInStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
			lazyent.WithFieldOutStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
		)).Comment("内部 ID，两端排除"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("posts", Post.Type).
			Annotations(lazyent.MergeAnnotations(
				lazyent.WithBizName("PostIDs"),
				lazyent.WithProtoName("post_ids"),
				lazyent.WithEdgeInStrategy(lazyent.EdgeProtoID|lazyent.EdgeBizID),
				lazyent.WithEdgeOutStrategy(lazyent.EdgeProtoID|lazyent.EdgeBizID),
			)),
		edge.From("groups", Group.Type).
			Ref("users"),
		// 4. 场景：入参隐藏，回包仅返回 ID 列表 (只读关联)
		edge.To("followers", User.Type).
			Annotations(lazyent.Annotation{
				EdgeInStrategy:  lazyent.EdgeProtoExcluded | lazyent.EdgeBizExcluded,
				EdgeOutStrategy: lazyent.EdgeProtoID | lazyent.EdgeBizPointer,
			}),

		// 5. 场景：入参传完整 Message，回包仅返回 ID 列表 (存档/引用风格)
		edge.To("co_authors_archive", User.Type).
			Annotations(lazyent.Annotation{
				ProtoName:       "co_authors_archive_test",
				EdgeInStrategy:  lazyent.EdgeProtoMessage | lazyent.EdgeBizPointer,
				EdgeOutStrategy: lazyent.EdgeProtoID | lazyent.EdgeBizPointer,
			}),
		edge.To("friends", User.Type). // Test Self-Reference
						Annotations(
				lazyent.WithEdgeInStrategy(lazyent.EdgeProtoExcluded|lazyent.EdgeBizPointer),
				lazyent.WithEdgeOutStrategy(lazyent.EdgeProtoExcluded|lazyent.EdgeBizPointer),
			),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}
