package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/auth/authtest"
	"tabletalk/backend/internal/handler"
	"tabletalk/backend/internal/middleware"
	"tabletalk/backend/internal/repository"
	"tabletalk/backend/internal/service"
	"tabletalk/backend/internal/storage"
	"tabletalk/backend/internal/testutil"
)

// testAPI drives the real router against a real database, with tokens signed by a test key.
type testAPI struct {
	t      *testing.T
	router http.Handler
	signer *authtest.Signer
	db     *pgxpool.Pool
	alice  auth.Identity
	bob    auth.Identity
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	slog.SetDefault(slog.New(slog.DiscardHandler)) // keep request logs out of test output
	db := testutil.DB(t, "handler")
	store := repository.NewStore(db)
	files, err := storage.NewLocal(t.TempDir(), "http://api.test/uploads")
	if err != nil {
		t.Fatal(err)
	}
	signer := authtest.NewSigner(t)
	reviews := service.NewReviewService(store)
	router := handler.NewRouter(handler.Deps{
		DB:           db,
		Auth:         middleware.NewAuthenticator(signer.Verifier(), store),
		Restaurants:  handler.NewRestaurantHandler(service.NewRestaurantService(store, files, time.Now), reviews),
		Reservations: handler.NewReservationHandler(service.NewReservationService(store, time.Now)),
		Reviews:      handler.NewReviewHandler(reviews),
		Uploads:      files.Handler(),
	})
	return &testAPI{
		t: t, router: router, signer: signer, db: db,
		alice: auth.Identity{UserID: uuid.New(), Username: "alice"},
		bob:   auth.Identity{UserID: uuid.New(), Username: "bob"},
	}
}

type response struct {
	Status int
	Body   []byte
	Header http.Header
}

func (r response) decode(t *testing.T, dst any) {
	t.Helper()
	if err := json.Unmarshal(r.Body, dst); err != nil {
		t.Fatalf("decode %s: %v", r.Body, err)
	}
}

func (r response) errorCode(t *testing.T) string {
	t.Helper()
	var body struct {
		Error struct {
			Code    string            `json:"code"`
			Details map[string]string `json:"details"`
		} `json:"error"`
	}
	r.decode(t, &body)
	return body.Error.Code
}

func (r response) errorDetails(t *testing.T) map[string]string {
	t.Helper()
	var body struct {
		Error struct {
			Details map[string]string `json:"details"`
		} `json:"error"`
	}
	r.decode(t, &body)
	return body.Error.Details
}

// do sends a request as user (nil = anonymous). body may be nil, a *multipartBody, or any JSON value.
func (a *testAPI) do(method, path string, user *auth.Identity, body any) response {
	a.t.Helper()
	var (
		reader      io.Reader
		contentType string
	)
	switch b := body.(type) {
	case nil:
	case *multipartBody:
		reader, contentType = b.build(a.t)
	case string:
		reader, contentType = bytes.NewBufferString(b), "application/json"
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			a.t.Fatal(err)
		}
		reader, contentType = bytes.NewReader(raw), "application/json"
	}

	req := httptest.NewRequest(method, path, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if user != nil {
		req.Header.Set("Authorization", "Bearer "+a.signer.Token(a.t, user.UserID, user.Username))
	}
	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, req)
	return response{Status: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
}

func (a *testAPI) exec(sql string, args ...any) {
	a.t.Helper()
	if _, err := a.db.Exec(context.Background(), sql, args...); err != nil {
		a.t.Fatalf("exec %q: %v", sql, err)
	}
}

type multipartBody struct {
	fields map[string]string
	files  [][]byte
}

func (m *multipartBody) build(t *testing.T) (io.Reader, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range m.fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for i, f := range m.files {
		part, err := w.CreateFormFile("images", "image"+string(rune('a'+i))+".png")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write(f)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, w.FormDataContentType()
}

func pngImage(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func restaurantForm(t *testing.T, overrides map[string]string, images int) *multipartBody {
	fields := map[string]string{
		"name":                      "Somtam House",
		"description":               "Isan food",
		"capacity":                  "10",
		"opening_time":              "10:00",
		"closing_time":              "22:00",
		"timezone":                  "Asia/Bangkok",
		"cancellation_notice_hours": "2",
	}
	for k, v := range overrides {
		fields[k] = v
	}
	files := make([][]byte, images)
	for i := range files {
		files[i] = pngImage(t)
	}
	return &multipartBody{fields: fields, files: files}
}

type restaurantResp struct {
	ID                      string   `json:"id"`
	OwnerID                 string   `json:"owner_id"`
	Name                    string   `json:"name"`
	Capacity                int      `json:"capacity"`
	OpeningTime             string   `json:"opening_time"`
	ClosingTime             string   `json:"closing_time"`
	CancellationNoticeHours int      `json:"cancellation_notice_hours"`
	AverageRating           *float64 `json:"average_rating"`
	ReviewCount             int      `json:"review_count"`
	IsOwner                 bool     `json:"is_owner"`
	Images                  []struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	} `json:"images"`
}

// createRestaurant creates a restaurant owned by user and returns it.
func (a *testAPI) createRestaurant(user auth.Identity, overrides map[string]string) restaurantResp {
	a.t.Helper()
	resp := a.do(http.MethodPost, "/api/restaurants", &user, restaurantForm(a.t, overrides, 1))
	if resp.Status != http.StatusCreated {
		a.t.Fatalf("create restaurant: %d %s", resp.Status, resp.Body)
	}
	var r restaurantResp
	resp.decode(a.t, &r)
	return r
}

// insertReservation writes a reservation directly (reservation endpoints arrive in a later phase).
func (a *testAPI) insertReservation(restaurantID string, user auth.Identity, start, end time.Time, size int, status string) {
	a.t.Helper()
	cancelledAt := any(nil)
	if status == "CANCELLED" {
		cancelledAt = time.Now()
	}
	a.exec(`INSERT INTO reservations (restaurant_id, user_id, party_size, start_at, end_at, status, cancelled_at)
	        VALUES ($1, $2, $3, $4, $5, $6, $7)`, restaurantID, user.UserID, size, start, end, status, cancelledAt)
}

// tomorrowAt returns tomorrow's date at hh:mm in Bangkok.
func tomorrowAt(hour, minute int) time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	d := time.Now().In(loc).AddDate(0, 0, 1)
	return time.Date(d.Year(), d.Month(), d.Day(), hour, minute, 0, 0, loc)
}
