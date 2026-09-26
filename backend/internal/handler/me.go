package handler

import (
	"net/http"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
)

type meResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email,omitempty"`
}

func Me(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		apperr.Write(w, r, apperr.ErrUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, meResponse{
		ID:          id.UserID.String(),
		Username:    id.Username,
		DisplayName: id.DisplayName,
		Email:       id.Email,
	})
}
