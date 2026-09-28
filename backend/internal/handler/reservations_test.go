package handler_test

import (
	"net/http"
	"tabletalk/backend/internal/auth"
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

type reservationDetailResp struct {
	reservationResp
	CancellationDeadline time.Time  `json:"cancellation_deadline"`
	CanModify            bool       `json:"can_modify"`
	CancelledAt          *time.Time `json:"cancelled_at"`
}

func (a *testAPI) book(t *testing.T, restaurantID string, user *auth.Identity, body map[string]any) reservationDetailResp {
	t.Helper()
	resp := a.do(http.MethodPost, "/api/restaurants/"+restaurantID+"/reservations", user, body)
	if resp.Status != http.StatusCreated {
		t.Fatalf("book: %d %s", resp.Status, resp.Body)
	}
	var r reservationDetailResp
	resp.decode(t, &r)
	return r
}

func TestManageReservationAPI(t *testing.T) {
	api := newTestAPI(t)
	restaurant := api.createRestaurant(api.alice, nil)
	res := api.book(t, restaurant.ID, &api.bob, booking(2, "18:00", "20:00"))
	path := "/api/reservations/" + res.ID

	if !res.CanModify || !res.CancellationDeadline.Equal(tomorrowAt(16, 0)) {
		t.Fatalf("deadline/can_modify: %+v", res)
	}

	// Read and edit as the owner.
	if resp := api.do(http.MethodGet, path, &api.bob, nil); resp.Status != http.StatusOK {
		t.Fatalf("get: %d %s", resp.Status, resp.Body)
	}
	resp := api.do(http.MethodPut, path, &api.bob, booking(3, "19:00", "21:00"))
	var edited reservationDetailResp
	resp.decode(t, &edited)
	if resp.Status != http.StatusOK || edited.PartySize != 3 || edited.StartTime != "19:00" || edited.ID != res.ID {
		t.Fatalf("edit: %d %s", resp.Status, resp.Body)
	}

	forged := booking(3, "19:00", "21:00")
	forged["user_id"] = api.alice.UserID.String()
	if resp := api.do(http.MethodPut, path, &api.bob, forged); resp.Status != http.StatusBadRequest {
		t.Fatalf("forged user_id on edit: %d", resp.Status)
	}

	// My reservations.
	var list struct {
		Items []reservationDetailResp `json:"items"`
		Total int                     `json:"total"`
	}
	api.do(http.MethodGet, "/api/me/reservations?filter=upcoming", &api.bob, nil).decode(t, &list)
	if list.Total != 1 || list.Items[0].ID != res.ID {
		t.Fatalf("my reservations: %+v", list)
	}
	api.do(http.MethodGet, "/api/me/reservations", &api.alice, nil).decode(t, &list)
	if list.Total != 0 {
		t.Fatalf("alice must not see bob's reservation: %+v", list)
	}
	if resp := api.do(http.MethodGet, "/api/me/reservations?filter=everything", &api.bob, nil); resp.Status != http.StatusBadRequest {
		t.Fatalf("bad filter: %d", resp.Status)
	}

	// DELETE is a logical cancellation that returns the cancelled reservation.
	resp = api.do(http.MethodDelete, path, &api.bob, nil)
	var cancelled reservationDetailResp
	resp.decode(t, &cancelled)
	if resp.Status != http.StatusOK || cancelled.Status != "CANCELLED" || cancelled.CancelledAt == nil || cancelled.CanModify {
		t.Fatalf("cancel: %d %s", resp.Status, resp.Body)
	}
	var status string
	if err := api.db.QueryRow(t.Context(), `SELECT status FROM reservations WHERE id = $1`, res.ID).Scan(&status); err != nil || status != "CANCELLED" {
		t.Fatalf("row must be kept as CANCELLED, got %q %v", status, err)
	}
	if resp := api.do(http.MethodDelete, path, &api.bob, nil); resp.Status != http.StatusConflict || resp.errorCode(t) != "RESERVATION_NOT_ACTIVE" {
		t.Fatalf("second cancel: %d %s", resp.Status, resp.Body)
	}
}

// Test #14 over HTTP, including the restaurant owner who is not the reservation owner.
func TestManageReservationAPI_Authorization(t *testing.T) {
	api := newTestAPI(t)
	restaurant := api.createRestaurant(api.alice, nil)
	res := api.book(t, restaurant.ID, &api.bob, booking(2, "18:00", "20:00"))
	path := "/api/reservations/" + res.ID

	for _, tc := range []struct {
		method string
		body   any
	}{
		{http.MethodGet, nil},
		{http.MethodPut, booking(1, "12:00", "13:00")},
		{http.MethodDelete, nil},
	} {
		if resp := api.do(tc.method, path, nil, tc.body); resp.Status != http.StatusUnauthorized {
			t.Errorf("%s anonymous: got %d", tc.method, resp.Status)
		}
		if resp := api.do(tc.method, path, &api.alice, tc.body); resp.Status != http.StatusForbidden {
			t.Errorf("%s by restaurant owner: got %d %s", tc.method, resp.Status, resp.Body)
		}
		if resp := api.do(tc.method, "/api/reservations/00000000-0000-0000-0000-000000000000", &api.bob, tc.body); resp.Status != http.StatusNotFound {
			t.Errorf("%s unknown id: got %d", tc.method, resp.Status)
		}
	}
	if resp := api.do(http.MethodGet, "/api/me/reservations", nil, nil); resp.Status != http.StatusUnauthorized {
		t.Errorf("list anonymous: got %d", resp.Status)
	}

	var after reservationDetailResp
	api.do(http.MethodGet, path, &api.bob, nil).decode(t, &after)
	if after.Status != "ACTIVE" || after.PartySize != 2 || after.StartTime != "18:00" {
		t.Fatalf("reservation changed by another user: %+v", after)
	}
}
