package imgx

import "testing"

func TestNormalizeDecodedFormat(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "webp", in: "webp", want: "WEBP"},
		{name: "jpeg uppercase", in: "JPEG", want: "JPEG"},
		{name: "tif alias", in: "tif", want: "TIFF"},
		{name: "trimmed custom", in: " heic ", want: "HEIC"},
		{name: "empty", in: " ", want: "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeDecodedFormat(tc.in)
			if got != tc.want {
				t.Fatalf("normalizeDecodedFormat(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestMimeFromDecodedFormat(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "webp", in: "webp", want: "image/webp"},
		{name: "jpeg", in: "jpeg", want: "image/jpeg"},
		{name: "tiff alias", in: "tif", want: "image/tiff"},
		{name: "unknown", in: "heic", want: "application/octet-stream"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := mimeFromDecodedFormat(tc.in)
			if got != tc.want {
				t.Fatalf("mimeFromDecodedFormat(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
