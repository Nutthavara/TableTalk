package handler

import (
	"net/http"
	"time"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/model"
	"tabletalk/backend/internal/service"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

type reviewJSON struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	UserID       string    `json:"user_id"`
	AuthorName   string    `json:"author_name"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	IsMine       bool      `json:"is_mine"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func reviewToJSON(r *http.Request, m model.Review) reviewJSON {
	viewer, signedIn := auth.FromContext(r.Context())
	return reviewJSON{
		ID:           m.ID.String(),
		RestaurantID: m.RestaurantID.String(),
		UserID:       m.UserID.String(),
		AuthorName:   m.AuthorName,
		Rating:       m.Rating,
		Comment:      m.Comment,
		IsMine:       signedIn && viewer.UserID == m.UserID,
		CreatedAt:    m.CreatedAt.UTC(),
		UpdatedAt:    m.UpdatedAt.UTC(),
	}
}

type reviewBody struct {
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
}

func (b reviewBody) input() (service.ReviewInput, error) {
	if b.Rating == nil {
		return service.ReviewInput{}, apperr.Validation(map[string]string{"rating": "This field is required."})
	}
	in := service.ReviewInput{Rating: *b.Rating}
	if b.Comment != nil {
		in.Comment = *b.Comment
	}
	return in, nil
}

// List handles GET /api/restaurants/{id}/reviews?page=&page_size= (newest first).
func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := pathID(r, "id", service.ErrRestaurantNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	q := r.URL.Query()
	details := map[string]string{}
	page := queryInt(q.Get("page"), "page", 1, details)
	pageSize := queryInt(q.Get("page_size"), "page_size", service.DefaultPageSize, details)
	if len(details) > 0 {
		apperr.Write(w, r, apperr.Validation(details))
		return
	}

	result, err := h.svc.List(r.Context(), restaurantID, page, pageSize)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	items := make([]reviewJSON, len(result.Items))
	for i, m := range result.Items {
		items[i] = reviewToJSON(r, m)
	}
	writeJSON(w, http.StatusOK, newPage(items, result.Page, result.PageSize, result.Total))
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	restaurantID, err := pathID(r, "id", service.ErrRestaurantNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	var body reviewBody
	if err := decodeJSON(w, r, &body); err != nil {
		apperr.Write(w, r, err)
		return
	}
	in, err := body.input()
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	m, err := h.svc.Create(r.Context(), user, restaurantID, in)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, reviewToJSON(r, m))
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	id, err := pathID(r, "id", service.ErrReviewNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	var body reviewBody
	if err := decodeJSON(w, r, &body); err != nil {
		apperr.Write(w, r, err)
		return
	}
	in, err := body.input()
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	m, err := h.svc.Update(r.Context(), user, id, in)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reviewToJSON(r, m))
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	id, err := pathID(r, "id", service.ErrReviewNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	if err := h.svc.Delete(r.Context(), user, id); err != nil {
		apperr.Write(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
