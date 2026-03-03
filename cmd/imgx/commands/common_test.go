package commands

import (
	"testing"

	"github.com/razzkumar/imgx"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    imgx.Format
		wantErr bool
	}{
		{"jpg", imgx.JPEG, false},
		{"jpeg", imgx.JPEG, false},
		{"png", imgx.PNG, false},
		{"gif", imgx.GIF, false},
		{"tif", imgx.TIFF, false},
		{"tiff", imgx.TIFF, false},
		{"bmp", imgx.BMP, false},
		{"webp", imgx.WEBP, false},
		{"WEBP", imgx.WEBP, false},
		{"unknown", imgx.JPEG, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatName(t *testing.T) {
	tests := []struct {
		format imgx.Format
		want   string
	}{
		{imgx.JPEG, "JPEG"},
		{imgx.PNG, "PNG"},
		{imgx.WEBP, "WEBP"},
		{imgx.Format(-1), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatName(tt.format)
			if got != tt.want {
				t.Errorf("FormatName(%v) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestParseFilter(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantErr  bool
	}{
		{"lanczos", imgx.Lanczos.Name, false},
		{"nearest", imgx.NearestNeighbor.Name, false},
		{"box", imgx.Box.Name, false},
		{"linear", imgx.Linear.Name, false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFilter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilter(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.wantName {
				t.Errorf("ParseFilter(%q).Name = %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}

func TestParseAnchor(t *testing.T) {
	tests := []struct {
		input   string
		want    imgx.Anchor
		wantErr bool
	}{
		{"center", imgx.Center, false},
		{"topleft", imgx.TopLeft, false},
		{"top-left", imgx.TopLeft, false},
		{"bottomright", imgx.BottomRight, false},
		{"unknown", imgx.Center, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAnchor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAnchor(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseAnchor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatAspectRatioCLI(t *testing.T) {
	if got := FormatAspectRatio(1920, 1080); got != "16:9" {
		t.Errorf("FormatAspectRatio(1920, 1080) = %q, want %q", got, "16:9")
	}
	if got := FormatAspectRatio(0, 100); got != "N/A" {
		t.Errorf("FormatAspectRatio(0, 100) = %q, want %q", got, "N/A")
	}
}

func TestChangeExtension(t *testing.T) {
	tests := []struct {
		path   string
		format imgx.Format
		want   string
	}{
		{"photo.jpg", imgx.WEBP, "photo.webp"},
		{"photo.jpg", imgx.PNG, "photo.png"},
		{"photo.png", imgx.JPEG, "photo.jpg"},
		{"photo.jpg", imgx.GIF, "photo.gif"},
		{"photo.jpg", imgx.TIFF, "photo.tiff"},
		{"photo.jpg", imgx.BMP, "photo.bmp"},
		{"photo.jpg", imgx.Format(-1), "photo.jpg"},
		{"path/to/photo.jpg", imgx.WEBP, "path/to/photo.webp"},
	}

	for _, tt := range tests {
		t.Run(tt.path+"->"+tt.want, func(t *testing.T) {
			got := changeExtension(tt.path, tt.format)
			if got != tt.want {
				t.Errorf("changeExtension(%q, %v) = %q, want %q", tt.path, tt.format, got, tt.want)
			}
		})
	}
}
