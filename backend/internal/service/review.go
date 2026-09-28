package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/model"
	"tabletalk/backend/internal/repository"
)

const maxCommentLength = 2000

var (
	ErrReviewNotFound      = apperr.New(http.StatusNotFound, "REVIEW_NOT_FOUND", "Review not found.")
	ErrNotReviewAuthor     = apperr.New(http.StatusForbidden, "FORBIDDEN", "This review belongs to another user.")
	ErrCannotReviewOwn     = apperr.New(http.StatusForbidden, "CANNOT_REVIEW_OWN_RESTAURANT", "You cannot review your own restaurant.")
	ErrReviewAlreadyExists = apperr.New(http.StatusConflict, "REVIEW_ALREADY_EXISTS", "You have already reviewed this restaurant. Edit your existing review instead.")
)

type ReviewInput struct {
	Rating  int
	Comment string
}

type ReviewService struct {
	store *repository.Store
}

func NewReviewService(store *repository.Store) *ReviewService {
	return &ReviewService{store: store}
}

type ReviewPage struct {
	Items    []model.Review
	Page     int
	PageSize int
	Total    int
}

func (s *ReviewService) List(ctx context.Context, restaurantID uuid.UUID, page, pageSize int) (ReviewPage, error) {
	details := map[string]string{}
	if page < 1 {
		details["page"] = "Page must be 1 or greater."
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		details["page_size"] = fmt.Sprintf("Page size must be between 1 and %d.", MaxPageSize)
	}
	if len(details) > 0 {
		return ReviewPage{}, apperr.Validation(details)
	}
	if _, err := s.restaurantOwner(ctx, restaurantID); err != nil {
		return ReviewPage{}, err
	}
	items, total, err := s.store.ListReviews(ctx, restaurantID, pageSize, (page-1)*pageSize)
	if err != nil {
		return ReviewPage{}, err
	}
	return ReviewPage{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

// MyReviewID returns the user's review of the restaurant, or nil when there is none.
func (s *ReviewService) MyReviewID(ctx context.Context, user auth.Identity, restaurantID uuid.UUID) (*uuid.UUID, error) {
	id, err := s.store.ReviewIDByUser(ctx, restaurantID, user.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Create adds the user's single review of a restaurant they do not own.
func (s *ReviewService) Create(ctx context.Context, user auth.Identity, restaurantID uuid.UUID, in ReviewInput) (model.Review, error) {
	in, details := validateReviewInput(in)
	if len(details) > 0 {
		return model.Review{}, apperr.Validation(details)
	}
	owner, err := s.restaurantOwner(ctx, restaurantID)
	if err != nil {
		return model.Review{}, err
	}
	if owner == user.UserID {
		return model.Review{}, ErrCannotReviewOwn
	}

	// The (user_id, restaurant_id) unique constraint is the source of truth, including when two
	// requests race; no pre-check is needed.
	id, err := s.store.InsertReview(ctx, model.Review{RestaurantID: restaurantID, UserID: user.UserID, Rating: in.Rating, Comment: in.Comment})
	if repository.IsUniqueViolation(err) {
		return model.Review{}, ErrReviewAlreadyExists
	}
	if err != nil {
		return model.Review{}, err
	}
	return s.store.GetReview(ctx, id)
}

func (s *ReviewService) Update(ctx context.Context, user auth.Identity, id uuid.UUID, in ReviewInput) (model.Review, error) {
	if _, err := s.ownReview(ctx, user, id); err != nil {
		return model.Review{}, err
	}
	in, details := validateReviewInput(in)
	if len(details) > 0 {
		return model.Review{}, apperr.Validation(details)
	}
	if err := s.store.UpdateReview(ctx, id, in.Rating, in.Comment); err != nil {
		return model.Review{}, err
	}
	return s.store.GetReview(ctx, id)
}

func (s *ReviewService) Delete(ctx context.Context, user auth.Identity, id uuid.UUID) error {
	if _, err := s.ownReview(ctx, user, id); err != nil {
		return err
	}
	return s.store.DeleteReview(ctx, id)
}

// ownReview loads a review and checks the caller wrote it. Authorship never changes, so the
// check cannot be invalidated between this read and the following write.
func (s *ReviewService) ownReview(ctx context.Context, user auth.Identity, id uuid.UUID) (model.Review, error) {
	r, err := s.store.GetReview(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Review{}, ErrReviewNotFound
	}
	if err != nil {
		return model.Review{}, err
	}
	if r.UserID != user.UserID {
		return model.Review{}, ErrNotReviewAuthor
	}
	return r, nil
}

func (s *ReviewService) restaurantOwner(ctx context.Context, restaurantID uuid.UUID) (uuid.UUID, error) {
	owner, err := s.store.RestaurantOwner(ctx, restaurantID)
	if errors.Is(err, repository.ErrNotFound) {
		return uuid.Nil, ErrRestaurantNotFound
	}
	return owner, err
}

func validateReviewInput(in ReviewInput) (ReviewInput, map[string]string) {
	details := map[string]string{}
	in.Comment = strings.TrimSpace(in.Comment)
	if in.Rating < 1 || in.Rating > 5 {
		details["rating"] = "Rating must be a whole number from 1 to 5."
	}
	if utf8.RuneCountInString(in.Comment) > maxCommentLength {
		details["comment"] = fmt.Sprintf("Comment must be at most %d characters.", maxCommentLength)
	}
	return in, details
}
