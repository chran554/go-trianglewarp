package morph

import "C"
import (
	"github.com/chran554/go-bresenham"
	"image"
	"image/color"
	"image/draw"
)

type Interpolation int

const (
	InterpolationNearestNeighbour = iota
	InterpolationBiLinear
)

type Triangle struct {
	P1, P2, P3 image.Point
}

// doubleSignedArea is the doubled (signed) area size for the triangle.
// The sign of the area value tells if the points of the triangle are in given clockwise or counter-clockwise order.
//
// The value is the same as the determinant for the edge vectors u and v
// where u=P2-P1 and v=P3-P1.
// This is also often defined as the cross-product of two vectors in 2D. det(u v)=u x v
// https://mathworld.wolfram.com/CrossProduct.html
func (t Triangle) doubleSignedArea() int {
	return (t.P2.X-t.P1.X)*(t.P3.Y-t.P1.Y) - (t.P2.Y-t.P1.Y)*(t.P3.X-t.P1.X)
}

func (t Triangle) orientationClockwise() int {
	return (t.P2.X-t.P1.X)*(t.P3.Y-t.P1.Y) - (t.P2.Y-t.P1.Y)*(t.P3.X-t.P1.X)
}

func (t Triangle) boundingBox() image.Rectangle {
	r := image.Rectangle{}
	r.Min.X = min(min(t.P1.X, t.P2.X), t.P3.X)
	r.Min.Y = min(min(t.P1.Y, t.P2.Y), t.P3.Y)
	r.Max.X = max(max(t.P1.X, t.P2.X), t.P3.X)
	r.Max.Y = max(max(t.P1.Y, t.P2.Y), t.P3.Y)

	return r
}

func Morph(srcImage image.Image, srcTriangle Triangle, dstImage draw.Image, dstTriangle Triangle, interpolation Interpolation) {
	// Calculate the edge function for the whole triangle (ABC)
	ABC := dstTriangle.doubleSignedArea()

	// Do not use the barycentric method if the triangle is back facing / triangle is left winding.
	//if ABC <= 0 {
	//	return
	//}

	// If the triangle has no area (not visible)
	if ABC == 0 {
		return
	}

	minX, maxX, boundingBox := triangleMinMaxX(dstTriangle)

	// Precalculated barycentric values for the triangle
	wa0, wb0, wc0, wadx, wady, wbdx, wbdy, wcdx, wcdy := preCalculatedBarycentricWeightDeltas(dstTriangle)

	p1XwaXdY := float64(srcTriangle.P1.X) * wadx
	p1YwaXdY := float64(srcTriangle.P1.Y) * wadx

	p2XwbXdY := float64(srcTriangle.P2.X) * wbdx
	p2YwbXdY := float64(srcTriangle.P2.Y) * wbdx

	p3XwcXdY := float64(srcTriangle.P3.X) * wcdx
	p3YwcXdY := float64(srcTriangle.P3.Y) * wcdx

	xWeightDelta := p1XwaXdY + p2XwbXdY + p3XwcXdY
	yWeightDelta := p1YwaXdY + p2YwbXdY + p3YwcXdY

	// Loop through all the pixels of the bounding box
	for py := boundingBox.Min.Y; py <= boundingBox.Max.Y; py++ {
		index := py - boundingBox.Min.Y
		iy := float64(py - boundingBox.Min.Y)

		x0 := float64(minX[index] - boundingBox.Min.X)

		waX0Y := wa0 + x0*wadx + iy*wady
		wbX0Y := wb0 + x0*wbdx + iy*wbdy
		wcX0Y := wc0 + x0*wcdx + iy*wcdy

		p1XwaXnY := float64(srcTriangle.P1.X) * waX0Y
		p1YwaXnY := float64(srcTriangle.P1.Y) * waX0Y

		p2XwbXnY := float64(srcTriangle.P2.X) * wbX0Y
		p2YwbXnY := float64(srcTriangle.P2.Y) * wbX0Y

		p3XwcXnY := float64(srcTriangle.P3.X) * wcX0Y
		p3YwcXnY := float64(srcTriangle.P3.Y) * wcX0Y

		xWeight := p1XwaXnY + p2XwbXnY + p3XwcXnY
		yWeight := p1YwaXnY + p2YwbXnY + p3YwcXnY

		for px := minX[index]; px <= maxX[index]; px++ {
			srcX := 0
			srcY := 0
			var c color.Color
			if interpolation == InterpolationBiLinear {
				c = bilinearInterpolation(srcImage, xWeight, yWeight)
			} else {
				// interpolation == InterpolationNearestNeighbour {
				srcX = int(xWeight + 0.5)
				srcY = int(yWeight + 0.5)
				c = srcImage.At(srcX, srcY)
			}

			dstImage.Set(px, py, c)

			xWeight += xWeightDelta
			yWeight += yWeightDelta
		}
	}
}

