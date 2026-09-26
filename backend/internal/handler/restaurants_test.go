package handler_test

import (
	"net/http"
	"strings"
	"tabletalk/backend/internal/auth"
	"testing"
	"time"
)

func TestCreateRestaurant_OwnerComesFromToken(t *testing.T) {
	api := newTestAPI(t)
	// A forged owner_id form field must be ignored: the owner is the authenticated user.
	api.do(http.MethodGet, "/api/me", &api.bob, nil) // make bob exist
	r := api.createRestaurant(api.alice, map[string]string{"owner_id": api.bob.UserID.String()})

	if r.OwnerID != api.alice.UserID.String() {
		t.Fatalf("owner_id = %s, want alice %s", r.OwnerID, api.alice.UserID)
	}
	if !r.IsOwner || r.Capacity != 10 || r.OpeningTime != "10:00" || r.ClosingTime != "22:00" {
		t.Fatalf("unexpected restaurant %+v", r)
	}
	if r.AverageRating != nil || r.ReviewCount != 0 {
		t.Fatalf("new restaurant must have no rating, got %v/%d", r.AverageRating, r.ReviewCount)
	}
	if len(r.Images) != 1 || !strings.HasPrefix(r.Images[0].URL, "http://api.test/uploads/restaurants/") {
		t.Fatalf("unexpected images %+v", r.Images)
	}

	// The stored image is served back with a content type derived from its content.
	img := api.do(http.MethodGet, strings.TrimPrefix(r.Images[0].URL, "http://api.test"), nil, nil)
	if img.Status != http.StatusOK || img.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("image fetch: %d %s", img.Status, img.Header.Get("Content-Type"))
	}
}

func TestCreateRestaurant_Validation(t *testing.T) {
	api := newTestAPI(t)
	tests := []struct {
		name      string
		overrides map[string]string
		images    int
		files     [][]byte
		field     string
	}{
		{"no images", nil, 0, nil, "images"},
		{"cancellation notice below 2 hours", map[string]string{"cancellation_notice_hours": "1"}, 1, nil, "cancellation_notice_hours"},
		{"capacity zero", map[string]string{"capacity": "0"}, 1, nil, "capacity"},
		{"capacity not a number", map[string]string{"capacity": "ten"}, 1, nil, "capacity"},
		{"closing before opening", map[string]string{"opening_time": "22:00", "closing_time": "10:00"}, 1, nil, "closing_time"},
		{"bad time format", map[string]string{"opening_time": "25:00"}, 1, nil, "opening_time"},
		{"unknown timezone", map[string]string{"timezone": "Mars/Olympus"}, 1, nil, "timezone"},
		{"blank name", map[string]string{"name": "   "}, 1, nil, "name"},
		{"non-image file", nil, 0, [][]byte{[]byte("#!/bin/sh\necho not an image")}, "images"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := restaurantForm(t, tc.overrides, tc.images)
			if tc.files != nil {
				form.files = tc.files
			}
			resp := api.do(http.MethodPost, "/api/restaurants", &api.alice, form)
			if resp.Status != http.StatusBadRequest || resp.errorCode(t) != "VALIDATION_ERROR" {
				t.Fatalf("got %d %s", resp.Status, resp.Body)
			}
			if _, ok := resp.errorDetails(t)[tc.field]; !ok {
				t.Fatalf("expected detail for %q, got %s", tc.field, resp.Body)
			}
		})
	}
}

func TestRestaurantWrites_RequireAuthentication(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)

	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/restaurants"},
		{http.MethodPut, "/api/restaurants/" + r.ID},
		{http.MethodDelete, "/api/restaurants/" + r.ID},
		{http.MethodPost, "/api/restaurants/" + r.ID + "/images"},
		{http.MethodDelete, "/api/restaurants/" + r.ID + "/images/" + r.Images[0].ID},
		{http.MethodGet, "/api/me/restaurants"},
	} {
		if resp := api.do(tc.method, tc.path, nil, nil); resp.Status != http.StatusUnauthorized {
			t.Errorf("%s %s: got %d, want 401", tc.method, tc.path, resp.Status)
		}
	}
}

