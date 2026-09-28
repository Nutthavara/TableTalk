package handler_test

import (
	"net/http"
	"testing"
)

type reviewResp struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	AuthorName string `json:"author_name"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
	IsMine     bool   `json:"is_mine"`
}

func TestReviewAPI(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	path := "/api/restaurants/" + r.ID + "/reviews"

	// Test #12 over HTTP.
	resp := api.do(http.MethodPost, path, &api.alice, map[string]any{"rating": 5, "comment": "Mine is the best"})
	if resp.Status != http.StatusForbidden || resp.errorCode(t) != "CANNOT_REVIEW_OWN_RESTAURANT" {
		t.Fatalf("owner review: %d %s", resp.Status, resp.Body)
	}

	resp = api.do(http.MethodPost, path, &api.bob, map[string]any{"rating": 4, "comment": "Great somtam"})
	var created reviewResp
	resp.decode(t, &created)
	if resp.Status != http.StatusCreated || created.UserID != api.bob.UserID.String() || !created.IsMine {
		t.Fatalf("create: %d %s", resp.Status, resp.Body)
	}
	if resp := api.do(http.MethodPost, path, &api.bob, map[string]any{"rating": 5}); resp.Status != http.StatusConflict || resp.errorCode(t) != "REVIEW_ALREADY_EXISTS" {
		t.Fatalf("duplicate: %d %s", resp.Status, resp.Body)
	}

	// The aggregate on the restaurant reflects real reviews.
	var detail struct {
		restaurantResp
		MyReviewID *string `json:"my_review_id"`
	}
	api.do(http.MethodGet, "/api/restaurants/"+r.ID, &api.bob, nil).decode(t, &detail)
	if detail.AverageRating == nil || *detail.AverageRating != 4 || detail.ReviewCount != 1 {
		t.Fatalf("rating aggregate: %v / %d", detail.AverageRating, detail.ReviewCount)
	}
	if detail.MyReviewID == nil || *detail.MyReviewID != created.ID {
		t.Fatalf("my_review_id = %v, want %s", detail.MyReviewID, created.ID)
	}
	detail.MyReviewID = nil
	api.do(http.MethodGet, "/api/restaurants/"+r.ID, nil, nil).decode(t, &detail)
	if detail.MyReviewID != nil {
		t.Fatalf("anonymous viewer must not get my_review_id")
	}

	// List is public; is_mine depends on the viewer.
	var list struct {
		Items []reviewResp `json:"items"`
		Total int          `json:"total"`
	}
	api.do(http.MethodGet, path, nil, nil).decode(t, &list)
	if list.Total != 1 || list.Items[0].IsMine || list.Items[0].AuthorName != "bob" {
		t.Fatalf("anonymous list: %+v", list)
	}

	reviewPath := "/api/reviews/" + created.ID
	resp = api.do(http.MethodPut, reviewPath, &api.bob, map[string]any{"rating": 2, "comment": "Went downhill"})
	var updated reviewResp
	resp.decode(t, &updated)
	if resp.Status != http.StatusOK || updated.Rating != 2 {
		t.Fatalf("update: %d %s", resp.Status, resp.Body)
	}
	api.do(http.MethodGet, "/api/restaurants/"+r.ID, nil, nil).decode(t, &detail)
	if *detail.AverageRating != 2 {
		t.Fatalf("aggregate after update: %v", *detail.AverageRating)
	}

	if resp := api.do(http.MethodDelete, reviewPath, &api.bob, nil); resp.Status != http.StatusNoContent {
		t.Fatalf("delete: %d %s", resp.Status, resp.Body)
	}
	api.do(http.MethodGet, "/api/restaurants/"+r.ID, nil, nil).decode(t, &detail)
	if detail.AverageRating != nil || detail.ReviewCount != 0 {
		t.Fatalf("no reviews must mean null rating, got %v / %d", detail.AverageRating, detail.ReviewCount)
	}
}

func TestReviewAPI_ValidationAndAuthorization(t *testing.T) {
	api := newTestAPI(t)
	r := api.createRestaurant(api.alice, nil)
	path := "/api/restaurants/" + r.ID + "/reviews"
	resp := api.do(http.MethodPost, path, &api.bob, map[string]any{"rating": 3})
	var review reviewResp
	resp.decode(t, &review)
	reviewPath := "/api/reviews/" + review.ID

	// Test #15 over HTTP: out of range, fractional, wrong type and missing ratings.
	for _, body := range []string{`{"rating":0}`, `{"rating":6}`, `{"rating":4.5}`, `{"rating":"5"}`, `{"comment":"no rating"}`} {
		if resp := api.do(http.MethodPut, reviewPath, &api.bob, body); resp.Status != http.StatusBadRequest {
			t.Errorf("PUT %s: got %d %s", body, resp.Status, resp.Body)
		}
	}
	forged := map[string]any{"rating": 5, "user_id": api.alice.UserID.String()}
	if resp := api.do(http.MethodPut, reviewPath, &api.bob, forged); resp.Status != http.StatusBadRequest {
		t.Errorf("forged user_id: got %d", resp.Status)
	}

	for _, tc := range []struct {
		method, path string
		body         any
	}{
		{http.MethodPost, path, map[string]any{"rating": 5}},
		{http.MethodPut, reviewPath, map[string]any{"rating": 5}},
		{http.MethodDelete, reviewPath, nil},
	} {
		if resp := api.do(tc.method, tc.path, nil, tc.body); resp.Status != http.StatusUnauthorized {
			t.Errorf("%s anonymous: got %d", tc.method, resp.Status)
		}
	}
	for _, tc := range []struct {
		method string
		body   any
	}{
		{http.MethodPut, map[string]any{"rating": 1}},
		{http.MethodDelete, nil},
	} {
		if resp := api.do(tc.method, reviewPath, &api.alice, tc.body); resp.Status != http.StatusForbidden {
			t.Errorf("%s by non-author: got %d", tc.method, resp.Status)
		}
	}
	for _, p := range []string{"/api/restaurants/00000000-0000-0000-0000-000000000000/reviews", "/api/restaurants/bad/reviews"} {
		if resp := api.do(http.MethodGet, p, nil, nil); resp.Status != http.StatusNotFound {
			t.Errorf("GET %s: got %d", p, resp.Status)
		}
	}
	if resp := api.do(http.MethodDelete, "/api/reviews/00000000-0000-0000-0000-000000000000", &api.bob, nil); resp.Status != http.StatusNotFound {
		t.Errorf("delete unknown: got %d", resp.Status)
	}

	var after reviewResp
	var list struct{ Items []reviewResp }
	api.do(http.MethodGet, path, nil, nil).decode(t, &list)
	after = list.Items[0]
	if after.Rating != 3 {
		t.Fatalf("review changed by failed requests: %+v", after)
	}
}