func triangleMinMaxX(dstTriangle Triangle) (minX []int, maxX []int, boundingBox image.Rectangle) {
	var l1 []image.Point
	var l2 []image.Point
	var l3 []image.Point
	bresenham.Bresenham(dstTriangle.P1.X, dstTriangle.P1.Y, dstTriangle.P2.X, dstTriangle.P2.Y, func(x, y int) { l1 = append(l1, image.Pt(x, y)) })
	bresenham.Bresenham(dstTriangle.P2.X, dstTriangle.P2.Y, dstTriangle.P3.X, dstTriangle.P3.Y, func(x, y int) { l2 = append(l2, image.Pt(x, y)) })
	bresenham.Bresenham(dstTriangle.P3.X, dstTriangle.P3.Y, dstTriangle.P1.X, dstTriangle.P1.Y, func(x, y int) { l3 = append(l3, image.Pt(x, y)) })

	boundingBox = dstTriangle.boundingBox()

	minX = make([]int, boundingBox.Dy()+1)
	maxX = make([]int, boundingBox.Dy()+1)

	for i := 0; i < boundingBox.Dy()+1; i++ {
		minX[i] = 100_000
		maxX[i] = -100_000
	}

	for _, p := range l1 {
		index := p.Y - boundingBox.Min.Y
		minX[index] = min(minX[index], p.X)
		maxX[index] = max(maxX[index], p.X)
	}

	for _, p := range l2 {
		index := p.Y - boundingBox.Min.Y
		minX[index] = min(minX[index], p.X)
		maxX[index] = max(maxX[index], p.X)
	}

	for _, p := range l3 {
		index := p.Y - boundingBox.Min.Y
		minX[index] = min(minX[index], p.X)
		maxX[index] = max(maxX[index], p.X)
	}

	return minX, maxX, boundingBox
}

// The pre-calculations require a destination triangle with a bounding box with dimensions of at least 2x2 pixels.
func preCalculatedBarycentricWeightDeltas(triangle Triangle) (wa0, wb0, wc0, wadx, wady, wbdx, wbdy, wcdx, wcdy float64) {
	// Get the bounding box of the triangle
	boundingBox := triangle.boundingBox()

	ABC := triangle.doubleSignedArea()
	invABC := 1.0 / float64(ABC)

	wa := [][]float64{{0, 0}, {0, 0}}
	wb := [][]float64{{0, 0}, {0, 0}}
	wc := [][]float64{{0, 0}, {0, 0}}

	// TODO Solve problem when bounding box is <2 pixels in either width or height.
	// TODO Write tests to confirm  that the function can handle calculations on triangles with bounding boxes of <2 pixels in either width or height.

	for py := boundingBox.Min.Y; py <= boundingBox.Min.Y+1; py++ {
		for px := boundingBox.Min.X; px <= boundingBox.Min.X+1; px++ {
			P := image.Point{X: px, Y: py}

			// Calculate our edge functions
			ABP := Triangle{P1: triangle.P1, P2: triangle.P2, P3: P}.doubleSignedArea()
			BCP := Triangle{P1: triangle.P2, P2: triangle.P3, P3: P}.doubleSignedArea()
			CAP := Triangle{P1: triangle.P3, P2: triangle.P1, P3: P}.doubleSignedArea()

			// Normalize the edge functions by dividing by the total area to get the barycentric coordinates
			weightA := float64(BCP) * invABC
			weightB := float64(CAP) * invABC
			weightC := float64(ABP) * invABC

			wa[px-boundingBox.Min.X][py-boundingBox.Min.Y] = weightA
			wb[px-boundingBox.Min.X][py-boundingBox.Min.Y] = weightB
			wc[px-boundingBox.Min.X][py-boundingBox.Min.Y] = weightC
		}
	}

	return wa[0][0], wb[0][0], wc[0][0], //          wa0, wb0, wc0
		wa[1][0] - wa[0][0], wa[0][1] - wa[0][0], // wadx, wady
		wb[1][0] - wb[0][0], wb[0][1] - wb[0][0], // wbdx, wbdy
		wc[1][0] - wc[0][0], wc[0][1] - wc[0][0] //  wcdx, wcdy
}
