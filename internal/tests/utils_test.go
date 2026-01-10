package tests

import (
	"testing"

	"github.com/Cromemadnd/lazyent"
)

// Unit tests for exported API
func TestNewExtension(t *testing.T) {
	cfg := lazyent.Config{
		ProtoOut: "api/v1",
	}
	ext := lazyent.NewExtension(cfg)
	if ext == nil {
		t.Error("NewExtension returned nil")
	}
}

// Note: Internal logic unit tests are located in internal/gen/utils_test.go