// Test #13: a user cannot modify another owner's restaurant.
func TestRestaurantWrites_NonOwnerForbidden(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)

	update := map[string]any{
		"name": "Hijacked", "description": "", "capacity": 99, "opening_time": "00:00",
		"closing_time": "23:59", "timezone": "Asia/Bangkok", "cancellation_notice_hours": 2,
	}
	for _, tc := range []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{"update", http.MethodPut, "/api/restaurants/" + r.ID, update},
		{"delete", http.MethodDelete, "/api/restaurants/" + r.ID, nil},
		{"add image", http.MethodPost, "/api/restaurants/" + r.ID + "/images", &multipartBody{files: [][]byte{pngImage(t)}}},
		{"delete image", http.MethodDelete, "/api/restaurants/" + r.ID + "/images/" + r.Images[0].ID, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := api.do(tc.method, tc.path, &api.bob, tc.body)
			if resp.Status != http.StatusForbidden || resp.errorCode(t) != "FORBIDDEN" {
				t.Fatalf("got %d %s", resp.Status, resp.Body)
			}
		})
	}

	var after restaurantResp
	api.do(http.MethodGet, "/api/restaurants/"+r.ID, nil, nil).decode(t, &after)
	if after.Name != r.Name || after.Capacity != r.Capacity || len(after.Images) != 1 {
		t.Fatalf("restaurant was modified by a non-owner: %+v", after)
	}
}

func TestUpdateRestaurant(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	body := map[string]any{
		"name": "Somtam House 2", "description": "New", "capacity": 20, "opening_time": "11:00",
		"closing_time": "23:00", "timezone": "Asia/Bangkok", "cancellation_notice_hours": 6,
	}

	resp := api.do(http.MethodPut, "/api/restaurants/"+r.ID, &api.alice, body)
	if resp.Status != http.StatusOK {
		t.Fatalf("update: %d %s", resp.Status, resp.Body)
	}
	var got restaurantResp
	resp.decode(t, &got)
	if got.Name != "Somtam House 2" || got.Capacity != 20 || got.CancellationNoticeHours != 6 || got.ClosingTime != "23:00" {
		t.Fatalf("unexpected %+v", got)
	}

	t.Run("owner_id in body is rejected", func(t *testing.T) {
		forged := map[string]any{"owner_id": api.bob.UserID.String()}
		for k, v := range body {
			forged[k] = v
		}
		resp := api.do(http.MethodPut, "/api/restaurants/"+r.ID, &api.alice, forged)
		if resp.Status != http.StatusBadRequest {
			t.Fatalf("got %d %s", resp.Status, resp.Body)
		}
	})
	t.Run("cancellation notice below minimum", func(t *testing.T) {
		bad := map[string]any{}
		for k, v := range body {
			bad[k] = v
		}
		bad["cancellation_notice_hours"] = 1
		resp := api.do(http.MethodPut, "/api/restaurants/"+r.ID, &api.alice, bad)
		if resp.Status != http.StatusBadRequest || resp.errorDetails(t)["cancellation_notice_hours"] == "" {
			t.Fatalf("got %d %s", resp.Status, resp.Body)
		}
	})
	t.Run("missing field", func(t *testing.T) {
		resp := api.do(http.MethodPut, "/api/restaurants/"+r.ID, &api.alice, map[string]any{"name": "x"})
		if resp.Status != http.StatusBadRequest || resp.errorDetails(t)["capacity"] == "" {
			t.Fatalf("got %d %s", resp.Status, resp.Body)
		}
	})
}

func TestUpdateRestaurant_MustKeepExistingReservationsValid(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	api.do(http.MethodGet, "/api/me", &api.bob, nil)
	// Upcoming: 6 people tomorrow 18:00–20:00. Past reservations and cancelled ones do not count.
	api.insertReservation(r.ID, api.bob, tomorrowAt(18, 0), tomorrowAt(20, 0), 6, "ACTIVE")
	api.insertReservation(r.ID, api.bob, time.Now().Add(-3*time.Hour), time.Now().Add(-2*time.Hour), 10, "ACTIVE")
	api.insertReservation(r.ID, api.bob, tomorrowAt(12, 0), tomorrowAt(13, 0), 10, "CANCELLED")

	update := func(capacity int, closing string) int {
		return api.do(http.MethodPut, "/api/restaurants/"+r.ID, &api.alice, map[string]any{
			"name": "Somtam House", "description": "", "capacity": capacity, "opening_time": "10:00",
			"closing_time": closing, "timezone": "Asia/Bangkok", "cancellation_notice_hours": 2,
		}).Status
	}

	if got := update(5, "22:00"); got != http.StatusConflict {
		t.Fatalf("capacity below booked party: got %d, want 409", got)
	}
	if got := update(10, "19:00"); got != http.StatusConflict {
		t.Fatalf("closing before booked end: got %d, want 409", got)
	}
	if got := update(6, "20:00"); got != http.StatusOK {
		t.Fatalf("change that still fits: got %d, want 200", got)
	}
}

