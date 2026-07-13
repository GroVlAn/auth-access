package grpc_handler

import (
	"context"

	"github.com/GroVlAn/auth-access/internal/domain"
	api "github.com/GroVlAn/auth-api/access"
	"github.com/GroVlAn/auth-base/ew/grpcx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *GRPCHandler) CreateRole(ctx context.Context, reqRole *api.Role) (*api.Success, error) {
	role := domain.Role{
		ID:          reqRole.ID,
		Name:        reqRole.Name,
		Description: reqRole.Description,
		IsDefault:   reqRole.IsDefault,
	}

	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	if err := h.s.CreateRole(ctx, role); err != nil {
		return nil, grpcx.HandleError(err)
	}

	return &api.Success{
		Success: true,
	}, nil
}

func (h *GRPCHandler) Role(ctx context.Context, userID *api.UserID) (*api.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, h.defaultTimeout)
	defer cancel()

	role, err := h.s.Role(ctx, userID.User_ID)
	if err != nil {
		return nil, grpcx.HandleError(err)
	}

	return &api.Role{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsDefault:   role.IsDefault,
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdateAt),
	}, nil
}

func (h *GRPCHandler) CreatePermission(ctx context.Context, req *api.PermissionReq) (*api.Success, error) {
	permissionReq := domain.PermissionReq{
		RoleName: req.RoleName,
		Permission: domain.Permission{
			ID:          req.Permission.ID,
			Name:        req.Permission.Name,
			Description: req.Permission.Description,
			IsDefault:   req.Permission.IsDefault,
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
		return nil, grpcx.HandleError(err)
	}

	return &api.Success{
		Success: true,
	}, nil
}

func (h *GRPCHandler) GetPermissionByUserID(ctx context.Context, userID *api.UserID) (*api.Permissions, error) {
	return h.permissions(ctx, userID.User_ID, h.s.PermissionsByUser)
}

func (h *GRPCHandler) GetPermissionByRoleName(ctx context.Context, roleName *api.RoleName) (*api.Permissions, error) {
	return h.permissions(ctx, roleName.RoleName, h.s.PermissionsByRole)
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
		return nil, grpcx.HandleError(err)
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
				IsDefault:   perm.IsDefault,
				CreatedAt:   timestamppb.New(perm.CreatedAt),
				UpdatedAt:   timestamppb.New(perm.UpdateAt),
			},
		)
	}

	return res, nil
}
