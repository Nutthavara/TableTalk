package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// Test #12: the owner cannot review their own restaurant.
func TestCreateReview_OwnerCannotReviewOwnRestaurant(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10) // owned by alice

	if _, err := f.reviews.Create(f.ctx, f.alice, r, ReviewInput{Rating: 5, Comment: "Best in town"}); !errors.Is(err, ErrCannotReviewOwn) {
		t.Fatalf("owner review: got %v", err)
	}
	got, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: 4, Comment: "  Tasty  "})
	if err != nil {
		t.Fatalf("customer review: %v", err)
	}
	if got.UserID != f.bob.UserID || got.Rating != 4 || got.Comment != "Tasty" || got.AuthorName != "Bob" {
		t.Fatalf("unexpected %+v", got)
	}
}

// Test #15: rating must be an integer from 1 to 5.
func TestCreateReview_RatingRange(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	for _, rating := range []int{0, -1, 6, 10} {
		if _, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: rating}); errCode(err) != "VALIDATION_ERROR" {
			t.Errorf("rating %d: got %v", rating, err)
		}
	}
	if _, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: 1}); err != nil {
		t.Errorf("rating 1: %v", err)
	}
	if _, err := f.reviews.Create(f.ctx, f.carol, r, ReviewInput{Rating: 5}); err != nil {
		t.Errorf("rating 5: %v", err)
	}
	long := make([]rune, maxCommentLength+1)
	for i := range long {
		long[i] = 'ก'
	}
	if _, err := f.reviews.Create(f.ctx, f.bob, f.restaurant(10), ReviewInput{Rating: 3, Comment: string(long)}); errCode(err) != "VALIDATION_ERROR" {
		t.Errorf("comment too long: got %v", err)
	}
}

func TestCreateReview_OnePerUserPerRestaurant(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	if _, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: 5}); !errors.Is(err, ErrReviewAlreadyExists) {
		t.Fatalf("second review: got %v", err)
	}
	if _, err := f.reviews.Create(f.ctx, f.bob, f.restaurant(10), ReviewInput{Rating: 5}); err != nil {
		t.Fatalf("review of another restaurant: %v", err)
	}
	if _, err := f.reviews.Create(f.ctx, f.bob, uuid.New(), ReviewInput{Rating: 5}); !errors.Is(err, ErrRestaurantNotFound) {
		t.Fatalf("unknown restaurant: got %v", err)
	}
}

// Simultaneous duplicate submissions are stopped by the unique constraint and reported as 409.
func TestCreateReview_ConcurrentDuplicates(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	results := runConcurrently(10, func(int) error {
		_, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: 4})
		return err
	})
	created := 0
	for _, err := range results {
		switch {
		case err == nil:
			created++
		case !errors.Is(err, ErrReviewAlreadyExists):
			t.Errorf("unexpected error: %v", err)
		}
	}
	if created != 1 {
		t.Fatalf("%d reviews created, want 1", created)
	}
}

func TestUpdateAndDeleteReview(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	review, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: 2, Comment: "Slow"})
	if err != nil {
		t.Fatal(err)
	}

	// Only the author may edit or delete; the restaurant owner may not.
	for _, other := range []struct {
		name string
		f    func() error
	}{
		{"carol edits", func() error {
			_, err := f.reviews.Update(f.ctx, f.carol, review.ID, ReviewInput{Rating: 5})
			return err
		}},
		{"owner edits", func() error {
			_, err := f.reviews.Update(f.ctx, f.alice, review.ID, ReviewInput{Rating: 5})
			return err
		}},
		{"carol deletes", func() error { return f.reviews.Delete(f.ctx, f.carol, review.ID) }},
		{"owner deletes", func() error { return f.reviews.Delete(f.ctx, f.alice, review.ID) }},
	} {
		if err := other.f(); !errors.Is(err, ErrNotReviewAuthor) {
			t.Errorf("%s: got %v", other.name, err)
		}
	}

	updated, err := f.reviews.Update(f.ctx, f.bob, review.ID, ReviewInput{Rating: 4, Comment: "Better now"})
	if err != nil || updated.Rating != 4 || updated.Comment != "Better now" {
		t.Fatalf("update: %+v %v", updated, err)
	}
	if _, err := f.reviews.Update(f.ctx, f.bob, review.ID, ReviewInput{Rating: 9}); errCode(err) != "VALIDATION_ERROR" {
		t.Fatalf("update with bad rating: got %v", err)
	}

	if err := f.reviews.Delete(f.ctx, f.bob, review.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.reviews.Delete(f.ctx, f.bob, review.ID); !errors.Is(err, ErrReviewNotFound) {
		t.Fatalf("delete twice: got %v", err)
	}
	// After deleting, the user may review again.
	if _, err := f.reviews.Create(f.ctx, f.bob, r, ReviewInput{Rating: 3}); err != nil {
		t.Fatalf("review after delete: %v", err)
	}
}
