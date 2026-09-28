package handler_test

import (
	"net/http"
	"testing"
	"time"
)

type reservationResp struct {
	ID             string    `json:"id"`
	RestaurantID   string    `json:"restaurant_id"`
	RestaurantName string    `json:"restaurant_name"`
	Timezone       string    `json:"timezone"`
	PartySize      int       `json:"party_size"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	Date           string    `json:"date"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
	Status         string    `json:"status"`
}

func tomorrow() string { return tomorrowAt(0, 0).Format(time.DateOnly) }

func booking(party int, start, end string) map[string]any {
	return map[string]any{"party_size": party, "date": tomorrow(), "start_time": start, "end_time": end}
}

func TestCreateReservationAPI(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	path := "/api/restaurants/" + r.ID + "/reservations"

	resp := api.do(http.MethodPost, path, &api.bob, booking(4, "18:00", "20:00"))
	if resp.Status != http.StatusCreated {
		t.Fatalf("create: %d %s", resp.Status, resp.Body)
	}
	var got reservationResp
	resp.decode(t, &got)
	if got.Status != "ACTIVE" || got.PartySize != 4 || got.RestaurantName != "Somtam House" || got.Timezone != "Asia/Bangkok" {
		t.Fatalf("unexpected %+v", got)
	}
	if got.Date != tomorrow() || got.StartTime != "18:00" || got.EndTime != "20:00" {
		t.Fatalf("local fields: %+v", got)
	}
	if !got.StartAt.Equal(tomorrowAt(18, 0)) || got.StartAt.Location() != time.UTC {
		t.Fatalf("start_at = %v, want %v as UTC", got.StartAt, tomorrowAt(18, 0).UTC())
	}
	if resp.Header.Get("Location") != "/api/reservations/"+got.ID {
		t.Fatalf("Location = %q", resp.Header.Get("Location"))
	}

	// The owner may also book their own restaurant (assumption A7).
	if resp := api.do(http.MethodPost, path, &api.alice, booking(2, "12:00", "13:00")); resp.Status != http.StatusCreated {
		t.Fatalf("owner booking: %d %s", resp.Status, resp.Body)
	}
}

func TestCreateReservationAPI_Errors(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	path := "/api/restaurants/" + r.ID + "/reservations"
	api.do(http.MethodPost, path, &api.bob, booking(7, "17:00", "18:00"))

	withUser := booking(1, "12:00", "13:00")
	withUser["user_id"] = api.alice.UserID.String()

	tests := []struct {
		name       string
		user       bool
		path       string
		body       any
		wantStatus int
		wantCode   string
	}{
		{"unauthenticated", false, path, booking(2, "18:00", "19:00"), http.StatusUnauthorized, "UNAUTHORIZED"},
		{"unknown restaurant", true, "/api/restaurants/00000000-0000-0000-0000-000000000000/reservations", booking(2, "18:00", "19:00"), http.StatusNotFound, "RESTAURANT_NOT_FOUND"},
		{"forged user_id", true, path, withUser, http.StatusBadRequest, "VALIDATION_ERROR"},
		{"missing fields", true, path, map[string]any{"party_size": 2}, http.StatusBadRequest, "VALIDATION_ERROR"},
		{"party_size as string", true, path, `{"party_size":"2","date":"` + tomorrow() + `","start_time":"18:00","end_time":"19:00"}`, http.StatusBadRequest, "VALIDATION_ERROR"},
		{"party exceeds capacity", true, path, booking(11, "18:00", "19:00"), http.StatusBadRequest, "PARTY_SIZE_EXCEEDS_CAPACITY"},
		{"outside opening hours", true, path, booking(2, "21:00", "23:00"), http.StatusBadRequest, "OUTSIDE_OPENING_HOURS"},
		{"insufficient capacity", true, path, booking(5, "17:30", "18:30"), http.StatusConflict, "INSUFFICIENT_CAPACITY"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var resp response
			if tc.user {
				resp = api.do(http.MethodPost, tc.path, &api.bob, tc.body)
			} else {
				resp = api.do(http.MethodPost, tc.path, nil, tc.body)
			}
			if resp.Status != tc.wantStatus || resp.errorCode(t) != tc.wantCode {
				t.Fatalf("got %d %s, want %d %s", resp.Status, resp.Body, tc.wantStatus, tc.wantCode)
			}
		})
	}

	var count int
	if err := api.db.QueryRow(t.Context(), `SELECT COUNT(*) FROM reservations WHERE user_id = $1`, api.alice.UserID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("forged user_id must not create a reservation for alice (count=%d, err=%v)", count, err)
	}
}
