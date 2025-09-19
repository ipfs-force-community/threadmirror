package ipfs

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// LocalStorage implements Storage interface using local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new LocalStorage instance
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// Add adds content to local storage and returns a CID
func (l *LocalStorage) Add(ctx context.Context, content io.ReadSeeker) (cid.Cid, error) {
	// Read all content to calculate hash
	data, err := io.ReadAll(content)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to read content: %w", err)
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(data)

	// Create multihash
	mh, err := multihash.Encode(hash[:], multihash.SHA2_256)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create multihash: %w", err)
	}

	// Create CID
	c := cid.NewCidV1(cid.Raw, mh)

	// Create file path based on CID
	filePath := l.getFilePath(c)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return cid.Undef, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write content to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return cid.Undef, fmt.Errorf("failed to write file: %w", err)
	}

	return c, nil
}

// Get retrieves content from local storage by CID
func (l *LocalStorage) Get(ctx context.Context, c cid.Cid) (io.ReadCloser, error) {
	filePath := l.getFilePath(c)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("content not found for CID %s", c.String())
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// getFilePath returns the file path for a given CID
func (l *LocalStorage) getFilePath(c cid.Cid) string {
	cidStr := c.String()
	// Create a directory structure based on the first few characters of the CID
	// This helps distribute files across directories to avoid having too many files in one directory
	return filepath.Join(l.basePath, cidStr[:2], cidStr[2:4], cidStr)
}

// Delete removes content from local storage by CID (utility method)
func (l *LocalStorage) Delete(ctx context.Context, c cid.Cid) error {
	filePath := l.getFilePath(c)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove empty directories (ignore errors)
	dir := filepath.Dir(filePath)
	os.Remove(dir)               // Remove immediate parent if empty
	os.Remove(filepath.Dir(dir)) // Remove grandparent if empty

	return nil
}

// Exists checks if content exists in local storage for the given CID (utility method)
func (l *LocalStorage) Exists(ctx context.Context, c cid.Cid) bool {
	filePath := l.getFilePath(c)
	_, err := os.Stat(filePath)
	return err == nil
}

// Size returns the size of content for the given CID (utility method)
func (l *LocalStorage) Size(ctx context.Context, c cid.Cid) (int64, error) {
	filePath := l.getFilePath(c)

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("content not found for CID %s", c.String())
		}
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return info.Size(), nil
}
