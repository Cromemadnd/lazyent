package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/Cromemadnd/lazyent"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty(),
		field.Text("content"),
		field.String("slug").Unique().Optional(),

		// 场景：数据库字段，但不暴露给 Biz 和 Proto (如搜索索引或加密盐)
		field.String("internal_code").Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithFieldInStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
			lazyent.WithFieldOutStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
		)),

		// 场景：输入层可选，输出层排除（类似于密码，但手动配置）
		field.String("management_key").Sensitive().Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithFieldInStrategy(lazyent.FieldProtoOptional|lazyent.FieldBizValue),
			lazyent.WithFieldOutStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
		)),

		// 场景：入参排除（由服务端逻辑生成），回包包含
		field.String("summary").Optional().Annotations(lazyent.MergeAnnotations(
			lazyent.WithFieldInStrategy(lazyent.FieldProtoExcluded|lazyent.FieldBizExcluded),
			lazyent.WithFieldOutStrategy(lazyent.FieldProtoOptional|lazyent.FieldBizValue),
		)),

		// 场景：Virtual 字段测试
		field.String("extra_data").Annotations(lazyent.Virtual()).Optional(),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		// 1. 常规：双端均为 Message 与 Pointer (默认行为)
		edge.From("author", User.Type).
			Ref("posts").
			Required().
			Unique().
			Annotations(lazyent.MergeAnnotations(
				lazyent.WithEdgeInStrategy(lazyent.EdgeProtoMessage|lazyent.EdgeBizPointer),
				lazyent.WithEdgeOutStrategy(lazyent.EdgeProtoMessage|lazyent.EdgeBizPointer),
			)),

		// 2. 场景：创建时传 ID 列表 (EdgeProtoID)，回包返回完整对象 (EdgeProtoMessage)
		edge.To("co_authors", User.Type).
			Annotations(lazyent.MergeAnnotations(
				lazyent.WithEdgeInStrategy(lazyent.EdgeProtoID|lazyent.EdgeBizPointer),
				lazyent.WithEdgeOutStrategy(lazyent.EdgeProtoMessage|lazyent.EdgeBizPointer),
			)),

		// 3. 场景：双端均为 ID 模式 (精简模式)
		edge.To("relevant_groups", Group.Type).
			Annotations(lazyent.MergeAnnotations(
				lazyent.WithEdgeInStrategy(lazyent.EdgeProtoID|lazyent.EdgeBizID),
				lazyent.WithEdgeOutStrategy(lazyent.EdgeProtoID|lazyent.EdgeBizID),
			)),

		// 4. 场景：入参隐藏，回包仅返回 ID 列表 (只读关联)
		edge.To("followers", User.Type).
			Annotations(lazyent.MergeAnnotations(
				lazyent.WithEdgeInStrategy(lazyent.EdgeProtoExcluded|lazyent.EdgeBizExcluded),
				lazyent.WithEdgeOutStrategy(lazyent.EdgeProtoID|lazyent.EdgeBizPointer),
			)),

		// 5. 场景：入参传完整 Message，回包仅返回 ID 列表 (存档/引用风格)
		edge.To("co_authors_archive", User.Type).
			Annotations(lazyent.MergeAnnotations(
				lazyent.WithEdgeInStrategy(lazyent.EdgeProtoMessage|lazyent.EdgeBizPointer),
				lazyent.WithEdgeOutStrategy(lazyent.EdgeProtoID|lazyent.EdgeBizPointer),
			)),
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
