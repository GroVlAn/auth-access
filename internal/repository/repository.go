package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GroVlAn/auth-access/internal/domain"
	"github.com/GroVlAn/auth-base/ew"
	"github.com/jmoiron/sqlx"
)

const (
	roleTable           = "role"
	permissionTable     = "permission"
	rolePermissionTable = "role_permissions"
	roleUserTable       = "user_roles"

	uniqueViolation = "23505"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateRole(ctx context.Context, role domain.Role) (string, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s 
			(id, name, description,	is_default, created_at, updated_at) 
			VALUES (:id, :name, :description, :is_default, :created_at, :updated_at)
			ON CONFLICT(name)
			DO UPDATE SET
				description = EXCLUDED.description,
    			updated_at = NOW()
			RETURNING id
		`,
		roleTable,
	)

	rows, err := r.db.NamedQueryContext(ctx, query, role)
	if err != nil {
		return "", handleDBError(
			fmt.Errorf("inserting new role: %w", err),
			DBErrorMessages{},
		)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", ew.New(
			ew.ErrorTypeInternal,
			errors.New("role id was not returned"),
		)
	}

	var id string

	if err := rows.Scan(&id); err != nil {
		return "", ew.New(
			ew.ErrorTypeInternal,
			fmt.Errorf("scanning role id: %w", err),
		)
	}

	return id, nil
}

func (r *Repository) Roles(ctx context.Context, userID string) ([]domain.Role, error) {
	query := fmt.Sprintf(
		`SELECT 
				r.id, 
				r.name, 
				r.description, 
				r.is_default,
				r.created_at,
				r.updated_at
				FROM %s r
				JOIN %s u 
					ON r.id = u.role_id
				WHERE u.user_id=$1`,
		roleTable,
		roleUserTable,
	)

	return r.selectRoles(ctx, query, userID)
}

func (r *Repository) RoleIDByName(ctx context.Context, name string) (string, error) {
	query := fmt.Sprintf(`
		SELECT id FROM %s WHERE name = $1
	`, roleTable)

	var id string

	if err := r.db.GetContext(ctx, &id, query, name); err != nil {
		return "", handleQueryError(err, "role not found")
	}

	return id, nil
}

func (r *Repository) DefaultRole(ctx context.Context) (domain.Role, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, is_default, created_at, updated_at
		FROM %s
		WHERE is_default = TRUE
	`,
		roleTable,
	)

	var role domain.Role

	if err := r.db.GetContext(ctx, &role, query); err != nil {
		return role, handleQueryError(err, "role not found")
	}

	return role, nil
}

func (r *Repository) CreatePermission(
	ctx context.Context,
	roleID string,
	permission domain.Permission,
) error {
	return withTx(ctx, r.db, func(tx *sqlx.Tx) error {
		permissionID, err := r.upsertPermission(tx, permission)
		if err != nil {
			return err
		}

		query := fmt.Sprintf(`
			INSERT INTO %s (
				role_id,
				permission_id
			)
			VALUES ($1, $2)
			ON CONFLICT (role_id, permission_id)
			DO NOTHING
		`, rolePermissionTable)

		_, err = tx.ExecContext(
			ctx,
			query,
			roleID,
			permissionID,
		)
		if err != nil {
			return handleDBError(
				fmt.Errorf("inserting new permission: %w", err),
				DBErrorMessages{
					NotFound: "role or permission not found",
				},
			)
		}

		return nil
	})
}

func (r *Repository) PermissionsByUser(ctx context.Context, userID string) ([]domain.Permission, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT
		p.id, 
		p.name,
		p.description,
		p.created_at,
		p.updated_at
		FROM %s ur 
		JOIN %s rp 
			ON ur.role_id = rp.role_id
		JOIN %s p
			ON rp.permission_id = p.id
		WHERE ur.user_id = $1
		`,
		roleUserTable,
		rolePermissionTable,
		permissionTable,
	)

	permissions, err := r.selectPermissions(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		return nil, r.permissionNotFound("permissions for user id", userID)
	}

	return permissions, nil
}

func (r *Repository) PermissionsByRole(ctx context.Context, roleName string) ([]domain.Permission, error) {
	query := fmt.Sprintf(`
		SELECT
		p.id, 
		p.name,
		p.description,
		p.created_at,
		p.updated_at
		FROM %s p 
		JOIN %s rp 
			ON p.id = rp.permission_id
		JOIN %s r
			ON rp.role_id = r.id
		WHERE r.name = $1
		`,
		permissionTable,
		rolePermissionTable,
		roleTable,
	)

	permissions, err := r.selectPermissions(ctx, query, roleName)
	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		return nil, r.permissionNotFound("permissions for role id", roleName)
	}

	return permissions, nil
}

func (r *Repository) SetUserRole(ctx context.Context, roleID, userID string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (role_id, user_id) VALUES ($1, $2)
	`, roleUserTable)

	res, err := r.db.ExecContext(ctx, query, roleID, userID)
	if err != nil {
		return handleDBError(
			fmt.Errorf("setting user role: %w", err),
			DBErrorMessages{
				Conflict: "user already exits this role",
				NotFound: "role not found",
			},
		)
	}

	if err := checkRowsAffected(res, "user roles", userID); err != nil {
		return fmt.Errorf("checking affected: %w", err)
	}

	return nil
}

