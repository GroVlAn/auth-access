package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/GroVlAn/auth-access/internal/domain"
	"github.com/GroVlAn/auth-base/ew"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateRole(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	validRole := domain.Role{
		Name:        "admin",
		Description: "administrator",
	}

	tests := []struct {
		name      string
		role      domain.Role
		setupMock func(m *mockrepo)
		check     func(t *testing.T, err error, m *mockrepo)
	}{
		{
			name: "validation error",
			role: domain.Role{},
			check: func(t *testing.T, err error, m *mockrepo) {
				require.ErrorContains(t, err, ew.NewErrValidation("validation role").Error())

				m.AssertNotCalled(t, "CreateRole")
			},
		},
		{
			name: "create role error",
			role: validRole,
			setupMock: func(m *mockrepo) {
				m.On(
					"CreateRole",
					mock.Anything,
					mock.MatchedBy(func(role domain.Role) bool {
						return role.Name == validRole.Name &&
							role.Description == validRole.Description &&
							role.ID != "" &&
							!role.CreatedAt.IsZero() &&
							!role.UpdatedAt.IsZero()
					}),
				).Return("", fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error, m *mockrepo) {
				require.ErrorContains(t, err, "creating new role")
			},
		},
		{
			name: "success",
			role: validRole,
			setupMock: func(m *mockrepo) {
				m.On(
					"CreateRole",
					mock.Anything,
					mock.MatchedBy(func(role domain.Role) bool {
						return role.Name == validRole.Name &&
							role.Description == validRole.Description &&
							role.ID != "" &&
							!role.CreatedAt.IsZero() &&
							!role.UpdatedAt.IsZero()
					}),
				).Return("role-id", nil).Once()
			},
			check: func(t *testing.T, err error, m *mockrepo) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			s := New(repo)

			err := s.CreateRole(ctx, tt.role)

			tt.check(t, err, repo)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Roles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := []domain.Role{
		{
			ID:   "1",
			Name: "admin",
		},
		{
			ID:   "2",
			Name: "user",
		},
	}

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, []domain.Role, error)
	}{
		{
			name: "repo error",
			setupMock: func(m *mockrepo) {
				m.On(
					"Roles",
					mock.Anything,
					"user-id",
				).Return(nil, fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, roles []domain.Role, err error) {
				require.Nil(t, roles)
				require.ErrorContains(t, err, "getting role by user id")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On(
					"Roles",
					mock.Anything,
					"user-id",
				).Return(expected, nil).Once()
			},
			check: func(t *testing.T, roles []domain.Role, err error) {
				require.NoError(t, err)
				require.Equal(t, expected, roles)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			roles, err := s.Roles(ctx, "user-id")

			tt.check(t, roles, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_CreatePermission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	permission := domain.Permission{
		Name:        "watch",
		Description: "Watch content",
	}

	tests := []struct {
		name       string
		permission domain.Permission
		setupMock  func(*mockrepo)
		check      func(t *testing.T, err error, m *mockrepo)
	}{
		{
			name:       "validation error",
			permission: domain.Permission{},
			check: func(t *testing.T, err error, m *mockrepo) {
				require.ErrorContains(t, err, ew.NewErrValidation("validation permission").Error())

				m.AssertNotCalled(t, "CreatePermission")
			},
		},
		{
			name:       "role id error",
			permission: permission,
			setupMock: func(m *mockrepo) {
				m.On(
					"RoleIDByName",
					mock.Anything,
					"admin",
				).Return("", fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error, m *mockrepo) {
				require.ErrorContains(t, err, "getting role id by role name")
			},
		},
		{
			name:       "create permission error",
			permission: permission,
			setupMock: func(m *mockrepo) {
				m.On(
					"RoleIDByName",
					mock.Anything,
					"admin",
				).Return("role-id", nil).Once()

				m.On(
					"CreatePermission",
					mock.Anything,
					"role-id",
					mock.MatchedBy(func(p domain.Permission) bool {
						return p.Name == permission.Name &&
							p.Description == permission.Description &&
							p.ID != ""
					}),
				).Return(fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error, m *mockrepo) {
				require.ErrorContains(t, err, "creating permission")
			},
		},
		{
			name:       "success",
			permission: permission,
			setupMock: func(m *mockrepo) {
				m.On(
					"RoleIDByName",
					mock.Anything,
					"admin",
				).Return("role-id", nil).Once()

				m.On(
					"CreatePermission",
					mock.Anything,
					"role-id",
					mock.MatchedBy(func(p domain.Permission) bool {
						return p.Name == permission.Name &&
							p.Description == permission.Description &&
							p.ID != ""
					}),
				).Return(nil).Once()
			},
			check: func(t *testing.T, err error, m *mockrepo) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			s := New(repo)

			err := s.CreatePermission(ctx, tt.permission, "admin")

			tt.check(t, err, repo)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_FullPermissions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := []string{
		"users.create",
		"users.read",
		"users.update",
	}

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, []string, error)
	}{
		{
			name: "repo error",
			setupMock: func(m *mockrepo) {
				m.On(
					"FullPermissions",
					mock.Anything,
					"user-id",
				).Return(nil, fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, permissions []string, err error) {
				require.Nil(t, permissions)
				require.ErrorContains(t, err, "getting full user permissions")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On(
					"FullPermissions",
					mock.Anything,
					"user-id",
				).Return(expected, nil).Once()
			},
			check: func(t *testing.T, permissions []string, err error) {
				require.NoError(t, err)
				require.Equal(t, expected, permissions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			permissions, err := s.FullPermissions(ctx, "user-id")

			tt.check(t, permissions, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_PermissionsByUser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := []domain.Permission{
		{
			ID:          "1",
			Name:        "users.read",
			Description: "read users",
		},
		{
			ID:          "2",
			Name:        "users.write",
			Description: "write users",
		},
	}

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, []domain.Permission, error)
	}{
		{
			name: "repo error",
			setupMock: func(m *mockrepo) {
				m.On(
					"PermissionsByUser",
					mock.Anything,
					"user-id",
				).Return(nil, fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, permissions []domain.Permission, err error) {
				require.Nil(t, permissions)
				require.ErrorContains(t, err, "getting permissions by user")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On(
					"PermissionsByUser",
					mock.Anything,
					"user-id",
				).Return(expected, nil).Once()
			},
			check: func(t *testing.T, permissions []domain.Permission, err error) {
				require.NoError(t, err)
				require.Equal(t, expected, permissions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			permissions, err := s.PermissionsByUser(ctx, "user-id")

			tt.check(t, permissions, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_PermissionsByRole(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := []domain.Permission{
		{
			ID:          "1",
			Name:        "users.read",
			Description: "read users",
		},
		{
			ID:          "2",
			Name:        "users.write",
			Description: "write users",
		},
	}

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, []domain.Permission, error)
	}{
		{
			name: "repo error",
			setupMock: func(m *mockrepo) {
				m.On(
					"PermissionsByRole",
					mock.Anything,
					"admin",
				).Return(nil, fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, permissions []domain.Permission, err error) {
				require.Nil(t, permissions)
				require.ErrorContains(t, err, "getting permissions by role name")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On(
					"PermissionsByRole",
					mock.Anything,
					"admin",
				).Return(expected, nil).Once()
			},
			check: func(t *testing.T, permissions []domain.Permission, err error) {
				require.NoError(t, err)
				require.Equal(t, expected, permissions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			permissions, err := s.PermissionsByRole(ctx, "admin")

			tt.check(t, permissions, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_AddUserRole(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	const (
		userID   = "user-id"
		roleName = "admin"
		roleID   = "role-id"
	)

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, error)
	}{
		{
			name: "role id error",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, roleName).
					Return("", fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "getting role id by name")
			},
		},
		{
			name: "set user role error",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, roleName).
					Return(roleID, nil).Once()

				m.On("SetUserRole", mock.Anything, roleID, userID).
					Return(fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "setting new user role")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, roleName).
					Return(roleID, nil).Once()

				m.On("SetUserRole", mock.Anything, roleID, userID).
					Return(nil).Once()
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			err := s.AddUserRole(ctx, userID, roleName)

			tt.check(t, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_BindUserRole(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	const userID = "user-id"

	defaultRole := domain.Role{
		ID:   "default-role-id",
		Name: "user",
	}

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, error)
	}{
		{
			name: "default role error",
			setupMock: func(m *mockrepo) {
				m.On("DefaultRole", mock.Anything).
					Return(domain.Role{}, fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "getting default role")
			},
		},
		{
			name: "set default role error",
			setupMock: func(m *mockrepo) {
				m.On("DefaultRole", mock.Anything).
					Return(defaultRole, nil).Once()

				m.On("SetUserRole", mock.Anything, defaultRole.ID, userID).
					Return(fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "setting default user role")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On("DefaultRole", mock.Anything).
					Return(defaultRole, nil).Once()

				m.On("SetUserRole", mock.Anything, defaultRole.ID, userID).
					Return(nil).Once()
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			err := s.BindUserRole(ctx, userID)

			tt.check(t, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ReplaceUserRole(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := domain.UpdateUserRoleReq{
		UserID:      "user-id",
		OldRoleName: "user",
		NewRoleName: "admin",
	}

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, error)
	}{
		{
			name: "old role error",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, req.OldRoleName).
					Return("", fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "getting role id by name")
			},
		},
		{
			name: "new role error",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, req.OldRoleName).
					Return("old-role-id", nil).Once()

				m.On("RoleIDByName", mock.Anything, req.NewRoleName).
					Return("", fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "getting role id by name")
			},
		},
		{
			name: "replace error",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, req.OldRoleName).
					Return("old-role-id", nil).Once()

				m.On("RoleIDByName", mock.Anything, req.NewRoleName).
					Return("new-role-id", nil).Once()

				m.On("ReplaceUserRole",
					mock.Anything,
					req.UserID,
					"old-role-id",
					"new-role-id",
				).Return(fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "replacing user role")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, req.OldRoleName).
					Return("old-role-id", nil).Once()

				m.On("RoleIDByName", mock.Anything, req.NewRoleName).
					Return("new-role-id", nil).Once()

				m.On("ReplaceUserRole",
					mock.Anything,
					req.UserID,
					"old-role-id",
					"new-role-id",
				).Return(nil).Once()
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			err := s.ReplaceUserRole(ctx, req)

			tt.check(t, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteUserRole(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	const (
		userID   = "user-id"
		roleName = "admin"
		roleID   = "role-id"
	)

	tests := []struct {
		name      string
		setupMock func(*mockrepo)
		check     func(*testing.T, error)
	}{
		{
			name: "role id error",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, roleName).
					Return("", fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "getting role id by name")
			},
		},
		{
			name: "delete role error",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, roleName).
					Return(roleID, nil).Once()

				m.On("DeleteUserRole",
					mock.Anything,
					userID,
					roleID,
				).Return(fmt.Errorf("db error")).Once()
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "deleting user role")
			},
		},
		{
			name: "success",
			setupMock: func(m *mockrepo) {
				m.On("RoleIDByName", mock.Anything, roleName).
					Return(roleID, nil).Once()

				m.On("DeleteUserRole",
					mock.Anything,
					userID,
					roleID,
				).Return(nil).Once()
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockrepo)

			tt.setupMock(repo)

			s := New(repo)

			err := s.DeleteUserRole(ctx, userID, roleName)

			tt.check(t, err)

			repo.AssertExpectations(t)
		})
	}
}
