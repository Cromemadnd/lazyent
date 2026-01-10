package auth

import "slices"

type UserRole string // 定义新类型

const (
	RolePublic  UserRole = "public"
	RoleUser    UserRole = "user"
	RoleTech    UserRole = "tech"
	RoleDev     UserRole = "dev"
	RoleLeader  UserRole = "leader"
	RoleManager UserRole = "manager"
)

func (r UserRole) Values() []string {
	return []string{
		string(RolePublic),
		string(RoleUser),
		string(RoleTech),
		string(RoleDev),
		string(RoleLeader),
		string(RoleManager),
	}
}

func (r UserRole) IsValid() bool { // 验证用户组是否合法
	return slices.Contains(r.Values(), string(r))
}
