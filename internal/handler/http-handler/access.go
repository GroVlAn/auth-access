package http_handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/GroVlAn/auth-access/internal/domain"
	"github.com/GroVlAn/auth-base/ew"
	"github.com/go-chi/chi/v5"
)

const (
	roleEndpoint       = "/"
	permissionEndpoint = "/permission"

	userIDKey   = "user-id"
	roleNameKey = "role-name"
)

func (h *HTTPHandler) accessRoute(r chi.Router) {
	r.Post(roleEndpoint, h.createRole)
	r.Get(roleEndpoint, h.role)
	r.Post(permissionEndpoint, h.createPermission)
	r.Get(permissionEndpoint, h.permissions)
}

func (h *HTTPHandler) createRole(w http.ResponseWriter, r *http.Request) {
	h.withBodyClose(r.Body, func(body io.ReadCloser) {
		var role domain.Role

		if err := json.NewDecoder(body).Decode(&role); err != nil {
			h.handleDecodeBody(w, err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
		defer cancel()

		if err := h.s.CreateRole(ctx, role); err != nil {
			h.handleError(w, err)
			return
		}

		h.sendSuccessResponse(w, "role created", http.StatusCreated)
	})
}

func (h *HTTPHandler) role(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userID := query.Get(userIDKey)

	ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
	defer cancel()

	role, err := h.s.Role(ctx, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.sendResponseWithData(w, role, http.StatusOK)
}

func (h *HTTPHandler) createPermission(w http.ResponseWriter, r *http.Request) {
	h.withBodyClose(r.Body, func(body io.ReadCloser) {
		var permReq domain.PermissionReq

		if err := json.NewDecoder(body).Decode(&permReq); err != nil {
			h.handleDecodeBody(w, err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
		defer cancel()

		err := h.s.CreatePermission(ctx, permReq.Permission, permReq.RoleName)
		if err != nil {
			h.handleError(w, err)
			return
		}

		h.sendSuccessResponse(w, "permission created", http.StatusCreated)
	})
}

func (h *HTTPHandler) permissions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userID := query.Get(userIDKey)
	roleName := query.Get(roleNameKey)

	if len(userID) == 0 || len(roleName) == 0 {
		h.handleError(
			w,
			ew.NewErrValidation("query must have role name or user id"),
		)
		return
	} else if len(userID) > 0 || len(roleName) > 0 {
		h.handleError(
			w,
			ew.NewErrValidation("query must have only role name or user id"),
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.DefaultTimeout)
	defer cancel()

	if len(userID) > 0 {
		h.permissionsByUserID(w, ctx, userID)
		return
	}

	h.permissionsByRoleName(w, ctx, roleName)
}

func (h *HTTPHandler) permissionsByUserID(w http.ResponseWriter, ctx context.Context, userID string) {
	permissions, err := h.s.PermissionsByUser(ctx, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.sendResponseWithData(w, permissions, http.StatusOK)
}

func (h *HTTPHandler) permissionsByRoleName(w http.ResponseWriter, ctx context.Context, roleName string) {
	permissions, err := h.s.PermissionsByRole(ctx, roleName)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.sendResponseWithData(w, permissions, http.StatusOK)
}
