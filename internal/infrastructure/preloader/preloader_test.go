package preloader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GroVlAn/auth-access/internal/domain"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "roles.json")

	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	return path
}

func TestPreloader_Preload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name      string
		config    string
		setupMock func(*mockrepo)
		check     func(t *testing.T, err error)
	}{
		{
			name:   "failed to read config",
			config: "",
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(
					t,
					err,
					"loading default roles from config",
				)
				require.ErrorContains(
					t,
					err,
					"reading file",
				)
			},
		},
		{
			name: "invalid json",
			config: `[
				{
					"name": "admin"
			`,
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(
					t,
					err,
					"loading default roles from config",
				)
				require.ErrorContains(
					t,
					err,
					"unmarshaling roles config",
				)
			},
		},
		{
			name:   "empty roles",
			config: `[]`,
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "create role error",
			config: `[
				{
					"name": "admin",
					"description": "Administrator",
					"is_default": true,
					"permissions": []
				}
			]`,
			setupMock: func(repo *mockrepo) {
				repo.EXPECT().
					CreateRole(
						mock.Anything,
						mock.MatchedBy(func(role domain.Role) bool {
							return role.Name == "admin" &&
								role.Description == "Administrator" &&
								role.IsDefault &&
								role.ID != "" &&
								!role.CreatedAt.IsZero() &&
								!role.UpdatedAt.IsZero()
						}),
					).
					Return("", errors.New("create role error")).
					Once()
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(
					t,
					err,
					"creating default role",
				)
				require.ErrorContains(
					t,
					err,
					"create role error",
				)
			},
		},
		{
			name: "create permission error",
			config: `[
				{
					"name": "admin",
					"description": "Administrator",
					"is_default": true,
					"permissions": [
						{
							"name": "read",
							"description": "Read access"
						}
					]
				}
			]`,
			setupMock: func(repo *mockrepo) {
				repo.EXPECT().
					CreateRole(
						mock.Anything,
						mock.MatchedBy(func(role domain.Role) bool {
							return role.Name == "admin" &&
								role.Description == "Administrator" &&
								role.IsDefault &&
								role.ID != ""
						}),
					).
					Return("role-admin", nil).
					Once()

				repo.EXPECT().
					CreatePermission(
						mock.Anything,
						"role-admin",
						mock.MatchedBy(func(permission domain.Permission) bool {
							return permission.Name == "read" &&
								permission.Description == "Read access" &&
								permission.ID != "" &&
								!permission.CreatedAt.IsZero() &&
								!permission.UpdatedAt.IsZero()
						}),
					).
					Return(errors.New("create permission error")).
					Once()
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.ErrorContains(
					t,
					err,
					"creating default permission",
				)
				require.ErrorContains(
					t,
					err,
					"create permission error",
				)
			},
		},
		{
			name: "multiple default roles",
			config: `[
				{
					"name": "admin",
					"description": "Administrator",
					"is_default": true,
					"permissions": []
				},
				{
					"name": "user",
					"description": "User",
					"is_default": true,
					"permissions": []
				}
			]`,
			setupMock: func(repo *mockrepo) {
				repo.EXPECT().
					CreateRole(
						mock.Anything,
						mock.MatchedBy(func(role domain.Role) bool {
							return role.Name == "admin" &&
								role.IsDefault
						}),
					).
					Return("role-admin", nil).
					Once()
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.EqualError(
					t,
					err,
					"set only one default role",
				)
			},
		},
		{
			name: "role without permissions",
			config: `[
				{
					"name": "user",
					"description": "Regular user",
					"is_default": false,
					"permissions": []
				}
			]`,
			setupMock: func(repo *mockrepo) {
				repo.EXPECT().
					CreateRole(
						mock.Anything,
						mock.MatchedBy(func(role domain.Role) bool {
							return role.Name == "user" &&
								role.Description == "Regular user" &&
								!role.IsDefault &&
								role.ID != ""
						}),
					).
					Return("role-user", nil).
					Once()
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success",
			config: `[
				{
					"name": "admin",
					"description": "Administrator",
					"is_default": true,
					"permissions": [
						{
							"name": "read",
							"description": "Read access"
						},
						{
							"name": "write",
							"description": "Write access"
						}
					]
				},
				{
					"name": "user",
					"description": "Regular user",
					"is_default": false,
					"permissions": [
						{
							"name": "read",
							"description": "Read access"
						}
					]
				}
			]`,
			setupMock: func(repo *mockrepo) {
				repo.EXPECT().
					CreateRole(
						mock.Anything,
						mock.MatchedBy(func(role domain.Role) bool {
							return role.Name == "admin" &&
								role.Description == "Administrator" &&
								role.IsDefault &&
								role.ID != ""
						}),
					).
					Return("role-admin", nil).
					Once()

				repo.EXPECT().
					CreatePermission(
						mock.Anything,
						"role-admin",
						mock.MatchedBy(func(permission domain.Permission) bool {
							return permission.Name == "read" &&
								permission.Description == "Read access" &&
								permission.ID != ""
						}),
					).
					Return(nil).
					Once()

				repo.EXPECT().
					CreatePermission(
						mock.Anything,
						"role-admin",
						mock.MatchedBy(func(permission domain.Permission) bool {
							return permission.Name == "write" &&
								permission.Description == "Write access" &&
								permission.ID != ""
						}),
					).
					Return(nil).
					Once()

				repo.EXPECT().
					CreateRole(
						mock.Anything,
						mock.MatchedBy(func(role domain.Role) bool {
							return role.Name == "user" &&
								role.Description == "Regular user" &&
								!role.IsDefault &&
								role.ID != ""
						}),
					).
					Return("role-user", nil).
					Once()

				repo.EXPECT().
					CreatePermission(
						mock.Anything,
						"role-user",
						mock.MatchedBy(func(permission domain.Permission) bool {
							return permission.Name == "read" &&
								permission.Description == "Read access" &&
								permission.ID != ""
						}),
					).
					Return(nil).
					Once()
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockrepo(t)

			configPath := tt.config

			if tt.config != "" {
				configPath = createConfigFile(t, tt.config)
			}

			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			p := New(
				repo,
				Deps{
					DefRolesConfigPath: configPath,
				},
			)

			err := p.Preload(ctx)

			tt.check(t, err)
		})
	}
}
