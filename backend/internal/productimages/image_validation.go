package productimages

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/webp"
)

// A compressed image can declare dimensions far larger than its byte size.
// Keep decoded pixel buffers within a practical Worker memory budget.
const maxImagePixels = 8_000_000

func validateImageBytes(data []byte, contentType string) (int, int, error) {
	if len(data) == 0 || len(data) > MaxImageBytes {
		return 0, 0, errors.New("image byte length is outside the allowed range")
	}
	var config func(io.Reader) (image.Config, error)
	var decode func(io.Reader) (image.Image, error)
	switch contentType {
	case "image/png":
		config, decode = png.DecodeConfig, png.Decode
	case "image/jpeg":
		config, decode = jpeg.DecodeConfig, jpeg.Decode
	case "image/webp":
		config, decode = webp.DecodeConfig, webp.Decode
	default:
		return 0, 0, errors.New("unsupported image format")
	}
	info, err := config(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	if info.Width <= 0 || info.Height <= 0 || info.Width > maxImagePixels ||
		info.Height > maxImagePixels/info.Width {
		return 0, 0, errors.New("image dimensions exceed the pixel limit")
	}
	decoded, err := decode(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	if decoded.Bounds().Dx() != info.Width || decoded.Bounds().Dy() != info.Height {
		return 0, 0, errors.New("decoded dimensions do not match image header")
	}
	return info.Width, info.Height, nil
}
