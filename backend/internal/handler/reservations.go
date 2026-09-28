package handler

import (
	"net/http"
	"time"

	"tabletalk/backend/internal/apperr"
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
}

func reservationToJSON(m model.Reservation) reservationJSON {
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
	writeJSON(w, http.StatusCreated, reservationToJSON(m))
}
