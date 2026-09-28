package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/model"
	"tabletalk/backend/internal/service"
)

type ReservationHandler struct {
	svc *service.ReservationService
}

func NewReservationHandler(svc *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{svc: svc}
}

// reservationJSON gives both the unambiguous UTC instants and the restaurant-local date/times the
// customer booked with.
type reservationJSON struct {
	ID             string     `json:"id"`
	RestaurantID   string     `json:"restaurant_id"`
	RestaurantName string     `json:"restaurant_name"`
	Timezone       string     `json:"timezone"`
	PartySize      int        `json:"party_size"`
	StartAt        time.Time  `json:"start_at"`
	EndAt          time.Time  `json:"end_at"`
	Date           string     `json:"date"`
	StartTime      string     `json:"start_time"`
	EndTime        string     `json:"end_time"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CancelledAt    *time.Time `json:"cancelled_at"`
	// Last instant the reservation can be edited or cancelled, and whether that is still possible now.
	CancellationNoticeHours int       `json:"cancellation_notice_hours"`
	CancellationDeadline    time.Time `json:"cancellation_deadline"`
	CanModify               bool      `json:"can_modify"`
}

func (h *ReservationHandler) toJSON(m model.Reservation) reservationJSON {
	loc, err := time.LoadLocation(m.RestaurantTimezone)
	if err != nil {
		loc = time.UTC
	}
	start, end := m.StartAt.In(loc), m.EndAt.In(loc)
	var cancelledAt *time.Time
	if m.CancelledAt != nil {
		t := m.CancelledAt.UTC()
		cancelledAt = &t
	}
	return reservationJSON{
		ID:             m.ID.String(),
		RestaurantID:   m.RestaurantID.String(),
		RestaurantName: m.RestaurantName,
		Timezone:       m.RestaurantTimezone,
		PartySize:      m.PartySize,
		StartAt:        m.StartAt.UTC(),
		EndAt:          m.EndAt.UTC(),
		Date:           start.Format(time.DateOnly),
		StartTime:      start.Format("15:04"),
		EndTime:        end.Format("15:04"),
		Status:         string(m.Status),
		CreatedAt:      m.CreatedAt.UTC(),
		UpdatedAt:      m.UpdatedAt.UTC(),
		CancelledAt:    cancelledAt,

		CancellationNoticeHours: m.CancellationNoticeHours,
		CancellationDeadline:    m.ModificationDeadline().UTC(),
		CanModify:               h.svc.CanModify(m),
	}
}

type reservationBody struct {
	PartySize *int    `json:"party_size"`
	Date      *string `json:"date"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
}

// input requires every field (create and full-replacement update).
func (b reservationBody) input() (service.ReservationInput, error) {
	missing := map[string]string{}
	if b.PartySize == nil {
		missing["party_size"] = "This field is required."
	}
	if b.Date == nil {
		missing["date"] = "This field is required."
	}
	if b.StartTime == nil {
		missing["start_time"] = "This field is required."
	}
	if b.EndTime == nil {
		missing["end_time"] = "This field is required."
	}
	if len(missing) > 0 {
		return service.ReservationInput{}, apperr.Validation(missing)
	}
	return service.ReservationInput{PartySize: *b.PartySize, Date: *b.Date, StartTime: *b.StartTime, EndTime: *b.EndTime}, nil
}

func (h *ReservationHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	var body reservationBody
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
	w.Header().Set("Location", "/api/reservations/"+m.ID.String())
	writeJSON(w, http.StatusCreated, h.toJSON(m))
}

// ListMine handles GET /api/me/reservations?filter=upcoming|past|cancelled&page=&page_size=.
func (h *ReservationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, err := identity(r)
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

	result, err := h.svc.ListMine(r.Context(), user, service.ListReservationsParams{Filter: q.Get("filter"), Page: page, PageSize: pageSize})
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	items := make([]reservationJSON, len(result.Items))
	for i, m := range result.Items {
		items[i] = h.toJSON(m)
	}
	writeJSON(w, http.StatusOK, newPage(items, result.Page, result.PageSize, result.Total))
}

func (h *ReservationHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.withReservation(w, r, func(user auth.Identity, id uuid.UUID) (model.Reservation, error) {
		return h.svc.Get(r.Context(), user, id)
	})
}

func (h *ReservationHandler) Update(w http.ResponseWriter, r *http.Request) {
	h.withReservation(w, r, func(user auth.Identity, id uuid.UUID) (model.Reservation, error) {
		var body reservationBody
		if err := decodeJSON(w, r, &body); err != nil {
			return model.Reservation{}, err
		}
		in, err := body.input()
		if err != nil {
			return model.Reservation{}, err
		}
		return h.svc.Update(r.Context(), user, id, in)
	})
}

// Cancel handles DELETE /api/reservations/{id}: a logical cancellation that returns the cancelled
// reservation; nothing is deleted.
func (h *ReservationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.withReservation(w, r, func(user auth.Identity, id uuid.UUID) (model.Reservation, error) {
		return h.svc.Cancel(r.Context(), user, id)
	})
}

// withReservation resolves the caller and the {id} path value, runs fn and writes the result as 200.
func (h *ReservationHandler) withReservation(w http.ResponseWriter, r *http.Request, fn func(auth.Identity, uuid.UUID) (model.Reservation, error)) {
	user, err := identity(r)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	id, err := pathID(r, "id", service.ErrReservationNotFound)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	m, err := fn(user, id)
	if err != nil {
		apperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, h.toJSON(m))
}
