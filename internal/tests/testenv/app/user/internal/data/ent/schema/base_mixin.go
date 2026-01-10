package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/Cromemadnd/lazyent"
	"github.com/google/uuid"
)

type BaseMixin struct {
	mixin.Schema
}

func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			StorageKey("uuid").
			Default(uuid.New).
			Comment("数据库主键").
			Annotations(lazyent.Annotation{
				BizName: "UUID",
				// uuid 在 biz 默认被映射为 string
				ProtoName: "uuid",
			}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

func (BaseMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}
