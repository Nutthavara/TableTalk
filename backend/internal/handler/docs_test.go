package handler_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"tabletalk/backend/api"
	"tabletalk/backend/internal/handler"
)

// The OpenAPI spec must describe exactly the routes the router serves, so the docs can't drift.
func TestSwaggerSpec_MatchesRouteTable(t *testing.T) {
	var spec struct {
		OpenAPI string                    `yaml:"openapi"`
		Paths   map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(api.Spec, &spec); err != nil {
		t.Fatalf("swagger.yaml is not valid YAML: %v", err)
	}
	if !strings.HasPrefix(spec.OpenAPI, "3.") {
		t.Fatalf("openapi = %q, want 3.x", spec.OpenAPI)
	}

	var documented []string
	for path, ops := range spec.Paths {
		if !strings.HasPrefix(path, "/api/") {
			continue
		}
		for method := range ops {
			if method == "parameters" {
				continue
			}
			documented = append(documented, strings.ToUpper(method)+" "+path)
		}
	}
	var served []string
	for _, rt := range handler.Routes(handler.Deps{}) {
		served = append(served, rt.Method+" "+rt.Path)
	}
	slices.Sort(documented)
	slices.Sort(served)

	for _, r := range served {
		if !slices.Contains(documented, r) {
			t.Errorf("route %s is missing from swagger.yaml", r)
		}
	}
	for _, d := range documented {
		if !slices.Contains(served, d) {
			t.Errorf("swagger.yaml documents %s, which the router does not serve", d)
		}
	}
}

func TestDocs_OnlyWhenEnabled(t *testing.T) {
	api := newTestAPI(t)
	if resp := api.do(http.MethodGet, "/docs", nil, nil); resp.Status != http.StatusNotFound {
		t.Fatalf("docs must be off by default, got %d", resp.Status)
	}

	router := handler.NewRouter(handler.Deps{DB: api.db, Auth: nil, Restaurants: &handler.RestaurantHandler{}, Reservations: &handler.ReservationHandler{}, Reviews: &handler.ReviewHandler{}, Docs: true})
	for path, wantType := range map[string]string{"/docs": "text/html", "/docs/swagger.yaml": "application/yaml"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), wantType) {
			t.Errorf("%s: got %d %q", path, rec.Code, rec.Header().Get("Content-Type"))
		}
	}
}
