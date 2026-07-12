package domain

type PermissionReq struct {
	RoleName   string     `json:"role_name"`
	Permission Permission `json:"permission"`
}
