package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/Cromemadnd/lazyent"
	"github.com/google/uuid"
)

type VirtualUser struct {
	ent.Schema
}

func (VirtualUser) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Annotations(
			lazyent.Annotation{
				BizName:   "UUID",
				ProtoName: "uuid",
			}).Comment("数据库主键"),
		field.String("id_auth").Comment("统一身份认证号"),
		field.String("id_student").Comment("学工号"),

		field.String("info_name").Comment("姓名"),
		field.Int32("info_campus").Comment("校区"),
		field.String("info_qq").Comment("QQ"),
		field.String("info_wechat").Comment("微信"),
		field.String("info_email").Comment("邮箱"),
		field.String("info_phone").Comment("电话"),
		field.Text("info_custom").Comment("默认特殊信息"),
		field.Int32("info_gender").Comment("性别"),

		field.Bool("lanker_valid").Comment("是否为蓝客"),
		field.Bool("lanker_trainee").Comment("是否为见习"),
		field.Int32("lanker_sheets_completed").Comment("完成单数"),
		field.Float("lanker_score").Comment("积分"),
		field.String("lanker_avatar_url").Comment("头像"),
		field.Text("lanker_brief").Comment("个人简介"),
		field.Int32("lanker_department").Comment("所在部门"),
		field.Int32("lanker_spot").Comment("职位"),
	}
}

func (VirtualUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// pkgMixin.BaseMixin{},
	}
}

func (VirtualUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Skip(),
	}
}
