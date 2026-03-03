package imgx

import "testing"

func TestGCD(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"12 and 8", 12, 8, 4},
		{"100 and 75", 100, 75, 25},
		{"7 and 13", 7, 13, 1},
		{"0 and 5", 0, 5, 5},
		{"5 and 0", 5, 0, 5},
		{"1 and 1", 1, 1, 1},
		{"1920 and 1080", 1920, 1080, 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GCD(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("GCD(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestFormatAspectRatio(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   string
	}{
		{"1920x1080", 1920, 1080, "16:9"},
		{"800x600", 800, 600, "4:3"},
		{"100x100", 100, 100, "1:1"},
		{"0 width", 0, 100, "N/A"},
		{"0 height", 100, 0, "N/A"},
		{"0x0", 0, 0, "N/A"},
		{"3840x2160", 3840, 2160, "16:9"},
		{"1080x1920", 1080, 1920, "9:16"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAspectRatio(tt.width, tt.height)
			if got != tt.want {
				t.Errorf("FormatAspectRatio(%d, %d) = %q, want %q", tt.width, tt.height, got, tt.want)
			}
		})
	}
}
