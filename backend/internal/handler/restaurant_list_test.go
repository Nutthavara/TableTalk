package handler_test

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type listResp struct {
	Items      []restaurantResp `json:"items"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	Total      int              `json:"total"`
	TotalPages int              `json:"total_pages"`
}

// seedRatedRestaurants creates restaurants named after their key with the given review ratings,
// and returns name → id.
func seedRatedRestaurants(t *testing.T, api *testAPI, ratings map[string][]int) map[string]string {
	t.Helper()
	reviewers := make([]uuid.UUID, 6)
	for i := range reviewers {
		reviewers[i] = uuid.New()
		api.exec(`INSERT INTO users (id, username, display_name) VALUES ($1, $2, $2)`, reviewers[i], fmt.Sprintf("reviewer%d", i))
	}
	ids := map[string]string{}
	for _, name := range slices.Sorted(func(yield func(string) bool) {
		for k := range ratings {
			if !yield(k) {
				return
			}
		}
	}) {
		r := api.createRestaurant(api.alice, map[string]string{"name": name})
		ids[name] = r.ID
		for i, rating := range ratings[name] {
			api.exec(`INSERT INTO reviews (restaurant_id, user_id, rating) VALUES ($1, $2, $3)`, r.ID, reviewers[i], rating)
		}
	}
	return ids
}

func names(items []restaurantResp) []string {
	out := make([]string, len(items))
	for i, r := range items {
		out[i] = r.Name
	}
	return out
}

func TestListRestaurants_SortingAndTieBreakers(t *testing.T) {
	api := newTestAPI(t)
	ids := seedRatedRestaurants(t, api, map[string][]int{
		"five-one":     {5},          // 5.0, 1 review
		"high-four":    {5, 5, 4, 4}, // 4.5, 4 reviews
		"high-two-a":   {5, 4},       // 4.5, 2 reviews
		"high-two-b":   {4, 5},       // 4.5, 2 reviews: full tie with high-two-a → id ASC
		"rounds-four":  {5, 4, 4, 4}, // 4.25 → shown as 4.3, 4 reviews
		"rounds-three": {4, 4, 5},    // 4.33 → shown as 4.3, 3 reviews
		"three-four":   {3, 3, 3, 3}, // 3.0, 4 reviews
		"no-reviews":   {},
	})
	// Full ties are broken by restaurant id ascending.
	twoA, twoB := "high-two-a", "high-two-b"
	if ids[twoB] < ids[twoA] {
		twoA, twoB = twoB, twoA
	}

	tests := []struct {
		sort string
		want []string
	}{
		{"rating", []string{"five-one", "high-four", twoA, twoB, "rounds-four", "rounds-three", "three-four", "no-reviews"}},
		{"", []string{"five-one", "high-four", twoA, twoB, "rounds-four", "rounds-three", "three-four", "no-reviews"}},
		{"reviews", []string{"high-four", "rounds-four", "three-four", "rounds-three", twoA, twoB, "five-one", "no-reviews"}},
	}
	for _, tc := range tests {
		t.Run("sort="+tc.sort, func(t *testing.T) {
			var got listResp
			resp := api.do(http.MethodGet, "/api/restaurants?page_size=50&sort="+tc.sort, nil, nil)
			if resp.Status != http.StatusOK {
				t.Fatalf("got %d %s", resp.Status, resp.Body)
			}
			resp.decode(t, &got)
			if !slices.Equal(names(got.Items), tc.want) {
				t.Fatalf("order\n got  %v\n want %v", names(got.Items), tc.want)
			}
		})
	}

	t.Run("rating values", func(t *testing.T) {
		var got listResp
		api.do(http.MethodGet, "/api/restaurants?page_size=50", nil, nil).decode(t, &got)
		byName := map[string]restaurantResp{}
		for _, r := range got.Items {
			byName[r.Name] = r
		}
		if r := byName["rounds-three"]; r.AverageRating == nil || *r.AverageRating != 4.3 || r.ReviewCount != 3 {
			t.Errorf("rounds-three: %v / %d", r.AverageRating, r.ReviewCount)
		}
		if r := byName["rounds-four"]; r.AverageRating == nil || *r.AverageRating != 4.3 || r.ReviewCount != 4 {
			t.Errorf("rounds-four: %v / %d", r.AverageRating, r.ReviewCount)
		}
		if r := byName["no-reviews"]; r.AverageRating != nil || r.ReviewCount != 0 {
			t.Errorf("no-reviews must have null rating, got %v / %d", r.AverageRating, r.ReviewCount)
		}
	})
}

func TestListRestaurants_PaginationIsStable(t *testing.T) {
	api := newTestAPI(t)
	ratings := map[string][]int{}
	for i := range 8 {
		ratings[fmt.Sprintf("r%d", i)] = []int{1 + i%3} // many ties to exercise the id tie-breaker
	}
	seedRatedRestaurants(t, api, ratings)

	for _, sort := range []string{"rating", "reviews"} {
		var all listResp
		api.do(http.MethodGet, "/api/restaurants?page_size=50&sort="+sort, nil, nil).decode(t, &all)

		var paged []string
		for page := 1; page <= 3; page++ {
			var p listResp
			api.do(http.MethodGet, fmt.Sprintf("/api/restaurants?sort=%s&page_size=3&page=%d", sort, page), nil, nil).decode(t, &p)
			if p.Total != 8 || p.TotalPages != 3 || p.Page != page || p.PageSize != 3 {
				t.Fatalf("%s page %d metadata: %+v", sort, page, p)
			}
			paged = append(paged, names(p.Items)...)
		}
		if !slices.Equal(paged, names(all.Items)) {
			t.Fatalf("%s: pages %v do not match full list %v", sort, paged, names(all.Items))
		}
	}

	var beyond listResp
	api.do(http.MethodGet, "/api/restaurants?page=4&page_size=3", nil, nil).decode(t, &beyond)
	if len(beyond.Items) != 0 || beyond.Total != 8 {
		t.Fatalf("page past the end: %+v", beyond)
	}
}

func TestListRestaurants_InvalidParams(t *testing.T) {
	api := newTestAPI(t)
	for _, query := range []string{"sort=name", "page=0", "page=-1", "page=abc", "page_size=0", "page_size=51"} {
		resp := api.do(http.MethodGet, "/api/restaurants?"+query, nil, nil)
		if resp.Status != http.StatusBadRequest || resp.errorCode(t) != "VALIDATION_ERROR" {
			t.Errorf("%s: got %d %s", query, resp.Status, resp.Body)
		}
	}

	var empty listResp
	resp := api.do(http.MethodGet, "/api/restaurants", nil, nil)
	resp.decode(t, &empty)
	if resp.Status != http.StatusOK || empty.Items == nil || empty.Total != 0 || empty.PageSize != 12 {
		t.Fatalf("empty list: %d %s", resp.Status, resp.Body)
	}
}

func TestListRestaurants_Search(t *testing.T) {
	api := newTestAPI(t)
	seedRatedRestaurants(t, api, map[string][]int{
		"Baan Suan Thai Kitchen": {4},
		"Thai Street Food":       {5},
		"Trattoria Lupo":         {3},
		"100% Vegan":             {},
		"Noodle_Bar":             {},
		"ร้านข้าวต้ม":            {},
	})

	for _, tc := range []struct {
		q    string
		want []string
	}{
		{"thai", []string{"Thai Street Food", "Baan Suan Thai Kitchen"}}, // case-insensitive, still sorted by rating
		{"  THAI  ", []string{"Thai Street Food", "Baan Suan Thai Kitchen"}},
		{"lupo", []string{"Trattoria Lupo"}},
		{"%", []string{"100% Vegan"}}, // wildcards are matched literally
		{"_", []string{"Noodle_Bar"}},
		{"e_B", []string{"Noodle_Bar"}},
		{"ข้าว", []string{"ร้านข้าวต้ม"}},
		{"sushi", []string{}},
	} {
		t.Run(tc.q, func(t *testing.T) {
			var got listResp
			resp := api.do(http.MethodGet, "/api/restaurants?page_size=50&q="+url.QueryEscape(tc.q), nil, nil)
			if resp.Status != http.StatusOK {
				t.Fatalf("got %d %s", resp.Status, resp.Body)
			}
			resp.decode(t, &got)
			if !slices.Equal(names(got.Items), tc.want) || got.Total != len(tc.want) {
				t.Fatalf("q=%q: got %v (total %d), want %v", tc.q, names(got.Items), got.Total, tc.want)
			}
		})
	}

	if resp := api.do(http.MethodGet, "/api/restaurants?q="+strings.Repeat("a", 121), nil, nil); resp.Status != http.StatusBadRequest {
		t.Fatalf("overlong search: got %d", resp.Status)
	}
	var all listResp
	api.do(http.MethodGet, "/api/restaurants?q=", nil, nil).decode(t, &all)
	if all.Total != 6 {
		t.Fatalf("empty search must list everything, got %d", all.Total)
	}
}
