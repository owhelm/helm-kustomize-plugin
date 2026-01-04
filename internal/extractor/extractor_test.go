package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTempDir(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v, want nil", err)
	}
	defer tempDir.Cleanup()

	if tempDir.Path == "" {
		t.Error("Expected non-empty path")
	}

	// Verify directory exists
	if _, err := os.Stat(tempDir.Path); os.IsNotExist(err) {
		t.Errorf("Directory does not exist: %s", tempDir.Path)
	}
}

func TestTempDir_Cleanup(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}

	path := tempDir.Path

	// Verify directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Directory should exist: %s", path)
	}

	// Cleanup
	tempDir.Cleanup()

	// Verify directory no longer exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Directory should not exist after cleanup: %s", path)
	}
}

func TestTempDir_ExtractFiles(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	files := map[string]string{
		"kustomization.yaml":       "resources:\n- all.yaml\n",
		"patches/deployment.yaml":  "apiVersion: apps/v1\nkind: Deployment\n",
		"overlays/prod/patch.yaml": "spec:\n  replicas: 3\n",
	}

	err = tempDir.ExtractFiles(files)
	if err != nil {
		t.Fatalf("ExtractFiles() error = %v, want nil", err)
	}

	// Verify all files were created
	for filePath, expectedContent := range files {
		fullPath := filepath.Join(tempDir.Path, filePath)

		// Check file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("File should exist: %s", filePath)
			continue
		}

		// Check content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", filePath, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("File %s content = %q, want %q", filePath, string(content), expectedContent)
		}
	}
}

func TestTempDir_ExtractFiles_InvalidPath(t *testing.T) {
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	files := map[string]string{
		"../../../etc/passwd": "malicious content",
	}

	err = tempDir.ExtractFiles(files)
	if err == nil {
		t.Fatal("ExtractFiles() should return error for directory traversal attempt")
	}
}

func TestTempDir_WriteFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content []byte
		wantErr bool
	}{
		{
			name:    "simple filename",
			path:    "file.yaml",
			content: []byte("test content"),
			wantErr: false,
		},
		{
			name:    "nested path",
			path:    "patches/deployment.yaml",
			content: []byte("nested content"),
			wantErr: false,
		},
		{
			name:    "deeply nested path",
			path:    "overlays/production/patches/deployment.yaml",
			content: []byte("deeply nested content"),
			wantErr: false,
		},
		{
			name:    "parent directory traversal",
			path:    "../etc/passwd",
			content: []byte("malicious"),
			wantErr: true,
		},
		{
			name:    "traversal in middle",
			path:    "foo/../../../etc/passwd",
			content: []byte("malicious"),
			wantErr: true,
		},
		{
			name:    "absolute path",
			path:    "/etc/passwd",
			content: []byte("malicious"),
			wantErr: true,
		},
		{
			name:    "current directory reference",
			path:    "./file.yaml",
			content: []byte("test content"),
			wantErr: false,
		},
		{
			name:    "multiple slashes",
			path:    "foo//bar.yaml",
			content: []byte("test content"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := NewTempDir()
			if err != nil {
				t.Fatalf("NewTempDir() error = %v", err)
			}
			defer tempDir.Cleanup()

			err = tempDir.WriteFile(tt.path, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected success, verify the file was created with correct content
			if !tt.wantErr {
				fullPath := filepath.Join(tempDir.Path, filepath.Clean(tt.path))
				readContent, err := os.ReadFile(fullPath)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}

				if string(readContent) != string(tt.content) {
					t.Errorf("File content = %q, want %q", string(readContent), string(tt.content))
				}
			}
		})
	}
}

func TestTempDir_ReadFile(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		content    []byte
		setupFile  bool
		wantErr    bool
		errContain string // expected error substring for validation errors
	}{
		{
			name:      "simple filename",
			path:      "file.yaml",
			content:   []byte("test content"),
			setupFile: true,
			wantErr:   false,
		},
		{
			name:      "nested path",
			path:      "patches/deployment.yaml",
			content:   []byte("nested content"),
			setupFile: true,
			wantErr:   false,
		},
		{
			name:      "deeply nested path",
			path:      "overlays/production/patches/deployment.yaml",
			content:   []byte("deeply nested content"),
			setupFile: true,
			wantErr:   false,
		},
		{
			name:       "parent directory traversal",
			path:       "../etc/passwd",
			content:    nil,
			setupFile:  false,
			wantErr:    true,
			errContain: "directory traversal",
		},
		{
			name:       "traversal in middle",
			path:       "foo/../../../etc/passwd",
			content:    nil,
			setupFile:  false,
			wantErr:    true,
			errContain: "directory traversal",
		},
		{
			name:       "absolute path",
			path:       "/etc/passwd",
			content:    nil,
			setupFile:  false,
			wantErr:    true,
			errContain: "absolute paths are not allowed",
		},
		{
			name:      "current directory reference",
			path:      "./file.yaml",
			content:   []byte("test content"),
			setupFile: true,
			wantErr:   false,
		},
		{
			name:      "multiple slashes",
			path:      "foo//bar.yaml",
			content:   []byte("test content"),
			setupFile: true,
			wantErr:   false,
		},
		{
			name:       "nonexistent file",
			path:       "nonexistent.yaml",
			content:    nil,
			setupFile:  false,
			wantErr:    true,
			errContain: "failed to read file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := NewTempDir()
			if err != nil {
				t.Fatalf("NewTempDir() error = %v", err)
			}
			defer tempDir.Cleanup()

			// Setup file if needed
			if tt.setupFile {
				err = tempDir.WriteFile(tt.path, tt.content)
				if err != nil {
					t.Fatalf("WriteFile() setup error = %v", err)
				}
			}

			// Read the file
			content, err := tempDir.ReadFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected success, verify content matches
			if !tt.wantErr {
				if string(content) != string(tt.content) {
					t.Errorf("ReadFile() = %q, want %q", string(content), string(tt.content))
				}
			}
		})
	}
}

