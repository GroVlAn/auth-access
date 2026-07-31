package http_handler

import (
	"context"
	"net/http"
	"time"

	"github.com/GroVlAn/auth-access/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type service interface {
	CreateRole(ctx context.Context, role domain.Role) error
	Roles(ctx context.Context, userID string) ([]domain.Role, error)
	CreatePermission(ctx context.Context, permission domain.Permission, roleName string) error
	PermissionsByUser(ctx context.Context, userID string) ([]domain.Permission, error)
	PermissionsByRole(ctx context.Context, roleName string) ([]domain.Permission, error)
	AddUserRole(ctx context.Context, userID, roleName string) error
	ReplaceUserRole(ctx context.Context, updUsRl domain.UpdateUserRoleReq) error
	DeleteUserRole(ctx context.Context, userID, roleName string) error
}

type MiddlewareConf struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type Deps struct {
	BasePath       string
	DefaultTimeout time.Duration
}

type HTTPHandler struct {
	l     zerolog.Logger
	s     service
	mConf MiddlewareConf
	Deps
}

func New(
	l zerolog.Logger,
	s service,
	deps Deps,
	mConf MiddlewareConf,
) *HTTPHandler {
	return &HTTPHandler{
		l:     l,
		s:     s,
		Deps:  deps,
		mConf: mConf,
	}
}

func (h *HTTPHandler) Handler() *chi.Mux {
	r := chi.NewRouter()

	h.cors(r)

	r.Route("/", func(r chi.Router) {
		r.Get("/home", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome to the Home Page!"))
		})
	})

	r.Route(h.BasePath, func(r chi.Router) {
		h.accessRoute(r)
	})

	return r
}
