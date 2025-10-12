package morph

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// morph2 iterates through every pixel in the destination triangle's bounding box
// to see if the pixel is included in the triangle.
//
// The barycentric weight of each triangle point is calculated from scratch for each pixel with no optimization.
func morph2(srcImage image.Image, srcTriangle Triangle, dstImage draw.Image, dstTriangle Triangle) {
	// Calculate the edge function for the whole triangle (ABC)
	ABC := dstTriangle.doubleSignedArea()

	// Our nifty trick: Don't bother drawing the triangle if it's back facing
	if ABC <= 0 {
		return
	}

	invABC := 1.0 / float64(ABC)

	// Get the bounding box of the triangle
	boundingBox := dstTriangle.boundingBox()

	// Loop through all the pixels of the bounding box
	for py := boundingBox.Min.Y; py <= boundingBox.Max.Y; py++ {
		for px := boundingBox.Min.X; px <= boundingBox.Max.X; px++ {

			P := image.Point{X: px, Y: py}

			// Calculate our edge functions
			ABP := Triangle{P1: dstTriangle.P1, P2: dstTriangle.P2, P3: P}.doubleSignedArea()
			BCP := Triangle{P1: dstTriangle.P2, P2: dstTriangle.P3, P3: P}.doubleSignedArea()
			CAP := Triangle{P1: dstTriangle.P3, P2: dstTriangle.P1, P3: P}.doubleSignedArea()

			// If all the edge functions are positive, the point is inside the triangle
			if ABP >= 0 && BCP >= 0 && CAP >= 0 {
				// Normalise the edge functions by dividing by the total area to get the barycentric coordinates
				weightA := float64(BCP) * invABC
				weightB := float64(CAP) * invABC
				weightC := float64(ABP) * invABC

				srcPoint := image.Point{
					X: int(math.Round(float64(srcTriangle.P1.X)*weightA + float64(srcTriangle.P2.X)*weightB + float64(srcTriangle.P3.X)*weightC)),
					Y: int(math.Round(float64(srcTriangle.P1.Y)*weightA + float64(srcTriangle.P2.Y)*weightB + float64(srcTriangle.P3.Y)*weightC)),
				}

				c := srcImage.At(srcPoint.X, srcPoint.Y)
				dstImage.Set(P.X, P.Y, c)
			} else {
				// Highlight/colorize dst triangle bounding box that is not the triangle area
				rgba := dstImage.At(P.X, P.Y).(color.RGBA)
				rgba.R = uint8((int(rgba.R) + 255) / 2)
				rgba.G = uint8((int(rgba.G) + 255) / 2)
				rgba.B = uint8((int(rgba.B) + 255) / 2)
				dstImage.Set(P.X, P.Y, rgba)
			}
		}
	}
}

// morph3 iterates through every pixel in the destination triangle's bounding box
// to see if the pixel is included in the triangle.
//
// The barycentric weight of each triangle point is calculated from
// a precalculated delta increment in both x and y-axis
// for each pixel in the destination triangle bounding box.
//
// The pre-calculations require a destination triangle with a bounding box with dimensions of at least 2x2 pixels.
func morph3(srcImage image.Image, srcTriangle Triangle, dstImage draw.Image, dstTriangle Triangle) {
	// Calculate the edge function for the whole triangle (ABC)
	ABC := dstTriangle.doubleSignedArea()

	// Our nifty trick: Don't bother drawing the triangle if it's back facing
	if ABC <= 0 {
		return
	}

	boundingBox := dstTriangle.boundingBox()

	// Precalculated barycentric values for the triangle
	wa0, wb0, wc0, wadx, wady, wbdx, wbdy, wcdx, wcdy := preCalculatedBarycentricWeightDeltas(dstTriangle)

	// Loop through all the pixels of the bounding box
	for py := boundingBox.Min.Y; py <= boundingBox.Max.Y; py++ {
		for px := boundingBox.Min.X; px <= boundingBox.Max.X; px++ {

			ix := float64(px - boundingBox.Min.X)
			iy := float64(py - boundingBox.Min.Y)

			// Normalise the edge functions by dividing by the total area to get the barycentric coordinates
			weightA := wa0 + ix*wadx + iy*wady
			weightB := wb0 + ix*wbdx + iy*wbdy
			weightC := wc0 + ix*wcdx + iy*wcdy

			ABP := weightA * float64(ABC)
			BCP := weightB * float64(ABC)
			CAP := weightC * float64(ABC)

			// If all the edge functions are positive, the point is inside the triangle
			if ABP >= 0 && BCP >= 0 && CAP >= 0 {
				srcPoint := image.Point{
					X: int(math.Round(float64(srcTriangle.P1.X)*weightA + float64(srcTriangle.P2.X)*weightB + float64(srcTriangle.P3.X)*weightC)),
					Y: int(math.Round(float64(srcTriangle.P1.Y)*weightA + float64(srcTriangle.P2.Y)*weightB + float64(srcTriangle.P3.Y)*weightC)),
				}

				c := srcImage.At(srcPoint.X, srcPoint.Y)
				dstImage.Set(px, py, c)
			} else {
				// Highlight/colorize dst triangle bounding box that is not the triangle area
				rgba := dstImage.At(px, py).(color.RGBA)
				rgba.R = uint8((int(rgba.R) + 255) / 2)
				rgba.G = uint8((int(rgba.G) + 255) / 2)
				rgba.B = uint8((int(rgba.B) + 255) / 2)
				dstImage.Set(px, py, rgba)
			}
		}
	}
}