func (r *Repository) ReplaceUserRole(
	ctx context.Context,
	userID,
	oldRoleID,
	newRoleID string,
) error {
	return withTx(ctx, r.db, func(tx *sqlx.Tx) error {
		query := fmt.Sprintf(`
			UPDATE %s
			SET role_id = $1
			WHERE user_id = $2
			  AND role_id = $3
		`, roleUserTable)

		res, err := tx.ExecContext(
			ctx,
			query,
			newRoleID,
			userID,
			oldRoleID,
		)
		if err != nil {
			return handleDBError(
				fmt.Errorf("replacing user role: %w", err),
				DBErrorMessages{
					NotFound: "role not found",
				},
			)
		}

		if err := checkRowsAffected(res, "user roles", userID); err != nil {
			return fmt.Errorf("checking affected: %w", err)
		}

		return nil
	})
}

func (r *Repository) DeleteUserRole(
	ctx context.Context,
	userID,
	roleID string,
) error {
	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE user_id = $1
		  AND role_id = $2
	`, roleUserTable)

	res, err := r.db.ExecContext(
		ctx,
		query,
		userID,
		roleID,
	)
	if err != nil {
		return handleDBError(
			fmt.Errorf("deleting user role: %w", err),
			DBErrorMessages{
				NotFound: "role not found",
			},
		)
	}

	if err := checkRowsAffected(res, "user roles", userID); err != nil {
		return fmt.Errorf("checking affected: %w", err)
	}

	return nil
}

func (r *Repository) upsertPermission(
	tx *sqlx.Tx,
	permission domain.Permission,
) (string, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id,
			name,
			description,
			created_at,
			updated_at
		)
		VALUES (
			:id,
			:name,
			:description,
			:created_at,
			:updated_at
		)
		ON CONFLICT (name)
		DO UPDATE SET
			description = EXCLUDED.description,
    		updated_at = NOW()
		RETURNING id
	`, permissionTable)

	rows, err := tx.NamedQuery(
		query,
		permission,
	)
	if err != nil {
		return "", handleDBError(
			fmt.Errorf("upserting permission: %w", err),
			DBErrorMessages{
				Conflict: "permission already exist",
				NotFound: "role not found",
			},
		)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", ew.New(
			ew.ErrorTypeInternal,
			errors.New("permission id was not returned"),
		)
	}

	var id string

	if err := rows.Scan(&id); err != nil {
		return "", ew.New(
			ew.ErrorTypeInternal,
			fmt.Errorf("scanning permission id: %w", err),
		)
	}

	return id, nil
}

func (r *Repository) selectRoles(ctx context.Context, query string, args ...any) ([]domain.Role, error) {
	var roles []domain.Role

	if err := r.db.SelectContext(ctx, &roles, query, args...); err != nil {
		return nil, ew.New(
			ew.ErrorTypeInternal,
			fmt.Errorf("selecting roles: %w", err),
		)
	}

	if len(roles) == 0 {
		return nil, ew.New(
			ew.ErrorTypeNotFound,
			fmt.Errorf("roles not found"),
		).Msg("user roles not found")
	}

	return roles, nil
}

func (r *Repository) selectPermissions(ctx context.Context, query string, args ...any) ([]domain.Permission, error) {
	var permissions []domain.Permission

	if err := r.db.SelectContext(ctx, &permissions, query, args...); err != nil {
		return nil, ew.New(
			ew.ErrorTypeInternal,
			fmt.Errorf("selecting permissions: %w", err),
		)
	}

	if len(permissions) == 0 {
		return nil, ew.New(
			ew.ErrorTypeNotFound,
			fmt.Errorf("permissions not found"),
		).Msg("user permissions not found")
	}

	return permissions, nil
}

func (r *Repository) permissionNotFound(entity, id string) error {
	return ew.New(
		ew.ErrorTypeNotFound,
		fmt.Errorf("%s %s not found", entity, id),
	).Msg("permission not found")
}

func checkRowsAffected(res sql.Result, entity string, id string) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return ew.New(ew.ErrorTypeInternal, fmt.Errorf("getting affected rows: %w", err))
	}
	if rows == 0 {
		return ew.New(ew.ErrorTypeNotFound, fmt.Errorf("%s with id %s not found", entity, id))
	}
	return nil
}
