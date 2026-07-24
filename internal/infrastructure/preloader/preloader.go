package preloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/GroVlAn/auth-access/internal/domain"
	"github.com/google/uuid"
)

type repo interface {
	CreateRole(ctx context.Context, role domain.Role) error
	CreatePermission(
		ctx context.Context,
		roleID string,
		permission domain.Permission,
	) error
}

type Deps struct {
	DefRolesConfigPath string
}

type Preloader struct {
	repo repo
	Deps
}

func New(repo repo, deps Deps) *Preloader {
	return &Preloader{
		repo: repo,
		Deps: deps,
	}
}

func (p *Preloader) Preload(ctx context.Context) error {
	rolesList, err := load(p.DefRolesConfigPath)
	if err != nil {
		return fmt.Errorf("loading default roles from config: %w", err)
	}

	var isOneDefault bool

	for _, el := range rolesList {
		if isOneDefault && el.IsDefault {
			return fmt.Errorf("set only one default role")
		}

		id, err := p.createRole(ctx, el)
		if err != nil {
			return err
		}

		if err := p.createPermission(ctx, el.Permissions, id); err != nil {
			return err
		}

		if el.IsDefault {
			isOneDefault = true
		}
	}

	return nil
}

func (p *Preloader) createRole(ctx context.Context, el domain.RoleElement) (string, error) {
	role := domain.Role{
		ID:          uuid.NewString(),
		Name:        el.Name,
		Description: el.Description,
		IsDefault:   el.IsDefault,
		CreatedAt:   time.Now(),
		UpdateAt:    time.Now(),
	}

	if err := p.repo.CreateRole(ctx, role); err != nil {
		return "", fmt.Errorf("creating default role: %w", err)
	}

	return role.ID, nil
}

func (p *Preloader) createPermission(
	ctx context.Context,
	permEls []domain.PermissionElement,
	roleID string,
) error {
	for _, el := range permEls {
		permission := domain.Permission{
			ID:          uuid.NewString(),
			Name:        el.Name,
			Description: el.Description,
			CreatedAt:   time.Now(),
			UpdateAt:    time.Now(),
		}

		if err := p.repo.CreatePermission(ctx, roleID, permission); err != nil {
			return fmt.Errorf("creating default permission: %w", err)
		}
	}

	return nil
}

func load(path string) ([]domain.RoleElement, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var roles []domain.RoleElement
	if err = json.Unmarshal(file, &roles); err != nil {
		return nil, fmt.Errorf("unmarshaling roles config: %w", err)
	}

	return roles, nil
}
