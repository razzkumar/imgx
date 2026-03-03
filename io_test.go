package imgx

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	errCreate = errors.New("failed to create file")
	errClose  = errors.New("failed to close file")
	errOpen   = errors.New("failed to open file")
)

type badFS struct{}

func (badFS) Create(name string) (io.WriteCloser, error) {
	if name == "badFile.jpg" {
		return badFile{io.Discard}, nil
	}
	return nil, errCreate
}

func (badFS) Open(name string) (io.ReadCloser, error) {
	return nil, errOpen
}

type badFile struct {
	io.Writer
}

func (badFile) Close() error {
	return errClose
}

type quantizer struct {
	palette []color.Color
}

func (q quantizer) Quantize(p color.Palette, m image.Image) color.Palette {
	pal := make([]color.Color, len(p), cap(p))
	copy(pal, p)
	n := cap(p) - len(p)
	if n > len(q.palette) {
		n = len(q.palette)
	}
	for i := 0; i < n; i++ {
		pal = append(pal, q.palette[i])
	}
	return pal
}

func TestOpenSave(t *testing.T) {
	imgWithoutAlpha := image.NewNRGBA(image.Rect(0, 0, 4, 6))
	imgWithoutAlpha.Pix = []uint8{
		0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff,
		0xff, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff,
		0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0x88, 0x88, 0x88, 0xff, 0x88, 0x88, 0x88, 0xff,
		0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0x88, 0x88, 0x88, 0xff, 0x88, 0x88, 0x88, 0xff,
	}
	imgWithAlpha := image.NewNRGBA(image.Rect(0, 0, 4, 6))
	imgWithAlpha.Pix = []uint8{
		0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0x00, 0x00, 0x80, 0xff, 0x00, 0x00, 0x80, 0x00, 0xff, 0x00, 0x80, 0x00, 0xff, 0x00, 0x80,
		0xff, 0x00, 0x00, 0x80, 0xff, 0x00, 0x00, 0x80, 0x00, 0xff, 0x00, 0x80, 0x00, 0xff, 0x00, 0x80,
		0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x88, 0x88, 0x88, 0x00, 0x88, 0x88, 0x88, 0x00,
		0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x88, 0x88, 0x88, 0x00, 0x88, 0x88, 0x88, 0x00,
	}

	options := [][]EncodeOption{
		{
			JPEGQuality(100),
		},
		{
			JPEGQuality(99),
			GIFDrawer(draw.FloydSteinberg),
			GIFNumColors(256),
			GIFQuantizer(quantizer{palette.Plan9}),
			PNGCompressionLevel(png.BestSpeed),
		},
	}

	dir, err := os.MkdirTemp("", "imaging")

	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	for _, ext := range []string{"jpg", "jpeg", "png", "gif", "bmp", "tif", "tiff", "webp"} {
		filename := filepath.Join(dir, "test."+ext)

		img := imgWithoutAlpha
		if ext == "png" {
			img = imgWithAlpha
		}

		for _, opts := range options {
			err := save(img, filename, opts...)
			if err != nil {
				t.Fatalf("failed to save image (%q): %v", filename, err)
			}

			img2, err := open(filename)
			if err != nil {
				t.Fatalf("failed to open image (%q): %v", filename, err)
			}
			got := Clone(img2)

			delta := 0
			if ext == "jpg" || ext == "jpeg" || ext == "gif" {
				delta = 3
			}
			if ext == "webp" {
				delta = 40 // lossy WebP on tiny 4x6 images has significant artifacts
			}

			if !compareNRGBA(got, img, delta) {
				t.Fatalf("bad encode-decode result (ext=%q): got %#v want %#v", ext, got, img)
			}
		}
	}

	buf := &bytes.Buffer{}
	err = Encode(buf, imgWithAlpha, JPEG)
	if err != nil {
		t.Fatalf("failed to encode alpha to JPEG: %v", err)
	}

	buf = &bytes.Buffer{}
	err = Encode(buf, imgWithAlpha, Format(100))
	if err != ErrUnsupportedFormat {
		t.Fatalf("got %v want ErrUnsupportedFormat", err)
	}

	buf = bytes.NewBuffer([]byte("bad data"))
	_, err = Decode(buf)
	if err == nil {
		t.Fatalf("decoding bad data: expected error got nil")
	}

	err = save(imgWithAlpha, filepath.Join(dir, "test.unknown"))
	if err != ErrUnsupportedFormat {
		t.Fatalf("got %v want ErrUnsupportedFormat", err)
	}

	prevFS := fs
	fs = badFS{}
	defer func() { fs = prevFS }()

	err = save(imgWithAlpha, "test.jpg")
	if err != errCreate {
		t.Fatalf("got error %v want errCreate", err)
	}

	err = save(imgWithAlpha, "badFile.jpg")
	if err != errClose {
		t.Fatalf("got error %v want errClose", err)
	}

	_, err = open("test.jpg")
	if err != errOpen {
		t.Fatalf("got error %v want errOpen", err)
	}
}

