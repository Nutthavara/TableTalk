package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/model"
	"tabletalk/backend/internal/repository"
	"tabletalk/backend/internal/storage"
)

const (
	MinCancellationNoticeHours = 2
	maxCancellationNoticeHours = 24 * 365
	maxCapacity                = 10_000
	maxNameLength              = 120
	maxDescriptionLength       = 2000
	MaxImagesPerRestaurant     = 10
	MaxImageBytes              = 5 << 20
	DefaultTimezone            = "Asia/Bangkok"
)

// Allowed image types (detected from content, not from the client's filename or header).
var imageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type RestaurantInput struct {
	Name                    string
	Description             string
	Capacity                int
	OpeningTime             string
	ClosingTime             string
	Timezone                string
	CancellationNoticeHours int
}

type ImageUpload struct {
	Size int64
	Open func() (io.ReadCloser, error)
}

type RestaurantService struct {
	store *repository.Store
	files storage.Storage
	now   func() time.Time
}

func NewRestaurantService(store *repository.Store, files storage.Storage, now func() time.Time) *RestaurantService {
	return &RestaurantService{store: store, files: files, now: now}
}

func (s *RestaurantService) Get(ctx context.Context, id uuid.UUID) (model.Restaurant, error) {
	r, err := s.store.GetRestaurant(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Restaurant{}, ErrRestaurantNotFound
	}
	return r, err
}

func (s *RestaurantService) ListMine(ctx context.Context, user auth.Identity) ([]model.Restaurant, error) {
	return s.store.ListRestaurantsByOwner(ctx, user.UserID)
}

func (s *RestaurantService) ImageURL(img model.Image) string {
	return s.files.URL(img.StorageKey)
}

