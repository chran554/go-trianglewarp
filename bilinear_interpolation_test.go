package warp

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"
)

func TestBilinearInterpolation(t *testing.T) {
	// Create a simple 3x3 image with distinct colors for testing
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))              //             0    1    2
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})       // Black     0 K----R----G
	img.Set(1, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})     // Red         |    |    |
	img.Set(2, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})     // Green     1 B----Y----C
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})     // Blue        |    |    |
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 0, A: 255})   // Yellow    2 M----Gr---W
	img.Set(2, 1, color.RGBA{R: 0, G: 255, B: 255, A: 255})   // Cyan
	img.Set(0, 2, color.RGBA{R: 255, G: 0, B: 255, A: 255})   // Magenta
	img.Set(1, 2, color.RGBA{R: 128, G: 128, B: 128, A: 255}) // Gray
	img.Set(2, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // White

	tests := []struct {
		name         string
		srcX, srcY   float64
		expectedRGBA color.RGBA
	}{
		{
			name:         "Exact integer pixel - Top Left (0, 0)",
			srcX:         0.0,
			srcY:         0.0,
			expectedRGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name:         "Exact integer pixel - Center (1, 1)",
			srcX:         1.0,
			srcY:         1.0,
			expectedRGBA: color.RGBA{R: 255, G: 255, B: 0, A: 255},
		},
		{
			name:         "Fractional coordinates - Interpolation between (0,0), (1,0), (0,1), (1,1)",
			srcX:         0.5,
			srcY:         0.5,
			expectedRGBA: color.RGBA{R: 128, G: 64, B: 64, A: 255}, // Weighted average
		},
		{
			name:         "Fractional coordinates - Interpolation between (1,0), (2,0), (1,1), (2,1)",
			srcX:         1.5,
			srcY:         0.5,
			expectedRGBA: color.RGBA{R: 128, G: 191, B: 64, A: 255}, // Weighted average
		},
		{
			name:         "Fractional coordinates - Interpolation between (0,1), (1,1), (0,2), (1,2)",
			srcX:         0.5,
			srcY:         1.5,
			expectedRGBA: color.RGBA{R: 160, G: 96, B: 160, A: 255}, // Weighted average
		},
		{
			name:         "Fractional coordinates - Interpolation between (1,1), (2,1), (1,2), (2,2)",
			srcX:         1.5,
			srcY:         1.5,
			expectedRGBA: color.RGBA{R: 160, G: 223, B: 160, A: 255}, // Weighted average
		},
		{
			name:         "Boundary pixel - Bottom Right (2, 2)",
			srcX:         2.0,
			srcY:         2.0,
			expectedRGBA: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name:         "Out of bounds - Clamp Top Left",
			srcX:         -1.0,
			srcY:         -1.0,
			expectedRGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name:         "Out of bounds - Clamp Bottom Right",
			srcX:         3.0,
			srcY:         3.0,
			expectedRGBA: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bilinearInterpolation(img, tt.srcX, tt.srcY).(color.RGBA)
			if result != tt.expectedRGBA {
				t.Errorf("bilinearInterpolation(%v, %v) = %v; want %v", tt.srcX, tt.srcY, result, tt.expectedRGBA)
			}
		})
	}
}

