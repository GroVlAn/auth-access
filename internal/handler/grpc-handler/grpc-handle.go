package grpc_handler

import (
	"context"
	"time"

	"github.com/GroVlAn/auth-access/internal/domain"
	api "github.com/GroVlAn/auth-api/access"
	"github.com/rs/zerolog"
)

type service interface {
	CreateRole(ctx context.Context, role domain.Role) error
	Roles(ctx context.Context, userID string) ([]domain.Role, error)
	CreatePermission(ctx context.Context, permission domain.Permission, roleName string) error
	PermissionsByUser(ctx context.Context, userID string) ([]domain.Permission, error)
	PermissionsByRole(ctx context.Context, roleName string) ([]domain.Permission, error)
	AddUserRole(ctx context.Context, userID, roleName string) error
	BindUserRole(ctx context.Context, userID string) error
	ReplaceUserRole(ctx context.Context, updUsRl domain.UpdateUserRoleReq) error
	DeleteUserRole(ctx context.Context, userID, roleName string) error
}

type GRPCHandler struct {
	api.UnimplementedAccessServiceServer
	l              zerolog.Logger
	s              service
	defaultTimeout time.Duration
}

func New(l zerolog.Logger, s service, defTimeout time.Duration) *GRPCHandler {
	return &GRPCHandler{
		l:              l,
		s:              s,
		defaultTimeout: defTimeout,
	}
}
