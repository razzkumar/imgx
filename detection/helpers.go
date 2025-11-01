package detection

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
)

// imageToJPEGBytes converts image.NRGBA to JPEG bytes
func imageToJPEGBytes(img *image.NRGBA) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// parseJSON parses JSON bytes into a Go value
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
