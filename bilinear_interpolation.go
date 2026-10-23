package warp

import (
	"image"
	"image/color"
	"math"
)

// bilinearInterpolation samples img at the floating-point coordinate (srcX, srcY)
// using bilinear interpolation. Coordinates outside the image bounds, including
// positive and negative infinities, are clamped to the nearest edge. If either
// coordinate is NaN or img has empty bounds, it returns the zero value in the
// selected color model.
//
// Samples are read through color.Color.RGBA, so their RGB channels are
// alpha-premultiplied 16-bit values. Interpolation therefore occurs in
// premultiplied-alpha space. For the core raster color models, the result uses
// the corresponding color type; unsupported models fall back to color.RGBA.
// No conversion to a linear-light color space is performed.
func bilinearInterpolation(img image.Image, srcX, srcY float64) color.Color {
	imageBounds := img.Bounds()
	colorModel := img.ColorModel()
	if imageBounds.Empty() {
		return colorForModel(colorModel, 0, 0, 0, 0)
	}

	x0, x1, fx, ok := interpolationCoordinate(srcX, imageBounds.Min.X, imageBounds.Max.X-1)
	if !ok {
		return colorForModel(colorModel, 0, 0, 0, 0)
	}
	y0, y1, fy, ok := interpolationCoordinate(srcY, imageBounds.Min.Y, imageBounds.Max.Y-1)
	if !ok {
		return colorForModel(colorModel, 0, 0, 0, 0)
	}

	// Calculate fractional weights.
	fx1, fy1 := 1-fx, 1-fy

	// Fetch the color values from neighboring pixels
	c00R, c00G, c00B, c00A := img.At(x0, y0).RGBA()
	c10R, c10G, c10B, c10A := img.At(x1, y0).RGBA()
	c01R, c01G, c01B, c01A := img.At(x0, y1).RGBA()
	c11R, c11G, c11B, c11A := img.At(x1, y1).RGBA()

	// Compute the weighted sum of the RGB channels
	r := fx1*fy1*float64(c00R) +
		fx*fy1*float64(c10R) +
		fx1*fy*float64(c01R) +
		fx*fy*float64(c11R)

	g := fx1*fy1*float64(c00G) +
		fx*fy1*float64(c10G) +
		fx1*fy*float64(c01G) +
		fx*fy*float64(c11G)

	b := fx1*fy1*float64(c00B) +
		fx*fy1*float64(c10B) +
		fx1*fy*float64(c01B) +
		fx*fy*float64(c11B)

	a := fx1*fy1*float64(c00A) +
		fx*fy1*float64(c10A) +
		fx1*fy*float64(c01A) +
		fx*fy*float64(c11A)

	return colorForModel(colorModel, r, g, b, a)
}

func colorForModel(model color.Model, r, g, b, a float64) color.Color {
	switch model {
	case color.RGBAModel:
		return color.RGBA{
			R: toUint8FromRGBA16(r),
			G: toUint8FromRGBA16(g),
			B: toUint8FromRGBA16(b),
			A: toUint8FromRGBA16(a),
		}
	case color.RGBA64Model:
		return color.RGBA64{
			R: toUint16(r),
			G: toUint16(g),
			B: toUint16(b),
			A: toUint16(a),
		}
	case color.NRGBAModel:
		return color.NRGBA{
			R: toUint8FromRGBA16(unpremultiply(r, a)),
			G: toUint8FromRGBA16(unpremultiply(g, a)),
			B: toUint8FromRGBA16(unpremultiply(b, a)),
			A: toUint8FromRGBA16(a),
		}
	case color.NRGBA64Model:
		return color.NRGBA64{
			R: toUint16(unpremultiply(r, a)),
			G: toUint16(unpremultiply(g, a)),
			B: toUint16(unpremultiply(b, a)),
			A: toUint16(a),
		}
	case color.GrayModel, color.Gray16Model:
		return model.Convert(color.RGBA64{
			R: toUint16(r),
			G: toUint16(g),
			B: toUint16(b),
			A: toUint16(a),
		})
	case color.AlphaModel:
		return color.Alpha{A: toUint8FromRGBA16(a)}
	case color.Alpha16Model:
		return color.Alpha16{A: toUint16(a)}
	default:
		return color.RGBA{
			R: toUint8FromRGBA16(r),
			G: toUint8FromRGBA16(g),
			B: toUint8FromRGBA16(b),
			A: toUint8FromRGBA16(a),
		}
	}
}

func unpremultiply(value, alpha float64) float64 {
	if alpha == 0 {
		return 0
	}
	return value * 65535 / alpha
}

// interpolationCoordinate clamps value to [min, max] before converting it to pixel coordinates.
// It returns the lower and upper neighboring indices, the fractional distance from lower,
// and false when the value is NaN.
// At an image edge, lower and upper are the same index.
func interpolationCoordinate(value float64, min, max int) (lower, upper int, fraction float64, ok bool) {
	if math.IsNaN(value) {
		return 0, 0, 0, false
	}

	// Handle out-of-range values before converting to int. This keeps very
	// large finite values and infinities away from an overflowing conversion.
	if value <= float64(min) {
		return min, min, 0, true
	}
	if value >= float64(max) {
		return max, max, 0, true
	}

	lower = int(math.Floor(value))
	lower = clampInt(lower, min, max)
	upper = lower
	if lower < max {
		upper++
	}
	return lower, upper, value - float64(lower), true
}

func toUint16(value float64) uint16 {
	return uint16(clampFloat(value, 0, 65535) + 0.5)
}

func toUint8FromRGBA16(value float64) uint8 {
	return toUint8(value / 257)
}

func toUint8(value float64) uint8 {
	return uint8(clampFloat(value, 0, 255) + 0.5)
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
