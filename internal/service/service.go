package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GroVlAn/auth-access/internal/domain"
	"github.com/google/uuid"
)

type repo interface {
	CreateRole(ctx context.Context, role domain.Role) error
	Role(ctx context.Context, userID string) (domain.Role, error)
	RoleIDByName(ctx context.Context, name string) (string, error)
	DefaultRole(ctx context.Context) (domain.Role, error)
	CreatePermission(
		ctx context.Context,
		permission domain.Permission,
		roleID, rpID string) error
	PermissionsByUser(ctx context.Context, userID string) ([]domain.Permission, error)
	PermissionsByRole(ctx context.Context, roleName string) ([]domain.Permission, error)
	SetUserRole(ctx context.Context, roleID, userID string) error
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
	role.UpdateAt = time.Now()

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return fmt.Errorf("creating new role: %w", err)
	}

	return nil
}

func (s *Service) Role(ctx context.Context, userID string) (domain.Role, error) {
	role, err := s.repo.Role(ctx, userID)
	if err != nil {
		return domain.Role{}, fmt.Errorf("getting role by user id: %w", err)
	}

	return role, nil
}

func (s *Service) CreatePermission(ctx context.Context, permission domain.Permission, roleName string) error {
	roleID, err := s.repo.RoleIDByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("getting role id by role name: %w", err)
	}

	rpID := uuid.NewString()

	if err := s.repo.CreatePermission(ctx, permission, roleID, rpID); err != nil {
		return fmt.Errorf("creating permission: %w", err)
	}

	return nil
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
