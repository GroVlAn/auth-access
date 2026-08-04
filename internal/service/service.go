package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GroVlAn/auth-access/internal/domain"
	"github.com/google/uuid"
)

type repo interface {
	CreateRole(ctx context.Context, role domain.Role) (string, error)
	Roles(ctx context.Context, userID string) ([]domain.Role, error)
	RoleIDByName(ctx context.Context, name string) (string, error)
	DefaultRole(ctx context.Context) (domain.Role, error)
	CreatePermission(
		ctx context.Context,
		roleID string,
		permission domain.Permission,
	) error
	FullPermissions(ctx context.Context, userID string) ([]string, error)
	PermissionsByUser(ctx context.Context, userID string) ([]domain.Permission, error)
	PermissionsByRole(ctx context.Context, roleName string) ([]domain.Permission, error)
	SetUserRole(ctx context.Context, roleID, userID string) error
	ReplaceUserRole(
		ctx context.Context,
		userID,
		oldRoleID,
		newRoleID string,
	) error
	DeleteUserRole(
		ctx context.Context,
		userID,
		roleID string,
	) error
}

type Service struct {
	repo repo
}

func New(repo repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateRole(ctx context.Context, role domain.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	role.ID = uuid.NewString()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	if _, err := s.repo.CreateRole(ctx, role); err != nil {
		return fmt.Errorf("creating new role: %w", err)
	}

	return nil
}

func (s *Service) Roles(ctx context.Context, userID string) ([]domain.Role, error) {
	roles, err := s.repo.Roles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting role by user id: %w", err)
	}

	return roles, nil
}

func (s *Service) CreatePermission(ctx context.Context, permission domain.Permission, roleName string) error {
	roleID, err := s.repo.RoleIDByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("getting role id by role name: %w", err)
	}

	permission.ID = uuid.NewString()

	if err := s.repo.CreatePermission(ctx, roleID, permission); err != nil {
		return fmt.Errorf("creating permission: %w", err)
	}

	return nil
}

func (s *Service) FullPermissions(ctx context.Context, userID string) ([]string, error) {
	permissions, err := s.repo.FullPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting full user permissions: %w", err)
	}

	return permissions, nil
}

func (s *Service) PermissionsByUser(ctx context.Context, userID string) ([]domain.Permission, error) {
	permissions, err := s.repo.PermissionsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting permissions by user: %w", err)
	}

	return permissions, nil
}

func (s *Service) PermissionsByRole(ctx context.Context, roleName string) ([]domain.Permission, error) {
	permissions, err := s.repo.PermissionsByRole(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("getting permissions by user: %w", err)
	}

	return permissions, nil
}

func (s *Service) AddUserRole(ctx context.Context, userID, roleName string) error {
	roleID, err := s.repo.RoleIDByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("getting role id by name: %w", err)
	}

	if err := s.repo.SetUserRole(ctx, roleID, userID); err != nil {
		return fmt.Errorf("setting new user role: %w", err)
	}

	return nil
}

func (s *Service) BindUserRole(ctx context.Context, userID string) error {
	defRole, err := s.repo.DefaultRole(ctx)
	if err != nil {
		return fmt.Errorf("getting default role: %w", err)
	}

	if err := s.repo.SetUserRole(ctx, defRole.ID, userID); err != nil {
		return fmt.Errorf("setting default user role: %w", err)
	}

	return nil
}

func (s *Service) ReplaceUserRole(ctx context.Context, updUsRl domain.UpdateUserRoleReq) error {
	oldRoleID, err := s.repo.RoleIDByName(ctx, updUsRl.OldRoleName)
	if err != nil {
		return fmt.Errorf("getting role id by name: %w", err)
	}

	newRoleID, err := s.repo.RoleIDByName(ctx, updUsRl.NewRoleName)
	if err != nil {
		return fmt.Errorf("getting role id by name: %w", err)
	}

	if err := s.repo.ReplaceUserRole(ctx, updUsRl.UserID, oldRoleID, newRoleID); err != nil {
		return fmt.Errorf("replacing user role: %w", err)
	}

	return nil
}

func (s *Service) DeleteUserRole(ctx context.Context, userID, roleName string) error {
	roleID, err := s.repo.RoleIDByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("getting role id by name: %w", err)
	}

	if err := s.repo.DeleteUserRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("deleting user role: %w", err)
	}

	return nil
}