func TestTempDir_Cleanup_EmptyPath(t *testing.T) {
	// Test cleanup with empty path (should be a no-op)
	tempDir := &TempDir{Path: ""}
	tempDir.Cleanup() // Should not panic or error
}

func TestTempDir_Cleanup_NonExistentPath(t *testing.T) {
	// Test cleanup with a path that doesn't exist (should print warning but not error)
	tempDir := &TempDir{Path: "/nonexistent/path/that/does/not/exist"}
	tempDir.Cleanup() // Should print warning to stderr but not panic
}

func TestTempDir_WriteFile_ReadOnlyParent(t *testing.T) {
	// Test WriteFile when we can create directories but not write files
	// This is challenging to test portably, but we can try creating a scenario
	// where WriteFile fails but MkdirAll succeeds
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	defer tempDir.Cleanup()

	// Create a directory first
	err = tempDir.WriteFile("subdir/test.yaml", []byte("initial"))
	if err != nil {
		t.Fatalf("Initial WriteFile() error = %v", err)
	}

	// Now make the file read-only
	filePath := filepath.Join(tempDir.Path, "subdir/test.yaml")
	if err := os.Chmod(filePath, 0444); err != nil {
		t.Skipf("Cannot chmod file: %v", err)
	}

	// Try to overwrite the read-only file - this should fail
	err = tempDir.WriteFile("subdir/test.yaml", []byte("overwrite"))
	if err == nil {
		t.Error("WriteFile() should fail when overwriting read-only file")
	}
}

func TestTempDir_Cleanup_RemoveAllError(t *testing.T) {
	// This test attempts to trigger a RemoveAll error
	// We'll create a directory with a read-only file and try to remove it
	tempDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("NewTempDir() error = %v", err)
	}
	// Don't defer cleanup yet

	// Create a subdirectory with a file
	err = tempDir.WriteFile("subdir/file.txt", []byte("content"))
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Make the subdirectory read-only (no write/execute permissions)
	// This should prevent RemoveAll from deleting the contents
	subdirPath := filepath.Join(tempDir.Path, "subdir")
	if err := os.Chmod(subdirPath, 0444); err != nil {
		// If we can't set permissions, skip this test
		tempDir.Cleanup()
		t.Skipf("Cannot chmod directory: %v", err)
	}

	// Attempt cleanup - this should print a warning
	tempDir.Cleanup()

	// Cleanup the permissions so we can actually remove the directory
	if err := os.Chmod(subdirPath, 0755); err != nil {
		t.Logf("Failed to restore permissions: %v", err)
	}
	tempDir.Cleanup()
}

func TestNewTempDir_MkdirTempFailure(t *testing.T) {
	// Test error handling when os.MkdirTemp fails
	// We'll try to set TMPDIR to a read-only directory
	originalTmpDir := os.Getenv("TMPDIR")
	defer func() {
		if originalTmpDir != "" {
			if err := os.Setenv("TMPDIR", originalTmpDir); err != nil {
				t.Logf("Failed to restore TMPDIR: %v", err)
			}
		} else {
			if err := os.Unsetenv("TMPDIR"); err != nil {
				t.Logf("Failed to unset TMPDIR: %v", err)
			}
		}
	}()

	// Create a temporary read-only directory
	readOnlyDir, err := os.MkdirTemp("", "readonly-*")
	if err != nil {
		t.Skipf("Cannot create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(readOnlyDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Make it read-only
	if err := os.Chmod(readOnlyDir, 0444); err != nil {
		t.Skipf("Cannot chmod directory: %v", err)
	}
	defer func() {
		if err := os.Chmod(readOnlyDir, 0755); err != nil {
			t.Logf("Failed to restore permissions: %v", err)
		}
	}()

	// Set TMPDIR to the read-only directory
	if err := os.Setenv("TMPDIR", readOnlyDir); err != nil {
		t.Fatalf("Failed to set TMPDIR: %v", err)
	}

	// This should fail
	_, err = NewTempDir()
	if err == nil {
		t.Error("NewTempDir() should fail when TMPDIR is read-only")
	}
}
