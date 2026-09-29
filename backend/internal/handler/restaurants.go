package handler

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/model"
	"tabletalk/backend/internal/service"
)

// Largest accepted multipart body: all images at their maximum size plus form fields.
const maxMultipartBody = service.MaxImagesPerRestaurant*service.MaxImageBytes + 1<<20

type RestaurantHandler struct {
	svc     *service.RestaurantService
	reviews *service.ReviewService
}

func NewRestaurantHandler(svc *service.RestaurantService, reviews *service.ReviewService) *RestaurantHandler {
	return &RestaurantHandler{svc: svc, reviews: reviews}
}

type imageJSON struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Position int    `json:"position"`
}

type restaurantJSON struct {
	ID                      string      `json:"id"`
	OwnerID                 string      `json:"owner_id"`
	OwnerName               string      `json:"owner_name"`
	Name                    string      `json:"name"`
	Description             string      `json:"description"`
	Capacity                int         `json:"capacity"`
	OpeningTime             string      `json:"opening_time"`
	ClosingTime             string      `json:"closing_time"`
	Timezone                string      `json:"timezone"`
	CancellationNoticeHours int         `json:"cancellation_notice_hours"`
	Images                  []imageJSON `json:"images"`
	AverageRating           *float64    `json:"average_rating"`
	ReviewCount             int         `json:"review_count"`
	IsOwner                 bool        `json:"is_owner"`
	// MyReviewID is the viewer's review of this restaurant (detail endpoint only).
	MyReviewID *string   `json:"my_review_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (h *RestaurantHandler) toJSON(r *http.Request, m model.Restaurant) restaurantJSON {
	images := make([]imageJSON, len(m.Images))
	for i, img := range m.Images {
		images[i] = imageJSON{ID: img.ID.String(), URL: h.svc.ImageURL(img), Position: img.Position}
	}
	viewer, signedIn := auth.FromContext(r.Context())
	return restaurantJSON{
		ID:                      m.ID.String(),
		OwnerID:                 m.OwnerID.String(),
		OwnerName:               m.OwnerName,
		Name:                    m.Name,
		Description:             m.Description,
		Capacity:                m.Capacity,
		OpeningTime:             m.OpeningTime.String(),
		ClosingTime:             m.ClosingTime.String(),
		Timezone:                m.Timezone,
		CancellationNoticeHours: m.CancellationNoticeHours,
		Images:                  images,
		AverageRating:           m.AverageRating,
		ReviewCount:             m.ReviewCount,
		IsOwner:                 signedIn && viewer.UserID == m.OwnerID,
		CreatedAt:               m.CreatedAt.UTC(),
		UpdatedAt:               m.UpdatedAt.UTC(),
	}
}

func (h *RestaurantHandler) list(r *http.Request, items []model.Restaurant) []restaurantJSON {
	out := make([]restaurantJSON, len(items))
	for i, m := range items {
		out[i] = h.toJSON(r, m)
	}
	return out
}

func (h *RestaurantHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id", service.ErrRestaurantNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	m, err := h.svc.Get(r.Context(), id)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	out := h.toJSON(r, m)
	if viewer, ok := auth.FromContext(r.Context()); ok {
		reviewID, err := h.reviews.MyReviewID(r.Context(), viewer, id)
		if err != nil {
			apperr.Write(w, r, err)
			return
		}
		if reviewID != nil {
			s := reviewID.String()
			out.MyReviewID = &s
		}
	}
	writeJSON(w, http.StatusOK, out)
}

type pageJSON[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func newPage[T any](items []T, page, pageSize, total int) pageJSON[T] {
	return pageJSON[T]{Items: items, Page: page, PageSize: pageSize, Total: total, TotalPages: (total + pageSize - 1) / pageSize}
}

// List handles GET /api/restaurants?q=&sort=rating|reviews&page=1&page_size=12.
func (h *RestaurantHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	details := map[string]string{}
	page := queryInt(q.Get("page"), "page", 1, details)
	pageSize := queryInt(q.Get("page_size"), "page_size", service.DefaultPageSize, details)
	if len(details) > 0 {
		apperr.Write(w, r, apperr.Validation(details))
		return
	}

	result, err := h.svc.List(r.Context(), service.ListRestaurantsParams{Sort: q.Get("sort"), Search: q.Get("q"), Page: page, PageSize: pageSize})
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(h.list(r, result.Items), result.Page, result.PageSize, result.Total))
}

func (h *RestaurantHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	items, err := h.svc.ListMine(r.Context(), user)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": h.list(r, items)})
}

// Create accepts multipart/form-data with the restaurant fields and one or more "images" files,
// so a restaurant can never exist without an image.
func (h *RestaurantHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	form, err := parseMultipart(w, r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	defer form.RemoveAll() //nolint:errcheck

	in, details := restaurantInputFromForm(form)
	if len(details) > 0 {
		apperr.Write(w, r, apperr.Validation(details))
		return
	}
	m, err := h.svc.Create(r.Context(), user, in, uploadsFromForm(form))
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/restaurants/"+m.ID.String())
	writeJSON(w, http.StatusCreated, h.toJSON(r, m))
}

type restaurantBody struct {
	Name                    *string `json:"name"`
	Description             *string `json:"description"`
	Capacity                *int    `json:"capacity"`
	OpeningTime             *string `json:"opening_time"`
	ClosingTime             *string `json:"closing_time"`
	Timezone                *string `json:"timezone"`
	CancellationNoticeHours *int    `json:"cancellation_notice_hours"`
}

// Update is a full replacement of the editable fields; every field is required.
func (h *RestaurantHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	id, err := pathID(r, "id", service.ErrRestaurantNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	var body restaurantBody
	if err := decodeJSON(w, r, &body); err != nil {
		apperr.Write(w, r, err)
		return
	}
	missing := map[string]string{}
	for field, present := range map[string]bool{
		"name": body.Name != nil, "description": body.Description != nil, "capacity": body.Capacity != nil,
		"opening_time": body.OpeningTime != nil, "closing_time": body.ClosingTime != nil,
		"timezone": body.Timezone != nil, "cancellation_notice_hours": body.CancellationNoticeHours != nil,
	} {
		if !present {
			missing[field] = "This field is required."
		}
	}
	if len(missing) > 0 {
		apperr.Write(w, r, apperr.Validation(missing))
		return
	}

	m, err := h.svc.Update(r.Context(), user, id, service.RestaurantInput{
		Name:                    *body.Name,
		Description:             *body.Description,
		Capacity:                *body.Capacity,
		OpeningTime:             *body.OpeningTime,
		ClosingTime:             *body.ClosingTime,
		Timezone:                *body.Timezone,
		CancellationNoticeHours: *body.CancellationNoticeHours,
	})
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, h.toJSON(r, m))
}

func (h *RestaurantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	id, err := pathID(r, "id", service.ErrRestaurantNotFound)
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

func (h *RestaurantHandler) AddImages(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	id, err := pathID(r, "id", service.ErrRestaurantNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	form, err := parseMultipart(w, r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	defer form.RemoveAll() //nolint:errcheck

	m, err := h.svc.AddImages(r.Context(), user, id, uploadsFromForm(form))
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, h.toJSON(r, m))
}

func (h *RestaurantHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	id, err := pathID(r, "id", service.ErrRestaurantNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	imageID, err := pathID(r, "imageId", service.ErrImageNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	if err := h.svc.DeleteImage(r.Context(), user, id, imageID); err != nil {
		apperr.Write(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseMultipart(w http.ResponseWriter, r *http.Request) (*multipart.Form, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartBody)
	mr, err := r.MultipartReader()
	if err != nil {
		return nil, apperr.Validation(map[string]string{"body": "Request must be multipart/form-data."})
	}
	// Files beyond 8 MB of memory are spooled to temporary files and removed by form.RemoveAll.
	form, err := mr.ReadForm(8 << 20)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return nil, apperr.Validation(map[string]string{"body": "Request body is too large."})
		}
		return nil, apperr.Validation(map[string]string{"body": "Could not read the form data."})
	}
	return form, nil
}

func uploadsFromForm(form *multipart.Form) []service.ImageUpload {
	files := form.File["images"]
	uploads := make([]service.ImageUpload, len(files))
	for i, fh := range files {
		uploads[i] = service.ImageUpload{
			Size: fh.Size,
			Open: func() (io.ReadCloser, error) { return fh.Open() },
		}
	}
	return uploads
}

func restaurantInputFromForm(form *multipart.Form) (service.RestaurantInput, map[string]string) {
	details := map[string]string{}
	value := func(name string) string {
		if v := form.Value[name]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	integer := func(name, message string) int {
		n, err := strconv.Atoi(strings.TrimSpace(value(name)))
		if err != nil {
			details[name] = message
		}
		return n
	}
	in := service.RestaurantInput{
		Name:                    value("name"),
		Description:             value("description"),
		Capacity:                integer("capacity", "Capacity must be a whole number."),
		OpeningTime:             value("opening_time"),
		ClosingTime:             value("closing_time"),
		Timezone:                value("timezone"),
		CancellationNoticeHours: integer("cancellation_notice_hours", "Cancellation notice must be a whole number of hours."),
	}
	return in, details
}
