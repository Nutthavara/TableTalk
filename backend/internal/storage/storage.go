// Package storage stores uploaded images. Local disk is used in development; an object-storage
// implementation (S3/MinIO/GCS) can satisfy the same interface later.
package storage

import (
	"context"
	"io"
)

type Storage interface {
	// Put stores the content under key, replacing any existing object.
	Put(ctx context.Context, key string, r io.Reader, contentType string) error
	// Delete removes the object; deleting a missing object is not an error.
	Delete(ctx context.Context, key string) error
	// URL returns the public URL clients use to fetch the object.
	URL(key string) string
}
