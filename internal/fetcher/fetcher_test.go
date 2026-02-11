package fetcher

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchFromFile(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T) string // Returns file path
		cleanup   func(string)
		expectErr bool
		checkData func(*testing.T, []byte)
	}{
		{
			name: "reads valid file",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "test.json")
				content := []byte(`{"test": "data"}`)
				err := os.WriteFile(tmpFile, content, 0644)
				require.NoError(t, err)
				return tmpFile
			},
			cleanup: func(path string) {},
			expectErr: false,
			checkData: func(t *testing.T, data []byte) {
				assert.Equal(t, `{"test": "data"}`, string(data))
			},
		},
		{
			name: "reads empty file",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "empty.json")
				err := os.WriteFile(tmpFile, []byte{}, 0644)
				require.NoError(t, err)
				return tmpFile
			},
			cleanup: func(path string) {},
			expectErr: false,
			checkData: func(t *testing.T, data []byte) {
				assert.Empty(t, data)
			},
		},
		{
			name: "reads large file",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "large.json")
				// Create a 1KB file
				content := make([]byte, 1024)
				for i := range content {
					content[i] = 'a'
				}
				err := os.WriteFile(tmpFile, content, 0644)
				require.NoError(t, err)
				return tmpFile
			},
			cleanup: func(path string) {},
			expectErr: false,
			checkData: func(t *testing.T, data []byte) {
				assert.Len(t, data, 1024)
			},
		},
		{
			name: "fails on non-existent file",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.json")
			},
			cleanup: func(path string) {},
			expectErr: true,
			checkData: nil,
		},
		{
			name: "fails on directory path",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			cleanup: func(path string) {},
			expectErr: true,
			checkData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			defer tt.cleanup(path)

			data, err := FetchFromFile(path)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				if tt.checkData != nil {
					tt.checkData(t, data)
				}
			}
		})
	}
}

func TestFetchFromURL(t *testing.T) {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		expectErr bool
		checkData func(*testing.T, []byte)
	}{
		{
			name: "fetches valid JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Verify User-Agent is set
				assert.Contains(t, r.Header.Get("User-Agent"), "delinea-netconfig")

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"test": "data"}`))
			},
			expectErr: false,
			checkData: func(t *testing.T, data []byte) {
				assert.Equal(t, `{"test": "data"}`, string(data))
			},
		},
		{
			name: "fetches empty response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte{})
			},
			expectErr: false,
			checkData: func(t *testing.T, data []byte) {
				assert.Empty(t, data)
			},
		},
		{
			name: "handles 404 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
			},
			expectErr: true,
			checkData: nil,
		},
		{
			name: "handles 500 error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectErr: true,
			checkData: nil,
		},
		{
			name: "handles 403 forbidden",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Forbidden"))
			},
			expectErr: true,
			checkData: nil,
		},
		{
			name: "fetches large response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Write 10KB of data
				data := make([]byte, 10*1024)
				for i := range data {
					data[i] = 'x'
				}
				w.Write(data)
			},
			expectErr: false,
			checkData: func(t *testing.T, data []byte) {
				assert.Len(t, data, 10*1024)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			// Fetch from server
			data, err := FetchFromURL(server.URL)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkData != nil {
					tt.checkData(t, data)
				}
			}
		})
	}
}

func TestFetchFromURL_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "invalid URL scheme",
			url:  "not-a-valid-url",
		},
		{
			name: "unreachable host",
			url:  "http://localhost:99999/nonexistent",
		},
		{
			name: "invalid port",
			url:  "http://example.com:99999999/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FetchFromURL(tt.url)
			assert.Error(t, err)
		})
	}
}