func TestFormats(t *testing.T) {
	formatNames := map[Format]string{
		JPEG:       "JPEG",
		PNG:        "PNG",
		GIF:        "GIF",
		BMP:        "BMP",
		TIFF:       "TIFF",
		WEBP:       "WEBP",
		Format(-1): "Unknown",
	}
	for format, name := range formatNames {
		got := format.String()
		if got != name {
			t.Fatalf("got format name %q want %q", got, name)
		}
	}
}

func TestFormatFromExtension(t *testing.T) {
	testCases := []struct {
		name string
		ext  string
		want Format
		err  error
	}{
		{
			name: "jpg without leading dot",
			ext:  "jpg",
			want: JPEG,
		},
		{
			name: "jpg with leading dot",
			ext:  ".jpg",
			want: JPEG,
		},
		{
			name: "jpg uppercase",
			ext:  ".JPG",
			want: JPEG,
		},
		{
			name: "webp without leading dot",
			ext:  "webp",
			want: WEBP,
		},
		{
			name: "webp with leading dot",
			ext:  ".webp",
			want: WEBP,
		},
		{
			name: "unsupported",
			ext:  ".unsupportedextension",
			want: -1,
			err:  ErrUnsupportedFormat,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FormatFromExtension(tc.ext)
			if err != tc.err {
				t.Errorf("got error %#v want %#v", err, tc.err)
			}
			if got != tc.want {
				t.Errorf("got result %#v want %#v", got, tc.want)
			}
		})
	}
}

func TestReadOrientation(t *testing.T) {
	t.Skip("orientation test images removed from testdata")
	testCases := []struct {
		path   string
		orient orientation
	}{
		{"docs/images/orientation_0.jpg", 0},
		{"docs/images/orientation_1.jpg", 1},
		{"docs/images/orientation_2.jpg", 2},
		{"docs/images/orientation_3.jpg", 3},
		{"docs/images/orientation_4.jpg", 4},
		{"docs/images/orientation_5.jpg", 5},
		{"docs/images/orientation_6.jpg", 6},
		{"docs/images/orientation_7.jpg", 7},
		{"docs/images/orientation_8.jpg", 8},
	}
	for _, tc := range testCases {
		f, err := os.Open(tc.path)
		if err != nil {
			t.Fatalf("%q: failed to open: %v", tc.path, err)
		}
		orient := readOrientation(f)
		if orient != tc.orient {
			t.Fatalf("%q: got orientation %d want %d", tc.path, orient, tc.orient)
		}
	}
}

func TestReadOrientationFails(t *testing.T) {
	testCases := []struct {
		name string
		data string
	}{
		{
			"empty",
			"",
		},
		{
			"missing SOI marker",
			"\xff\xe1",
		},
		{
			"missing APP1 marker",
			"\xff\xd8",
		},
		{
			"short read marker",
			"\xff\xd8\xff",
		},
		{
			"short read block size",
			"\xff\xd8\xff\xe1\x00",
		},
		{
			"invalid marker",
			"\xff\xd8\x00\xe1\x00\x00",
		},
		{
			"block size too small",
			"\xff\xd8\xff\xe0\x00\x01",
		},
		{
			"short read block",
			"\xff\xd8\xff\xe0\x00\x08\x00",
		},
		{
			"missing EXIF header",
			"\xff\xd8\xff\xe1\x00\xff",
		},
		{
			"invalid EXIF header",
			"\xff\xd8\xff\xe1\x00\xff\x00\x00\x00\x00",
		},
		{
			"missing EXIF header tail",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66",
		},
		{
			"missing byte order tag",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00",
		},
		{
			"invalid byte order tag",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x00\x00",
		},
		{
			"missing byte order tail",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x49\x49",
		},
		{
			"missing exif offset",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x49\x49\x00\x2a",
		},
		{
			"invalid exif offset",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x07",
		},
		{
			"read exif offset error",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x09",
		},
		{
			"missing number of tags",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08",
		},
		{
			"zero number of tags",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08\x00\x00",
		},
		{
			"missing tag",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08\x00\x01",
		},
		{
			"missing tag offset",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08\x00\x01\x00\x00",
		},
		{
			"missing orientation tag",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00",
		},
		{
			"missing orientation tag value offset",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08\x00\x01\x01\x12",
		},
		{
			"missing orientation value",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08\x00\x01\x01\x12\x00\x03\x00\x00\x00\x01",
		},
		{
			"invalid orientation value",
			"\xff\xd8\xff\xe1\x00\xff\x45\x78\x69\x66\x00\x00\x4d\x4d\x00\x2a\x00\x00\x00\x08\x00\x01\x01\x12\x00\x03\x00\x00\x00\x01\x00\x09",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if o := readOrientation(strings.NewReader(tc.data)); o != orientationUnspecified {
				t.Fatalf("got orientation %d want %d", o, orientationUnspecified)
			}
		})
	}
}

