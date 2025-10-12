package warp

import (
	"image"
	"image/color"
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
