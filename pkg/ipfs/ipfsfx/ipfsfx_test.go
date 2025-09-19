package ipfsfx

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLocalBackendConfig(t *testing.T) {
	// Test LocalConfig validation
	localConfig := &LocalConfig{}

	// Should fail validation with empty BasePath
	err := localConfig.Validate()
	if err == nil {
		t.Fatal("Expected validation error for empty BasePath")
	}

	// Should pass validation with valid BasePath
	localConfig.BasePath = "/tmp/test"
	err = localConfig.Validate()
	if err != nil {
		t.Fatalf("Expected no validation error, got: %v", err)
	}

	// Test GetBackend
	if localConfig.GetBackend() != "local" {
		t.Fatalf("Expected backend 'local', got: %s", localConfig.GetBackend())
	}
}

func TestConfigWithLocalBackend(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ipfs-local-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test valid local configuration
	config := &Config{
		Backend: "local",
		Local: &LocalConfig{
			BasePath: tmpDir,
		},
	}

	backendConfig, err := config.GetBackendConfig()
	if err != nil {
		t.Fatalf("Failed to get backend config: %v", err)
	}

	if backendConfig.GetBackend() != "local" {
		t.Fatalf("Expected backend 'local', got: %s", backendConfig.GetBackend())
	}

	// Test missing local configuration
	config = &Config{
		Backend: "local",
		Local:   nil,
	}

	_, err = config.GetBackendConfig()
	if err == nil {
		t.Fatal("Expected error for missing local configuration")
	}

	// Test invalid local configuration
	config = &Config{
		Backend: "local",
		Local: &LocalConfig{
			BasePath: "", // Empty path should fail validation
		},
	}

	_, err = config.GetBackendConfig()
	if err == nil {
		t.Fatal("Expected validation error for empty BasePath")
	}
}

func TestNewStorageWithLocalBackend(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ipfs-local-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Test creating storage with local backend
	config := &Config{
		Backend: "local",
		Local: &LocalConfig{
			BasePath: tmpDir,
		},
	}

	storage, err := NewStorage(config, logger)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test that the storage works
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testContent := "Hello from local storage test!"
	reader := strings.NewReader(testContent)

	// Add content
	cid, err := storage.Add(ctx, reader)
	if err != nil {
		t.Fatalf("Failed to add content: %v", err)
	}

	// Get content back
	readCloser, err := storage.Get(ctx, cid)
	if err != nil {
		t.Fatalf("Failed to get content: %v", err)
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

func TestUnsupportedBackend(t *testing.T) {
	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Test unsupported backend
	config := &Config{
		Backend: "unsupported",
	}

	_, err := NewStorage(config, logger)
	if err == nil {
		t.Fatal("Expected error for unsupported backend")
	}

	if !strings.Contains(err.Error(), "unsupported backend: unsupported") {
		t.Fatalf("Expected error message to contain unsupported backend info, got: %v", err)
	}
}

func TestAllBackendTypes(t *testing.T) {
	// Test that all backend types are properly handled in GetBackendConfig
	testCases := []struct {
		name          string
		config        *Config
		expectError   bool
		expectedError string
	}{
		{
			name: "valid kubo config",
			config: &Config{
				Backend: "kubo",
				Kubo: &KuboConfig{
					NodeURL: "http://localhost:5001",
				},
			},
			expectError: false,
		},
		{
			name: "missing kubo config",
			config: &Config{
				Backend: "kubo",
				Kubo:    nil,
			},
			expectError:   true,
			expectedError: "kubo configuration is required",
		},
		{
			name: "valid local config",
			config: &Config{
				Backend: "local",
				Local: &LocalConfig{
					BasePath: "/tmp/test",
				},
			},
			expectError: false,
		},
		{
			name: "missing local config",
			config: &Config{
				Backend: "local",
				Local:   nil,
			},
			expectError:   true,
			expectedError: "local configuration is required",
		},
		{
			name: "valid pdp config",
			config: &Config{
				Backend: "pdp",
				PDP: &PDPConfig{
					ServiceURL:  "http://localhost:8080",
					ServiceName: "test-service",
					PrivateKey:  "test-key",
					ProofSetID:  1,
				},
			},
			expectError: false,
		},
		{
			name: "missing pdp config",
			config: &Config{
				Backend: "pdp",
				PDP:     nil,
			},
			expectError:   true,
			expectedError: "pdp configuration is required",
		},
		{
			name: "unsupported backend",
			config: &Config{
				Backend: "unknown",
			},
			expectError:   true,
			expectedError: "unsupported backend: unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.config.GetBackendConfig()

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("Expected error to contain %q, got: %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
