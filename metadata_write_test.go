package imgx

import (
	"errors"
	"fmt"
	"testing"
)

func TestMetadataWriteWarningError(t *testing.T) {
	inner := errors.New("exiftool not found")
	w := &MetadataWriteWarning{Err: inner}
	want := "metadata write warning: exiftool not found"
	if got := w.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestMetadataWriteWarningUnwrap(t *testing.T) {
	inner := errors.New("some inner error")
	w := &MetadataWriteWarning{Err: inner}
	if got := w.Unwrap(); got != inner {
		t.Errorf("Unwrap() = %v, want %v", got, inner)
	}
}

func TestMetadataWriteWarningAs(t *testing.T) {
	inner := errors.New("write failed")
	warning := &MetadataWriteWarning{Err: inner}
	wrapped := fmt.Errorf("save failed: %w", warning)

	var target *MetadataWriteWarning
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As() could not extract *MetadataWriteWarning from wrapped error chain")
	}
	if target != warning {
		t.Errorf("errors.As() extracted %v, want %v", target, warning)
	}
}

func TestMetadataWriteWarningIs(t *testing.T) {
	inner := errors.New("sentinel error")
	warning := &MetadataWriteWarning{Err: inner}
	wrapped := fmt.Errorf("save failed: %w", warning)

	if !errors.Is(wrapped, inner) {
		t.Error("errors.Is() could not find inner error through unwrap chain")
	}
}
