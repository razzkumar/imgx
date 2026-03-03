package imgx

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestImageMethodChain(t *testing.T) {
	// Create a 100x80 test image
	img := NewImage(100, 80, color.NRGBA{R: 128, G: 64, B: 32, A: 255})

	// Chain: Resize → Blur → AdjustBrightness → Grayscale
	result := img.
		Resize(50, 40, Lanczos).
		Blur(1.0).
		AdjustBrightness(10).
		Grayscale()

	// Verify dimensions
	bounds := result.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 40 {
		t.Errorf("expected 50x40, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify metadata operations list
	ops := result.GetMetadata().Operations
	if len(ops) != 4 {
		t.Fatalf("expected 4 operations, got %d", len(ops))
	}

	expectedActions := []string{"resize", "blur", "adjustBrightness", "grayscale"}
	for i, expected := range expectedActions {
		if ops[i].Action != expected {
			t.Errorf("operation %d: expected action %q, got %q", i, expected, ops[i].Action)
		}
	}
}

func TestImageSaveRoundTrip(t *testing.T) {
	// Create a test image
	img := NewImage(60, 40, color.NRGBA{R: 200, G: 100, B: 50, A: 255})

	// Transform it
	result := img.Resize(30, 20, Lanczos)

	// Save to temp file
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_roundtrip.png")
	if err := result.Save(path, WithoutMetadata()); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("saved file is empty")
	}

	// Load it back
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify dimensions preserved
	bounds := loaded.Bounds()
	if bounds.Dx() != 30 || bounds.Dy() != 20 {
		t.Errorf("round-trip dimensions: expected 30x20, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestImageCloneIndependence(t *testing.T) {
	// Create original with some operations
	original := NewImage(100, 100, color.White)
	original = original.Resize(50, 50, Lanczos)

	originalOpsCount := len(original.GetMetadata().Operations)

	// Clone the metadata and add operation to clone
	clonedMeta := original.GetMetadata().Clone()
	clonedMeta.AddOperation("test_op", "test_params")

	// Verify original is unchanged
	if len(original.GetMetadata().Operations) != originalOpsCount {
		t.Errorf("original operations changed: expected %d, got %d",
			originalOpsCount, len(original.GetMetadata().Operations))
	}

	// Verify clone has the extra operation
	if len(clonedMeta.Operations) != originalOpsCount+1 {
		t.Errorf("clone operations: expected %d, got %d",
			originalOpsCount+1, len(clonedMeta.Operations))
	}
}

func TestImageCloneDetectionResultIndependence(t *testing.T) {
	meta := &ProcessingMetadata{
		Software: "imgx",
		DetectionResult: map[string]any{
			"labels": []any{"cat", "dog"},
		},
	}

	cloned := meta.Clone()

	// Modify the clone's detection result
	if m, ok := cloned.DetectionResult.(map[string]any); ok {
		m["labels"] = []any{"bird"}
	}

	// Verify original is unchanged
	if origMap, ok := meta.DetectionResult.(map[string]any); ok {
		if labels, ok := origMap["labels"].([]any); ok {
			if len(labels) != 2 {
				t.Errorf("original detection result modified: expected 2 labels, got %d", len(labels))
			}
		}
	}
}

func TestImageMetadataTracking(t *testing.T) {
	img := NewImage(100, 100, color.White)

	testCases := []struct {
		name       string
		transform  func(*Image) *Image
		wantAction string
	}{
		{
			name:       "resize",
			transform:  func(i *Image) *Image { return i.Resize(50, 50, Lanczos) },
			wantAction: "resize",
		},
		{
			name:       "blur",
			transform:  func(i *Image) *Image { return i.Blur(1.0) },
			wantAction: "blur",
		},
		{
			name:       "sharpen",
			transform:  func(i *Image) *Image { return i.Sharpen(1.0) },
			wantAction: "sharpen",
		},
		{
			name:       "grayscale",
			transform:  func(i *Image) *Image { return i.Grayscale() },
			wantAction: "grayscale",
		},
		{
			name:       "brightness",
			transform:  func(i *Image) *Image { return i.AdjustBrightness(10) },
			wantAction: "adjustBrightness",
		},
		{
			name:       "contrast",
			transform:  func(i *Image) *Image { return i.AdjustContrast(10) },
			wantAction: "adjustContrast",
		},
		{
			name:       "fit",
			transform:  func(i *Image) *Image { return i.Fit(50, 50, Lanczos) },
			wantAction: "fit",
		},
		{
			name:       "fill",
			transform:  func(i *Image) *Image { return i.Fill(50, 50, Center, Lanczos) },
			wantAction: "fill",
		},
		{
			name:       "thumbnail",
			transform:  func(i *Image) *Image { return i.Thumbnail(50, 50, Lanczos) },
			wantAction: "thumbnail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.transform(img)
			ops := result.GetMetadata().Operations
			if len(ops) != 1 {
				t.Fatalf("expected 1 operation, got %d", len(ops))
			}
			if ops[0].Action != tc.wantAction {
				t.Errorf("expected action %q, got %q", tc.wantAction, ops[0].Action)
			}
			if ops[0].Parameters == "" {
				t.Error("expected non-empty parameters")
			}
			if ops[0].Timestamp.IsZero() {
				t.Error("expected non-zero timestamp")
			}
		})
	}
}

func TestImageMethodChainPreservesImage(t *testing.T) {
	// Verify the original image is not mutated by method chains
	src := &image.NRGBA{
		Rect:   image.Rect(0, 0, 4, 4),
		Stride: 4 * 4,
		Pix:    make([]uint8, 4*4*4),
	}
	img := FromImage(src)

	// Chain operations
	_ = img.Resize(2, 2, Box).Blur(0.5)

	// Original should still be 4x4
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Errorf("original mutated: expected 4x4, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Original should have no operations
	if len(img.GetMetadata().Operations) != 0 {
		t.Errorf("original metadata mutated: expected 0 operations, got %d", len(img.GetMetadata().Operations))
	}
}
