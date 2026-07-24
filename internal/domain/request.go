package domain

type PermissionReq struct {
	RoleName   string     `json:"role_name"`
	Permission Permission `json:"permission"`
}

type UpdateUserRoleReq struct {
	UserID      string `json:"user_id"`
	OldRoleName string `json:"old_role_name"`
	NewRoleName string `json:"new_role_name"`
}