func TestBilinearInterpolationInvalidCoordinates(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(2, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img.Set(2, 1, color.RGBA{R: 0, G: 255, B: 255, A: 255})
	img.Set(0, 2, color.RGBA{R: 255, G: 0, B: 255, A: 255})
	img.Set(2, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	tests := []struct {
		name         string
		srcX, srcY   float64
		expectedRGBA color.RGBA
	}{
		{
			name:         "NaN x returns transparent black",
			srcX:         math.NaN(),
			srcY:         1,
			expectedRGBA: color.RGBA{},
		},
		{
			name:         "NaN y returns transparent black",
			srcX:         1,
			srcY:         math.NaN(),
			expectedRGBA: color.RGBA{},
		},
		{
			name:         "negative infinity clamps to the lower edge",
			srcX:         math.Inf(-1),
			srcY:         1,
			expectedRGBA: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
		{
			name:         "positive infinity clamps to the upper edge",
			srcX:         math.Inf(1),
			srcY:         1,
			expectedRGBA: color.RGBA{R: 0, G: 255, B: 255, A: 255},
		},
		{
			name:         "very negative coordinates clamp to the top left",
			srcX:         -math.MaxFloat64,
			srcY:         -math.MaxFloat64,
			expectedRGBA: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name:         "very positive coordinates clamp to the bottom right",
			srcX:         math.MaxFloat64,
			srcY:         math.MaxFloat64,
			expectedRGBA: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bilinearInterpolation(img, tt.srcX, tt.srcY).(color.RGBA)
			if result != tt.expectedRGBA {
				t.Errorf("bilinearInterpolation(%v, %v) = %v; want %v", tt.srcX, tt.srcY, result, tt.expectedRGBA)
			}
		})
	}

	emptyImage := image.NewRGBA(image.Rectangle{})
	result := bilinearInterpolation(emptyImage, 0, 0).(color.RGBA)
	if result != (color.RGBA{}) {
		t.Errorf("bilinearInterpolation on an empty image = %v; want transparent black", result)
	}
}

func TestBilinearInterpolationPreservesColorType(t *testing.T) {
	bounds := image.Rect(0, 0, 2, 2)
	tests := []struct {
		name         string
		img          image.Image
		expectedType reflect.Type
	}{
		{name: "RGBA", img: image.NewRGBA(bounds), expectedType: reflect.TypeOf(color.RGBA{})},
		{name: "RGBA64", img: image.NewRGBA64(bounds), expectedType: reflect.TypeOf(color.RGBA64{})},
		{name: "NRGBA", img: image.NewNRGBA(bounds), expectedType: reflect.TypeOf(color.NRGBA{})},
		{name: "NRGBA64", img: image.NewNRGBA64(bounds), expectedType: reflect.TypeOf(color.NRGBA64{})},
		{name: "Gray", img: image.NewGray(bounds), expectedType: reflect.TypeOf(color.Gray{})},
		{name: "Gray16", img: image.NewGray16(bounds), expectedType: reflect.TypeOf(color.Gray16{})},
		{name: "Alpha", img: image.NewAlpha(bounds), expectedType: reflect.TypeOf(color.Alpha{})},
		{name: "Alpha16", img: image.NewAlpha16(bounds), expectedType: reflect.TypeOf(color.Alpha16{})},
		{
			name: "Paletted fallback",
			img: image.NewPaletted(bounds, color.Palette{
				color.Black,
				color.White,
			}),
			expectedType: reflect.TypeOf(color.RGBA{}),
		},
		{
			name:         "YCbCr fallback",
			img:          image.NewYCbCr(bounds, image.YCbCrSubsampleRatio444),
			expectedType: reflect.TypeOf(color.RGBA{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bilinearInterpolation(tt.img, 0.5, 0.5)
			if resultType := reflect.TypeOf(result); resultType != tt.expectedType {
				t.Errorf("bilinearInterpolation returned %v; want %v", resultType, tt.expectedType)
			}
		})
	}
}

func TestBilinearInterpolationPreserves16BitValues(t *testing.T) {
	rgba64Image := image.NewRGBA64(image.Rect(0, 0, 1, 1))
	expectedRGBA64 := color.RGBA64{R: 0x1234, G: 0x3456, B: 0x5678, A: 0x789a}
	rgba64Image.SetRGBA64(0, 0, expectedRGBA64)

	result := bilinearInterpolation(rgba64Image, 0, 0)
	if result != expectedRGBA64 {
		t.Errorf("bilinearInterpolation returned %v for RGBA64; want %v", result, expectedRGBA64)
	}

	gray16Image := image.NewGray16(image.Rect(0, 0, 1, 1))
	expectedGray16 := color.Gray16{Y: 0x4321}
	gray16Image.SetGray16(0, 0, expectedGray16)

	result = bilinearInterpolation(gray16Image, 0, 0)
	if result != expectedGray16 {
		t.Errorf("bilinearInterpolation returned %v for Gray16; want %v", result, expectedGray16)
	}
}

func TestBilinearInterpolationInvalidCoordinatesPreserveColorType(t *testing.T) {
	gray := image.NewGray(image.Rect(0, 0, 1, 1))
	result := bilinearInterpolation(gray, math.NaN(), 0)
	if _, ok := result.(color.Gray); !ok {
		t.Errorf("bilinearInterpolation returned %T for a Gray image; want color.Gray", result)
	}

	emptyGray16 := image.NewGray16(image.Rectangle{})
	result = bilinearInterpolation(emptyGray16, 0, 0)
	if _, ok := result.(color.Gray16); !ok {
		t.Errorf("bilinearInterpolation returned %T for an empty Gray16 image; want color.Gray16", result)
	}
}