func TestDeleteRestaurant(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	imagePath := strings.TrimPrefix(r.Images[0].URL, "http://api.test")
	api.do(http.MethodGet, "/api/me", &api.bob, nil)
	api.insertReservation(r.ID, api.bob, tomorrowAt(18, 0), tomorrowAt(20, 0), 2, "ACTIVE")

	resp := api.do(http.MethodDelete, "/api/restaurants/"+r.ID, &api.alice, nil)
	if resp.Status != http.StatusConflict || resp.errorCode(t) != "RESTAURANT_HAS_UPCOMING_RESERVATIONS" {
		t.Fatalf("delete with upcoming reservation: %d %s", resp.Status, resp.Body)
	}

	api.exec(`UPDATE reservations SET status = 'CANCELLED', cancelled_at = now()`)
	if resp := api.do(http.MethodDelete, "/api/restaurants/"+r.ID, &api.alice, nil); resp.Status != http.StatusNoContent {
		t.Fatalf("delete: %d %s", resp.Status, resp.Body)
	}
	if resp := api.do(http.MethodGet, "/api/restaurants/"+r.ID, nil, nil); resp.Status != http.StatusNotFound {
		t.Fatalf("get after delete: %d", resp.Status)
	}
	if resp := api.do(http.MethodGet, imagePath, nil, nil); resp.Status != http.StatusNotFound {
		t.Fatalf("image file should be removed, got %d", resp.Status)
	}
}

func TestRestaurantImages(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	path := "/api/restaurants/" + r.ID + "/images"

	resp := api.do(http.MethodDelete, path+"/"+r.Images[0].ID, &api.alice, nil)
	if resp.Status != http.StatusConflict || resp.errorCode(t) != "LAST_IMAGE" {
		t.Fatalf("deleting the last image: %d %s", resp.Status, resp.Body)
	}

	resp = api.do(http.MethodPost, path, &api.alice, &multipartBody{files: [][]byte{pngImage(t), pngImage(t)}})
	if resp.Status != http.StatusCreated {
		t.Fatalf("add images: %d %s", resp.Status, resp.Body)
	}
	var withThree restaurantResp
	resp.decode(t, &withThree)
	if len(withThree.Images) != 3 {
		t.Fatalf("want 3 images, got %d", len(withThree.Images))
	}

	if resp := api.do(http.MethodDelete, path+"/"+r.Images[0].ID, &api.alice, nil); resp.Status != http.StatusNoContent {
		t.Fatalf("delete image: %d %s", resp.Status, resp.Body)
	}

	tooMany := &multipartBody{}
	for range 9 {
		tooMany.files = append(tooMany.files, pngImage(t))
	}
	if resp := api.do(http.MethodPost, path, &api.alice, tooMany); resp.Status != http.StatusConflict {
		t.Fatalf("exceeding 10 images: got %d", resp.Status)
	}
}

func TestGetRestaurant(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)

	for _, tc := range []struct {
		name    string
		viewer  any
		isOwner bool
	}{
		{"anonymous", nil, false},
		{"owner", &api.alice, true},
		{"other user", &api.bob, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var resp response
			if tc.viewer == nil {
				resp = api.do(http.MethodGet, "/api/restaurants/"+r.ID, nil, nil)
			} else {
				resp = api.do(http.MethodGet, "/api/restaurants/"+r.ID, tc.viewer.(*auth.Identity), nil)
			}
			var got restaurantResp
			resp.decode(t, &got)
			if resp.Status != http.StatusOK || got.IsOwner != tc.isOwner {
				t.Fatalf("got %d is_owner=%v", resp.Status, got.IsOwner)
			}
		})
	}

	for _, path := range []string{"/api/restaurants/00000000-0000-0000-0000-000000000000", "/api/restaurants/not-a-uuid"} {
		if resp := api.do(http.MethodGet, path, nil, nil); resp.Status != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", path, resp.Status)
		}
	}
}

func TestListMyRestaurants(t *testing.T) {
	api := newTestAPI(t)
	mine := api.createRestaurant(api.alice, map[string]string{"name": "Alice's"})
	api.createRestaurant(api.bob, map[string]string{"name": "Bob's"})

	var body struct {
		Items []restaurantResp `json:"items"`
	}
	api.do(http.MethodGet, "/api/me/restaurants", &api.alice, nil).decode(t, &body)
	if len(body.Items) != 1 || body.Items[0].ID != mine.ID {
		t.Fatalf("want only alice's restaurant, got %+v", body.Items)
	}
}

func TestRestaurantTimestampsAreUTC(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	var raw map[string]any
	api.do(http.MethodGet, "/api/restaurants/"+r.ID, nil, nil).decode(t, &raw)
	for _, field := range []string{"created_at", "updated_at"} {
		if s, _ := raw[field].(string); !strings.HasSuffix(s, "Z") {
			t.Errorf("%s = %q, want a UTC timestamp", field, s)
		}
	}
}