func TestAutoOrientation(t *testing.T) {
	t.Skip("orientation test images removed from testdata")
	toBW := func(img image.Image) []byte {
		b := img.Bounds()
		data := make([]byte, 0, b.Dx()*b.Dy())
		for x := b.Min.X; x < b.Max.X; x++ {
			for y := b.Min.Y; y < b.Max.Y; y++ {
				c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
				if c.Y < 128 {
					data = append(data, 1)
				} else {
					data = append(data, 0)
				}
			}
		}
		return data
	}

	f, err := os.Open("docs/images/orientation_0.jpg")
	if err != nil {
		t.Fatalf("os.Open(%q): %v", "docs/images/orientation_0.jpg", err)
	}
	orig, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("image.Decode(%q): %v", "docs/images/orientation_0.jpg", err)
	}
	origBW := toBW(orig)

	testCases := []struct {
		path string
	}{
		{"docs/images/orientation_0.jpg"},
		{"docs/images/orientation_1.jpg"},
		{"docs/images/orientation_2.jpg"},
		{"docs/images/orientation_3.jpg"},
		{"docs/images/orientation_4.jpg"},
		{"docs/images/orientation_5.jpg"},
		{"docs/images/orientation_6.jpg"},
		{"docs/images/orientation_7.jpg"},
		{"docs/images/orientation_8.jpg"},
	}
	for _, tc := range testCases {
		img, err := open(tc.path, AutoOrientation(true))
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds() != orig.Bounds() {
			t.Fatalf("%s: got bounds %v want %v", tc.path, img.Bounds(), orig.Bounds())
		}
		imgBW := toBW(img)
		if !bytes.Equal(imgBW, origBW) {
			t.Fatalf("%s: got bw data %v want %v", tc.path, imgBW, origBW)
		}
	}

	if _, err := Decode(strings.NewReader("invalid data"), AutoOrientation(true)); err == nil {
		t.Fatal("expected error got nil")
	}
}

func TestDefaultJPEGQuality(t *testing.T) {
	if DefaultJPEGQuality != 95 {
		t.Errorf("DefaultJPEGQuality = %d, want 95", DefaultJPEGQuality)
	}
	if defaultEncodeConfig.jpegQuality != DefaultJPEGQuality {
		t.Errorf("defaultEncodeConfig.jpegQuality = %d, want %d", defaultEncodeConfig.jpegQuality, DefaultJPEGQuality)
	}
}

func TestEncodeOptionClamping(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	testCases := []struct {
		name   string
		option EncodeOption
		check  func(cfg encodeConfig) (got, want int)
	}{
		{
			name:   "JPEGQuality(0) clamps to 1",
			option: JPEGQuality(0),
			check:  func(cfg encodeConfig) (int, int) { return cfg.jpegQuality, 1 },
		},
		{
			name:   "JPEGQuality(200) clamps to 100",
			option: JPEGQuality(200),
			check:  func(cfg encodeConfig) (int, int) { return cfg.jpegQuality, 100 },
		},
		{
			name:   "JPEGQuality(50) passes through",
			option: JPEGQuality(50),
			check:  func(cfg encodeConfig) (int, int) { return cfg.jpegQuality, 50 },
		},
		{
			name:   "WebPQuality(-5) clamps to 0",
			option: WebPQuality(-5),
			check:  func(cfg encodeConfig) (int, int) { return cfg.webpQuality, 0 },
		},
		{
			name:   "WebPQuality(150) clamps to 100",
			option: WebPQuality(150),
			check:  func(cfg encodeConfig) (int, int) { return cfg.webpQuality, 100 },
		},
		{
			name:   "WebPQuality(50) passes through",
			option: WebPQuality(50),
			check:  func(cfg encodeConfig) (int, int) { return cfg.webpQuality, 50 },
		},
		{
			name:   "GIFNumColors(0) clamps to 1",
			option: GIFNumColors(0),
			check:  func(cfg encodeConfig) (int, int) { return cfg.gifNumColors, 1 },
		},
		{
			name:   "GIFNumColors(500) clamps to 256",
			option: GIFNumColors(500),
			check:  func(cfg encodeConfig) (int, int) { return cfg.gifNumColors, 256 },
		},
		{
			name:   "GIFNumColors(128) passes through",
			option: GIFNumColors(128),
			check:  func(cfg encodeConfig) (int, int) { return cfg.gifNumColors, 128 },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := defaultEncodeConfig
			tc.option(&cfg)
			got, want := tc.check(cfg)
			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}

	_ = img
}

func TestSaveCleanupOnError(t *testing.T) {
	t.Run("unsupported extension does not leave file", func(t *testing.T) {
		dir := t.TempDir()
		img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

		target := filepath.Join(dir, "test.xyz")
		err := save(img, target)
		if err == nil {
			t.Fatal("expected error for unsupported format, got nil")
		}

		if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
			t.Errorf("expected file %q to not exist after failed save, but os.Stat returned: %v", target, statErr)
		}
	})

	t.Run("non-existent directory returns error", func(t *testing.T) {
		img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

		target := filepath.Join(t.TempDir(), "nonexistent", "subdir", "test.png")
		err := save(img, target)
		if err == nil {
			t.Fatal("expected error saving to non-existent directory, got nil")
		}
	})
}
