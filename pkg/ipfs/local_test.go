package ipfs

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLocalStorage(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ipfs-local-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create local storage instance
	storage, err := NewLocalStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data
	testContent := "Hello, IPFS Local Storage!"
	reader := strings.NewReader(testContent)

	// Test Add
	cid, err := storage.Add(ctx, reader)
	if err != nil {
		t.Fatalf("Failed to add content: %v", err)
	}

	if !cid.Defined() {
		t.Fatal("CID is undefined")
	}

	t.Logf("Generated CID: %s", cid.String())

	// Test Exists (utility method)
	if !storage.Exists(ctx, cid) {
		t.Fatal("Content should exist after adding")
	}

	// Test Size (utility method)
	size, err := storage.Size(ctx, cid)
	if err != nil {
		t.Fatalf("Failed to get size: %v", err)
	}

	expectedSize := int64(len(testContent))
	if size != expectedSize {
		t.Fatalf("Expected size %d, got %d", expectedSize, size)
	}

	// Test Get
	readCloser, err := storage.Get(ctx, cid)
	if err != nil {
		t.Fatalf("Failed to get content: %v", err)
	}
	defer readCloser.Close()

	// Read the content back
	retrievedData, err := io.ReadAll(readCloser)
	if err != nil {
		t.Fatalf("Failed to read retrieved content: %v", err)
	}

	retrievedContent := string(retrievedData)
	if retrievedContent != testContent {
		t.Fatalf("Expected content %q, got %q", testContent, retrievedContent)
	}

	// Test Delete (utility method)
	err = storage.Delete(ctx, cid)
	if err != nil {
		t.Fatalf("Failed to delete content: %v", err)
	}

	// Verify content is deleted
	if storage.Exists(ctx, cid) {
		t.Fatal("Content should not exist after deletion")
	}

	// Test Get after deletion should fail
	_, err = storage.Get(ctx, cid)
	if err == nil {
		t.Fatal("Expected error when getting deleted content")
	}
}

func TestLocalStorageNonExistentContent(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ipfs-local-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create local storage instance
	storage, err := NewLocalStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a fake CID
	testContent := "fake content"
	reader := strings.NewReader(testContent)
	fakeCid, err := storage.Add(ctx, reader)
	if err != nil {
		t.Fatalf("Failed to create fake CID: %v", err)
	}

	// Delete the content
	err = storage.Delete(ctx, fakeCid)
	if err != nil {
		t.Fatalf("Failed to delete content: %v", err)
	}

	// Test Get with non-existent CID
	_, err = storage.Get(ctx, fakeCid)
	if err == nil {
		t.Fatal("Expected error when getting non-existent content")
	}

	// Test Size with non-existent CID
	_, err = storage.Size(ctx, fakeCid)
	if err == nil {
		t.Fatal("Expected error when getting size of non-existent content")
	}

	// Test Exists with non-existent CID
	if storage.Exists(ctx, fakeCid) {
		t.Fatal("Expected false for non-existent content")
	}
}

func TestLocalStorageDirectoryCreation(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ipfs-local-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with nested path that doesn't exist
	nestedPath := tmpDir + "/nested/path/that/does/not/exist"

	storage, err := NewLocalStorage(nestedPath)
	if err != nil {
		t.Fatalf("Failed to create LocalStorage with nested path: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Fatal("Expected directory to be created")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test that we can still add content
	testContent := "test content in nested directory"
	reader := strings.NewReader(testContent)

	cid, err := storage.Add(ctx, reader)
	if err != nil {
		t.Fatalf("Failed to add content to nested directory: %v", err)
	}

	// Verify we can retrieve it
	readCloser, err := storage.Get(ctx, cid)
	if err != nil {
		t.Fatalf("Failed to get content from nested directory: %v", err)
	}
	defer readCloser.Close()

	retrievedData, err := io.ReadAll(readCloser)
	if err != nil {
		t.Fatalf("Failed to read retrieved content: %v", err)
	}

	if string(retrievedData) != testContent {
		t.Fatalf("Expected content %q, got %q", testContent, string(retrievedData))
	}
}
