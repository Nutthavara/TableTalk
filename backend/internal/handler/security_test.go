package handler_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"tabletalk/backend/internal/handler"
)

// Every route that changes data must require authentication, and anonymous access is limited to
// reads. Checked against the router's route table so new routes are covered automatically.
func TestAllRoutes_WritesRequireAuthentication(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)

	for _, rt := range handler.Routes(handler.Deps{}) {
		if rt.Auth == handler.Public && rt.Method != http.MethodGet {
			t.Errorf("%s %s is public but is not a read", rt.Method, rt.Path)
		}
		if rt.Auth != handler.Required {
			continue
		}
		// Use real IDs so a missing auth check could not hide behind a 404.
		path := strings.ReplaceAll(rt.ExamplePath(r.ID), r.ID+"/images/"+r.ID, r.ID+"/images/"+r.Images[0].ID)
		resp := api.do(rt.Method, path, nil, nil)
		if resp.Status != http.StatusUnauthorized || resp.errorCode(t) != "UNAUTHORIZED" {
			t.Errorf("%s %s anonymous: got %d %s", rt.Method, path, resp.Status, resp.Body)
		}
	}
}

// Identity and ownership fields sent by the client are rejected, never trusted.
func TestForgedIdentityFieldsAreRejected(t *testing.T) {
	api := newTestAPI(t)
	restaurant := api.createRestaurant(api.alice, nil)
	other := api.createRestaurant(api.alice, map[string]string{"name": "Other"})
	res := api.book(t, restaurant.ID, &api.bob, booking(2, "18:00", "20:00"))
	review := api.do(http.MethodPost, "/api/restaurants/"+restaurant.ID+"/reviews", &api.bob, map[string]any{"rating": 4})
	var rv reviewResp
	review.decode(t, &rv)

	with := func(base map[string]any, key string, value any) map[string]any {
		out := map[string]any{key: value}
		for k, v := range base {
			out[k] = v
		}
		return out
	}
	alice := api.alice.UserID.String()

	for _, tc := range []struct {
		name, method, path string
		body               any
	}{
		{"reservation create user_id", http.MethodPost, "/api/restaurants/" + restaurant.ID + "/reservations", with(booking(1, "12:00", "13:00"), "user_id", alice)},
		{"reservation edit user_id", http.MethodPut, "/api/reservations/" + res.ID, with(booking(2, "18:00", "20:00"), "user_id", alice)},
		{"reservation edit restaurant_id", http.MethodPut, "/api/reservations/" + res.ID, with(booking(2, "18:00", "20:00"), "restaurant_id", other.ID)},
		{"reservation edit status", http.MethodPut, "/api/reservations/" + res.ID, with(booking(2, "18:00", "20:00"), "status", "CANCELLED")},
		{"review create user_id", http.MethodPost, "/api/restaurants/" + other.ID + "/reviews", map[string]any{"rating": 5, "user_id": alice}},
		{"review edit restaurant_id", http.MethodPut, "/api/reviews/" + rv.ID, map[string]any{"rating": 5, "restaurant_id": other.ID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := api.do(tc.method, tc.path, &api.bob, tc.body)
			if resp.Status != http.StatusBadRequest || resp.errorCode(t) != "VALIDATION_ERROR" {
				t.Fatalf("got %d %s", resp.Status, resp.Body)
			}
		})
	}

	var (
		reservationUser, reservationRestaurant, reservationStatus string
		reviews                                                   int
	)
	ctx := context.Background()
	if err := api.db.QueryRow(ctx, `SELECT user_id::text, restaurant_id::text, status FROM reservations WHERE id = $1`, res.ID).
		Scan(&reservationUser, &reservationRestaurant, &reservationStatus); err != nil {
		t.Fatal(err)
	}
	if reservationUser != api.bob.UserID.String() || reservationRestaurant != restaurant.ID || reservationStatus != "ACTIVE" {
		t.Fatalf("reservation changed: %s %s %s", reservationUser, reservationRestaurant, reservationStatus)
	}
	if err := api.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE user_id = $1 OR restaurant_id = $2`, api.alice.UserID, other.ID).Scan(&reviews); err != nil || reviews != 0 {
		t.Fatalf("forged review was stored (count=%d, err=%v)", reviews, err)
	}
}

func TestAPIResponses_SecurityHeadersAndErrorFormat(t *testing.T) {
	api := newTestAPI(t)
	for _, path := range []string{"/api/restaurants", "/api/nope", "/api/restaurants/not-a-uuid"} {
		resp := api.do(http.MethodGet, path, nil, nil)
		if resp.Header.Get("X-Content-Type-Options") != "nosniff" || resp.Header.Get("Cache-Control") != "no-store" {
			t.Errorf("%s: missing security headers %v", path, resp.Header)
		}
		if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
			t.Errorf("%s: content type %q", path, resp.Header.Get("Content-Type"))
		}
	}
	if resp := api.do(http.MethodPatch, "/api/restaurants", &api.bob, "{}"); resp.errorCode(t) == "" {
		t.Errorf("unsupported method must still use the error format: %d %s", resp.Status, resp.Body)
	}
}