func morph4(srcImage image.Image, srcTriangle Triangle, dstImage draw.Image, dstTriangle Triangle) {
	// Calculate the edge function for the whole triangle (ABC)
	ABC := dstTriangle.doubleSignedArea()

	// Our nifty trick: Don't bother drawing the triangle if it's back facing
	if ABC <= 0 {
		return
	}

	// If the triangle has no area (not visible)
	if ABC == 0 {
		return
	}

	invABC := 1.0 / float64(ABC)

	minX, maxX, boundingBox := triangleMinMaxX(dstTriangle)

	// Loop through all the pixels of the bounding box
	for py := boundingBox.Min.Y; py <= boundingBox.Max.Y; py++ {
		index := py - boundingBox.Min.Y

		for px := minX[index]; px <= maxX[index]; px++ {
			P := image.Point{X: px, Y: py}

			// Calculate our edge functions
			ABP := Triangle{P1: dstTriangle.P1, P2: dstTriangle.P2, P3: P}.doubleSignedArea()
			BCP := Triangle{P1: dstTriangle.P2, P2: dstTriangle.P3, P3: P}.doubleSignedArea()
			CAP := Triangle{P1: dstTriangle.P3, P2: dstTriangle.P1, P3: P}.doubleSignedArea()

			// Normalise the edge functions by dividing by the total area to get the barycentric coordinates
			weightA := float64(BCP) * invABC
			weightB := float64(CAP) * invABC
			weightC := float64(ABP) * invABC

			// If all the edge functions are positive, the point is inside the triangle
			srcPoint := image.Point{
				X: int(math.Round(float64(srcTriangle.P1.X)*weightA + float64(srcTriangle.P2.X)*weightB + float64(srcTriangle.P3.X)*weightC)),
				Y: int(math.Round(float64(srcTriangle.P1.Y)*weightA + float64(srcTriangle.P2.Y)*weightB + float64(srcTriangle.P3.Y)*weightC)),
			}

			c := srcImage.At(srcPoint.X, srcPoint.Y)
			dstImage.Set(P.X, P.Y, c)
		}
	}
}

func morph5(srcImage image.Image, srcTriangle Triangle, dstImage draw.Image, dstTriangle Triangle) {
	// Calculate the edge function for the whole triangle (ABC)
	ABC := dstTriangle.doubleSignedArea()

	// Our nifty trick: Don't bother drawing the triangle if it's back facing
	if ABC <= 0 {
		return
	}

	// If the triangle has no area (not visible)
	if ABC == 0 {
		return
	}

	minX, maxX, boundingBox := triangleMinMaxX(dstTriangle)

	// Precalculated barycentric values for the triangle
	wa0, wb0, wc0, wadx, wady, wbdx, wbdy, wcdx, wcdy := preCalculatedBarycentricWeightDeltas(dstTriangle)

	// Loop through all the pixels of the bounding box
	for py := boundingBox.Min.Y; py <= boundingBox.Max.Y; py++ {
		index := py - boundingBox.Min.Y

		for px := minX[index]; px <= maxX[index]; px++ {
			ix := float64(px - boundingBox.Min.X)
			iy := float64(py - boundingBox.Min.Y)

			// Normalise the edge functions by dividing by the total area to get the barycentric coordinates
			// Multiplication could be swapped with iterative addition, but that might increase accumulative errors(?).
			// However that may be too small errors to ever visually appear.
			weightA := wa0 + ix*wadx + iy*wady
			weightB := wb0 + ix*wbdx + iy*wbdy
			weightC := wc0 + ix*wcdx + iy*wcdy

			// If all the edge functions are positive, the point is inside the triangle
			srcPoint := image.Point{
				X: int(math.Round(float64(srcTriangle.P1.X)*weightA + float64(srcTriangle.P2.X)*weightB + float64(srcTriangle.P3.X)*weightC)),
				Y: int(math.Round(float64(srcTriangle.P1.Y)*weightA + float64(srcTriangle.P2.Y)*weightB + float64(srcTriangle.P3.Y)*weightC)),
			}

			c := srcImage.At(srcPoint.X, srcPoint.Y)
			dstImage.Set(px, py, c)
		}
	}
}