// Create makes the authenticated user the owner. At least one image is required.
func (s *RestaurantService) Create(ctx context.Context, user auth.Identity, in RestaurantInput, uploads []ImageUpload) (model.Restaurant, error) {
	r, details := validateRestaurantInput(in)
	if len(uploads) == 0 {
		details["images"] = "At least one image is required."
	} else if len(uploads) > MaxImagesPerRestaurant {
		details["images"] = fmt.Sprintf("At most %d images are allowed.", MaxImagesPerRestaurant)
	}
	if len(details) > 0 {
		return model.Restaurant{}, apperr.Validation(details)
	}
	r.OwnerID = user.UserID

	images, err := s.storeImages(ctx, uploads, 0)
	if err != nil {
		return model.Restaurant{}, err
	}

	var id uuid.UUID
	err = s.store.InTx(ctx, func(q *repository.Queries) error {
		if id, err = q.InsertRestaurant(ctx, r); err != nil {
			return err
		}
		for _, img := range images {
			if err := q.InsertImage(ctx, id, img); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.deleteFiles(ctx, images)
		return model.Restaurant{}, err
	}
	return s.Get(ctx, id)
}

// Update replaces the editable fields. Changes to capacity, hours or timezone are rejected when an
// ACTIVE reservation that has not finished yet would no longer fit.
func (s *RestaurantService) Update(ctx context.Context, user auth.Identity, id uuid.UUID, in RestaurantInput) (model.Restaurant, error) {
	err := s.store.InTx(ctx, func(q *repository.Queries) error {
		current, err := s.lockOwned(ctx, q, user, id)
		if err != nil {
			return err
		}
		updated, details := validateRestaurantInput(in)
		if len(details) > 0 {
			return apperr.Validation(details)
		}
		updated.ID, updated.OwnerID = current.ID, current.OwnerID

		if err := s.checkReservationsStillFit(ctx, q, updated); err != nil {
			return err
		}
		return q.UpdateRestaurant(ctx, updated)
	})
	if err != nil {
		return model.Restaurant{}, err
	}
	return s.Get(ctx, id)
}

// Delete removes the restaurant with its images, reviews and past reservations. It is refused while
// ACTIVE reservations have not finished, so customers are never left with an orphaned booking.
func (s *RestaurantService) Delete(ctx context.Context, user auth.Identity, id uuid.UUID) error {
	var images []model.Image
	err := s.store.InTx(ctx, func(q *repository.Queries) error {
		if _, err := s.lockOwned(ctx, q, user, id); err != nil {
			return err
		}
		active, err := q.ListActiveReservationsEndingAfter(ctx, id, s.now())
		if err != nil {
			return err
		}
		if len(active) > 0 {
			return ErrHasActiveReservations
		}
		if images, err = q.ListImages(ctx, id); err != nil {
			return err
		}
		return q.DeleteRestaurant(ctx, id)
	})
	if err != nil {
		return err
	}
	s.deleteFiles(ctx, images)
	return nil
}

func (s *RestaurantService) AddImages(ctx context.Context, user auth.Identity, id uuid.UUID, uploads []ImageUpload) (model.Restaurant, error) {
	if len(uploads) == 0 {
		return model.Restaurant{}, apperr.Validation(map[string]string{"images": "At least one image is required."})
	}

	var stored []model.Image
	err := s.store.InTx(ctx, func(q *repository.Queries) error {
		if _, err := s.lockOwned(ctx, q, user, id); err != nil {
			return err
		}
		existing, err := q.ListImages(ctx, id)
		if err != nil {
			return err
		}
		if len(existing)+len(uploads) > MaxImagesPerRestaurant {
			return ErrTooManyImages
		}
		nextPosition := 0
		for _, img := range existing {
			nextPosition = max(nextPosition, img.Position+1)
		}
		if stored, err = s.storeImages(ctx, uploads, nextPosition); err != nil {
			return err
		}
		for _, img := range stored {
			if err := q.InsertImage(ctx, id, img); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.deleteFiles(ctx, stored)
		return model.Restaurant{}, err
	}
	return s.Get(ctx, id)
}

func (s *RestaurantService) DeleteImage(ctx context.Context, user auth.Identity, id, imageID uuid.UUID) error {
	var removed model.Image
	err := s.store.InTx(ctx, func(q *repository.Queries) error {
		if _, err := s.lockOwned(ctx, q, user, id); err != nil {
			return err
		}
		images, err := q.ListImages(ctx, id)
		if err != nil {
			return err
		}
		found := false
		for _, img := range images {
			if img.ID == imageID {
				removed, found = img, true
			}
		}
		if !found {
			return ErrImageNotFound
		}
		if len(images) == 1 {
			return ErrLastImage
		}
		return q.DeleteImage(ctx, id, imageID)
	})
	if err != nil {
		return err
	}
	s.deleteFiles(ctx, []model.Image{removed})
	return nil
}

// lockOwned locks the restaurant row and verifies the authenticated user owns it.
func (s *RestaurantService) lockOwned(ctx context.Context, q *repository.Queries, user auth.Identity, id uuid.UUID) (model.Restaurant, error) {
	r, err := q.LockRestaurant(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Restaurant{}, ErrRestaurantNotFound
	}
	if err != nil {
		return model.Restaurant{}, err
	}
	if r.OwnerID != user.UserID {
		return model.Restaurant{}, ErrNotRestaurantOwner
	}
	return r, nil
}

func (s *RestaurantService) checkReservationsStillFit(ctx context.Context, q *repository.Queries, r model.Restaurant) error {
	active, err := q.ListActiveReservationsEndingAfter(ctx, r.ID, s.now())
	if err != nil || len(active) == 0 {
		return err
	}
	loc, err := time.LoadLocation(r.Timezone)
	if err != nil {
		return err
	}
	from, to := active[0].StartAt, active[0].EndAt
	for _, res := range active {
		if !withinOpeningHours(res.StartAt, res.EndAt, loc, r.OpeningTime, r.ClosingTime) {
			return ErrConflictsWithReservations
		}
		from, to = earlier(from, res.StartAt), later(to, res.EndAt)
	}
	if peakOccupancy(active, from, to) > r.Capacity {
		return ErrConflictsWithReservations
	}
	return nil
}

// storeImages validates each upload by content and writes it to storage. On error, files already
// written by this call are removed.
func (s *RestaurantService) storeImages(ctx context.Context, uploads []ImageUpload, firstPosition int) ([]model.Image, error) {
	stored := make([]model.Image, 0, len(uploads))
	for i, u := range uploads {
		img, err := s.storeImage(ctx, u, firstPosition+i)
		if err != nil {
			s.deleteFiles(ctx, stored)
			return nil, err
		}
		stored = append(stored, img)
	}
	return stored, nil
}

func (s *RestaurantService) storeImage(ctx context.Context, u ImageUpload, position int) (model.Image, error) {
	invalid := func(msg string) error { return apperr.Validation(map[string]string{"images": msg}) }
	if u.Size > MaxImageBytes {
		return model.Image{}, invalid("Each image must be 5 MB or smaller.")
	}

	f, err := u.Open()
	if err != nil {
		return model.Image{}, fmt.Errorf("open upload: %w", err)
	}
	defer f.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return model.Image{}, fmt.Errorf("read upload: %w", err)
	}
	head = head[:n]
	contentType := http.DetectContentType(head)
	ext, ok := imageExtensions[contentType]
	if !ok {
		return model.Image{}, invalid("Images must be JPEG, PNG or WebP.")
	}

	key := "restaurants/" + uuid.NewString() + ext
	// The size limit is enforced on the bytes actually read, not only on the declared size.
	body := io.LimitReader(io.MultiReader(bytes.NewReader(head), f), MaxImageBytes+1)
	counter := &countingReader{r: body}
	if err := s.files.Put(ctx, key, counter, contentType); err != nil {
		return model.Image{}, fmt.Errorf("store image: %w", err)
	}
	if counter.n > MaxImageBytes {
		_ = s.files.Delete(ctx, key)
		return model.Image{}, invalid("Each image must be 5 MB or smaller.")
	}
	return model.Image{StorageKey: key, ContentType: contentType, Position: position}, nil
}

func (s *RestaurantService) deleteFiles(ctx context.Context, images []model.Image) {
	for _, img := range images {
		if err := s.files.Delete(ctx, img.StorageKey); err != nil {
			slog.WarnContext(ctx, "could not delete image file", "key", img.StorageKey, "error", err)
		}
	}
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func validateRestaurantInput(in RestaurantInput) (model.Restaurant, map[string]string) {
	details := map[string]string{}
	r := model.Restaurant{
		Name:                    strings.TrimSpace(in.Name),
		Description:             strings.TrimSpace(in.Description),
		Capacity:                in.Capacity,
		Timezone:                strings.TrimSpace(in.Timezone),
		CancellationNoticeHours: in.CancellationNoticeHours,
	}

	if n := utf8.RuneCountInString(r.Name); n == 0 || n > maxNameLength {
		details["name"] = fmt.Sprintf("Name is required and must be at most %d characters.", maxNameLength)
	}
	if utf8.RuneCountInString(r.Description) > maxDescriptionLength {
		details["description"] = fmt.Sprintf("Description must be at most %d characters.", maxDescriptionLength)
	}
	if r.Capacity < 1 || r.Capacity > maxCapacity {
		details["capacity"] = fmt.Sprintf("Capacity must be between 1 and %d.", maxCapacity)
	}
	if r.CancellationNoticeHours < MinCancellationNoticeHours || r.CancellationNoticeHours > maxCancellationNoticeHours {
		details["cancellation_notice_hours"] = fmt.Sprintf("Cancellation notice must be at least %d hours.", MinCancellationNoticeHours)
	}

	if r.Timezone == "" {
		r.Timezone = DefaultTimezone
	}
	if !validTimezone(r.Timezone) {
		details["timezone"] = "Timezone must be a valid IANA name, e.g. Asia/Bangkok."
	}

	var err error
	if r.OpeningTime, err = model.ParseTimeOfDay(in.OpeningTime); err != nil {
		details["opening_time"] = "Opening time must be in HH:MM format."
	}
	if r.ClosingTime, err = model.ParseTimeOfDay(in.ClosingTime); err != nil {
		details["closing_time"] = "Closing time must be in HH:MM format."
	}
	if details["opening_time"] == "" && details["closing_time"] == "" && r.ClosingTime <= r.OpeningTime {
		details["closing_time"] = "Closing time must be after opening time (overnight hours are not supported)."
	}
	return r, details
}

func validTimezone(name string) bool {
	if name == "Local" {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}
