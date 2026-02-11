package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogInfo(t *testing.T) {
	// Helper to capture stderr (logInfo writes to stderr)
	captureOutput := func(f func()) string {
		old := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		f()

		w.Close()
		os.Stderr = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name   string
		quiet  bool
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple message",
			quiet:  false,
			format: "Test message",
			args:   []interface{}{},
			want:   "Test message\n",
		},
		{
			name:   "formatted message",
			quiet:  false,
			format: "Processing %d items",
			args:   []interface{}{5},
			want:   "Processing 5 items\n",
		},
		{
			name:   "multiple args",
			quiet:  false,
			format: "%s: %d/%d",
			args:   []interface{}{"Progress", 10, 100},
			want:   "Progress: 10/100\n",
		},
		{
			name:   "quiet mode suppresses output",
			quiet:  true,
			format: "Should not appear",
			args:   []interface{}{},
			want:   "", // No output in quiet mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set quiet mode
			oldQuiet := quiet
			quiet = tt.quiet
			defer func() { quiet = oldQuiet }()

			output := captureOutput(func() {
				logInfo(tt.format, tt.args...)
			})
			assert.Equal(t, tt.want, output)
		})
	}
}

func TestLogVerbose(t *testing.T) {
	// Helper to capture stderr
	captureOutput := func(f func()) string {
		old := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		f()

		w.Close()
		os.Stderr = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name    string
		verbose bool
		quiet   bool
		format  string
		args    []interface{}
		want    string
	}{
		{
			name:    "verbose mode enabled",
			verbose: true,
			quiet:   false,
			format:  "Debug: %s",
			args:    []interface{}{"test"},
			want:    "[VERBOSE] Debug: test\n",
		},
		{
			name:    "verbose mode disabled",
			verbose: false,
			quiet:   false,
			format:  "Debug: %s",
			args:    []interface{}{"test"},
			want:    "", // No output when verbose is false
		},
		{
			name:    "quiet mode suppresses verbose",
			verbose: true,
			quiet:   true,
			format:  "Debug: %s",
			args:    []interface{}{"test"},
			want:    "", // No output when quiet is true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set verbose and quiet modes
			oldVerbose := verbose
			oldQuiet := quiet
			verbose = tt.verbose
			quiet = tt.quiet
			defer func() {
				verbose = oldVerbose
				quiet = oldQuiet
			}()

			output := captureOutput(func() {
				logVerbose(tt.format, tt.args...)
			})
			assert.Equal(t, tt.want, output)
		})
	}
}

func TestLogError(t *testing.T) {
	// Helper to capture stderr
	captureOutput := func(f func()) string {
		old := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		f()

		w.Close()
		os.Stderr = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple error",
			format: "Error occurred",
			args:   []interface{}{},
			want:   "Error: Error occurred\n",
		},
		{
			name:   "formatted error",
			format: "Failed to process %s",
			args:   []interface{}{"file.txt"},
			want:   "Error: Failed to process file.txt\n",
		},
		{
			name:   "error with multiple args",
			format: "Error at line %d: %s",
			args:   []interface{}{42, "syntax error"},
			want:   "Error: Error at line 42: syntax error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				logError(tt.format, tt.args...)
			})
			assert.Equal(t, tt.want, output)
		})
	}
}