func morph6(srcImage image.Image, srcTriangle Triangle, dstImage draw.Image, dstTriangle Triangle) {
	// Calculate the edge function for the whole triangle (ABC)
	ABC := dstTriangle.doubleSignedArea()

	// Our nifty trick: Don't bother drawing the triangle if it's back facing
	if ABC <= 0 {
		return
	}

	// If the triangle has no area (not visible)
	if ABC == 0 {
		return
	}

	minX, maxX, boundingBox := triangleMinMaxX(dstTriangle)

	// Precalculated barycentric values for the triangle
	wa0, wb0, wc0, wadx, wady, wbdx, wbdy, wcdx, wcdy := preCalculatedBarycentricWeightDeltas(dstTriangle)

	// Loop through all the pixels of the bounding box
	for py := boundingBox.Min.Y; py <= boundingBox.Max.Y; py++ {
		index := py - boundingBox.Min.Y
		iy := float64(py - boundingBox.Min.Y)

		x0 := float64(minX[index] - boundingBox.Min.X)
		x1 := float64(minX[index] - boundingBox.Min.X + 1)

		waX0Y := wa0 + x0*wadx + iy*wady
		waX1Y := wa0 + x1*wadx + iy*wady

		wbX0Y := wb0 + x0*wbdx + iy*wbdy
		wbX1Y := wb0 + x1*wbdx + iy*wbdy

		wcX0Y := wc0 + x0*wcdx + iy*wcdy
		wcX1Y := wc0 + x1*wcdx + iy*wcdy

		p1XwaXdY := float64(srcTriangle.P1.X) * (waX1Y - waX0Y)
		p1XwaXnY := float64(srcTriangle.P1.X) * waX0Y
		p1YwaXdY := float64(srcTriangle.P1.Y) * (waX1Y - waX0Y)
		p1YwaXnY := float64(srcTriangle.P1.Y) * waX0Y

		p2XwbXdY := float64(srcTriangle.P2.X) * (wbX1Y - wbX0Y)
		p2XwbXnY := float64(srcTriangle.P2.X) * wbX0Y
		p2YwbXdY := float64(srcTriangle.P2.Y) * (wbX1Y - wbX0Y)
		p2YwbXnY := float64(srcTriangle.P2.Y) * wbX0Y

		p3XwcXdY := float64(srcTriangle.P3.X) * (wcX1Y - wcX0Y)
		p3XwcXnY := float64(srcTriangle.P3.X) * wcX0Y
		p3YwcXdY := float64(srcTriangle.P3.Y) * (wcX1Y - wcX0Y)
		p3YwcXnY := float64(srcTriangle.P3.Y) * wcX0Y

		for px := minX[index]; px <= maxX[index]; px++ {
			// If all the edge functions are positive, the point is inside the triangle
			srcPoint := image.Point{
				X: int(math.Round(p1XwaXnY + p2XwbXnY + p3XwcXnY)),
				Y: int(math.Round(p1YwaXnY + p2YwbXnY + p3YwcXnY)),
			}

			p1XwaXnY += p1XwaXdY
			p1YwaXnY += p1YwaXdY

			p2XwbXnY += p2XwbXdY
			p2YwbXnY += p2YwbXdY

			p3XwcXnY += p3XwcXdY
			p3YwcXnY += p3YwcXdY

			c := srcImage.At(srcPoint.X, srcPoint.Y)
			dstImage.Set(px, py, c)
		}
	}
}
