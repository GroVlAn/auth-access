package grpc_handler

import (
	"context"

	"github.com/GroVlAn/auth-access/internal/domain"
	api "github.com/GroVlAn/auth-api/access"
	"github.com/GroVlAn/auth-base/ew/grpcx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *GRPCHandler) CreateRole(ctx context.Context, req *api.Role) (*api.Success, error) {
	role := domain.Role{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   req.IsDefault,
	}

	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.CreateRole(ctx, role); err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	return success(), nil
}

func (h *GRPCHandler) Role(ctx context.Context, req *api.UserID) (*api.Roles, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	roles, err := h.s.Roles(ctx, req.User_ID)
	if err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	respRoles := make([]*api.Role, 0, len(roles))

	for _, role := range roles {
		respRole := &api.Role{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsDefault:   role.IsDefault,
			CreatedAt:   timestamppb.New(role.CreatedAt),
			UpdatedAt:   timestamppb.New(role.UpdatedAt),
		}

		respRoles = append(respRoles, respRole)
	}

	return &api.Roles{
		Roles: respRoles,
	}, nil
}

func (h *GRPCHandler) CreatePermission(ctx context.Context, req *api.PermissionReq) (*api.Success, error) {
	permissionReq := domain.PermissionReq{
		RoleName: req.RoleName,
		Permission: domain.Permission{
			ID:          req.Permission.ID,
			Name:        req.Permission.Name,
			Description: req.Permission.Description,
		},
	}

	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	err := h.s.CreatePermission(
		ctx,
		permissionReq.Permission,
		permissionReq.RoleName,
	)
	if err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	return success(), nil
}

func (h *GRPCHandler) GetPermissionByUserID(ctx context.Context, req *api.UserID) (*api.Permissions, error) {
	return h.permissions(ctx, req.User_ID, h.s.PermissionsByUser)
}

func (h *GRPCHandler) GetPermissionByRoleName(ctx context.Context, req *api.RoleName) (*api.Permissions, error) {
	return h.permissions(ctx, req.RoleName, h.s.PermissionsByRole)
}

func (h *GRPCHandler) BindUserRole(ctx context.Context, req *api.UserID) (*api.Success, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.BindUserRole(ctx, req.User_ID); err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	return success(), nil
}

func (h *GRPCHandler) AddUserRole(ctx context.Context, req *api.UserIDRoleName) (*api.Success, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.AddUserRole(ctx, req.User_ID, req.RoleName); err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	return success(), nil
}

func (h *GRPCHandler) ReplaceUserRole(ctx context.Context, req *api.UpdateUserRoleReq) (*api.Success, error) {
	updUsRl := domain.UpdateUserRoleReq{
		UserID:      req.User_ID,
		OldRoleName: req.OldRoleName,
		NewRoleName: req.NewRoleName,
	}

	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.ReplaceUserRole(ctx, updUsRl); err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	return success(), nil
}

func (h *GRPCHandler) DeleteUserRole(ctx context.Context, req *api.UserIDRoleName) (*api.Success, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.DeleteUserRole(ctx, req.User_ID, req.RoleName); err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	return success(), nil
}

func (h *GRPCHandler) permissions(
	ctx context.Context,
	field string,
	fn func(ctx context.Context, field string) ([]domain.Permission, error),
) (*api.Permissions, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	permissions, err := fn(ctx, field)
	if err != nil {
		return nil, grpcx.HandleError(h.l, err)
	}

	res := &api.Permissions{
		Permissions: make([]*api.Permission, 0, len(permissions)),
	}

	for _, perm := range permissions {
		res.Permissions = append(
			res.Permissions,
			&api.Permission{
				ID:          perm.ID,
				Name:        perm.Name,
				Description: perm.Description,
				CreatedAt:   timestamppb.New(perm.CreatedAt),
				UpdatedAt:   timestamppb.New(perm.UpdatedAt),
			},
		)
	}

	return res, nil
}

func success() *api.Success {
	return &api.Success{
		Success: true,
	}
}
