package warp

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/chran554/go-bresenham"
)

type Interpolation int

const (
	InterpolationBiLinear Interpolation = iota // Default interpolation value
	InterpolationNearestNeighbour
	InterpolationDefault = InterpolationBiLinear // Default interpolation value
)

type Options struct {
	Interpolation      Interpolation
	TriangleSideColor  *color.Color
	TrianglePointColor *color.Color
}

// Warp warps a triangular excerpt from a source image onto a triangular area on a destination image.
// The pixels written onto the destination area are interpolated using specified options like the interpolation method.
func Warp(srcImage image.Image, srcTriangle Triangle, dstImage draw.Image, dstTriangle Triangle, options *Options) {
	if options == nil {
		options = &Options{}
	}

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
			switch options.Interpolation {
			case InterpolationNearestNeighbour:
				// interpolation == InterpolationNearestNeighbour {
				srcX = int(xWeight + 0.5)
				srcY = int(yWeight + 0.5)
				c = srcImage.At(srcX, srcY)
			case InterpolationBiLinear:
				fallthrough
			default:
				c = bilinearInterpolation(srcImage, xWeight, yWeight)
			}

			dstImage.Set(px, py, c)

			xWeight += xWeightDelta
			yWeight += yWeightDelta
		}
	}

	dstTriangle.PaintTriangle(dstImage, options.TriangleSideColor, options.TrianglePointColor)
}

// Morph interpolates a triangular area from a first image (srcImage1) to a triangular area on a second image (srcImage2).
// The result is written to the dstImage image.
// The image content of the triangular area is interpolated from srcImage1 to srcImage2 with the progress parameter [0.0, 1.0].
// The triangular area, or rather the location of the three points of the triangle, are also interpolated from srcTriangle1 to srcTriangle3 with the progress parameter [0.0, 1.0]
func Morph(srcImage1 image.Image, srcTriangle1 Triangle, srcImage2 image.Image, srcTriangle2 Triangle, dstImage draw.Image, progress float64, options *Options) {
	dstTriangle := Triangle{
		P1: interpolatePoint(srcTriangle1.P1, srcTriangle2.P1, progress),
		P2: interpolatePoint(srcTriangle1.P2, srcTriangle2.P2, progress),
		P3: interpolatePoint(srcTriangle1.P3, srcTriangle2.P3, progress),
	}

	MorphWithFixedDestination(srcImage1, srcTriangle1, srcImage2, srcTriangle2, dstImage, dstTriangle, progress, options)
}

// MorphWithFixedDestination interpolates a triangular area from a first image (srcImage1) to a triangular area on a second image (srcImage2).
// The result is written to the dstTriangle on the dstImage image.
// The image content of the triangular area is interpolated from srcImage1 to srcImage2 with the progress parameter [0.0, 1.0].
func MorphWithFixedDestination(srcImage1 image.Image, srcTriangle1 Triangle, srcImage2 image.Image, srcTriangle2 Triangle, dstImage draw.Image, dstTriangle Triangle, progress float64, options *Options) {
	if options == nil {
		options = &Options{}
	}

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

	// Triangle #1 calculations
	t1p1XwaXdY := float64(srcTriangle1.P1.X) * wadx
	t1p1YwaXdY := float64(srcTriangle1.P1.Y) * wadx

	t1p2XwbXdY := float64(srcTriangle1.P2.X) * wbdx
	t1p2YwbXdY := float64(srcTriangle1.P2.Y) * wbdx

	t1p3XwcXdY := float64(srcTriangle1.P3.X) * wcdx
	t1p3YwcXdY := float64(srcTriangle1.P3.Y) * wcdx

	t1xWeightDelta := t1p1XwaXdY + t1p2XwbXdY + t1p3XwcXdY
	t1yWeightDelta := t1p1YwaXdY + t1p2YwbXdY + t1p3YwcXdY

	// Triangle #2 calculations
	t2p1XwaXdY := float64(srcTriangle2.P1.X) * wadx
	t2p1YwaXdY := float64(srcTriangle2.P1.Y) * wadx

	t2p2XwbXdY := float64(srcTriangle2.P2.X) * wbdx
	t2p2YwbXdY := float64(srcTriangle2.P2.Y) * wbdx

	t2p3XwcXdY := float64(srcTriangle2.P3.X) * wcdx
	t2p3YwcXdY := float64(srcTriangle2.P3.Y) * wcdx

	t2xWeightDelta := t2p1XwaXdY + t2p2XwbXdY + t2p3XwcXdY
	t2yWeightDelta := t2p1YwaXdY + t2p2YwbXdY + t2p3YwcXdY

	// Loop through all the pixels of the bounding box
	for py := boundingBox.Min.Y; py <= boundingBox.Max.Y; py++ {
		index := py - boundingBox.Min.Y
		iy := float64(py - boundingBox.Min.Y)

		x0 := float64(minX[index] - boundingBox.Min.X)

		waX0Y := wa0 + x0*wadx + iy*wady
		wbX0Y := wb0 + x0*wbdx + iy*wbdy
		wcX0Y := wc0 + x0*wcdx + iy*wcdy

		// Triangle #1 calculations
		t1p1XwaXnY := float64(srcTriangle1.P1.X) * waX0Y
		t1p1YwaXnY := float64(srcTriangle1.P1.Y) * waX0Y

		t1p2XwbXnY := float64(srcTriangle1.P2.X) * wbX0Y
		t1p2YwbXnY := float64(srcTriangle1.P2.Y) * wbX0Y

		t1p3XwcXnY := float64(srcTriangle1.P3.X) * wcX0Y
		t1p3YwcXnY := float64(srcTriangle1.P3.Y) * wcX0Y

		t1xWeight := t1p1XwaXnY + t1p2XwbXnY + t1p3XwcXnY
		t1yWeight := t1p1YwaXnY + t1p2YwbXnY + t1p3YwcXnY

		// Triangle #2 calculations
		t2p1XwaXnY := float64(srcTriangle2.P1.X) * waX0Y
		t2p1YwaXnY := float64(srcTriangle2.P1.Y) * waX0Y

		t2p2XwbXnY := float64(srcTriangle2.P2.X) * wbX0Y
		t2p2YwbXnY := float64(srcTriangle2.P2.Y) * wbX0Y

		t2p3XwcXnY := float64(srcTriangle2.P3.X) * wcX0Y
		t2p3YwcXnY := float64(srcTriangle2.P3.Y) * wcX0Y

		t2xWeight := t2p1XwaXnY + t2p2XwbXnY + t2p3XwcXnY
		t2yWeight := t2p1YwaXnY + t2p2YwbXnY + t2p3YwcXnY

		for px := minX[index]; px <= maxX[index]; px++ {
			// Triangle #1 color
			var t1c color.Color
			switch options.Interpolation {
			case InterpolationNearestNeighbour:
				// interpolation == InterpolationNearestNeighbour {
				srcX := int(t1xWeight + 0.5)
				srcY := int(t1yWeight + 0.5)
				t1c = srcImage1.At(srcX, srcY)
			case InterpolationBiLinear:
				fallthrough
			default:
				t1c = bilinearInterpolation(srcImage1, t1xWeight, t1yWeight)
			}

			// Triangle #2 color
			var t2c color.Color
			switch options.Interpolation {
			case InterpolationNearestNeighbour:
				// interpolation == InterpolationNearestNeighbour {
				srcX := int(t2xWeight + 0.5)
				srcY := int(t2yWeight + 0.5)
				t2c = srcImage2.At(srcX, srcY)
			case InterpolationBiLinear:
				fallthrough
			default:
				t2c = bilinearInterpolation(srcImage2, t2xWeight, t2yWeight)
			}

			c := interpolateColor(t1c, t2c, progress)

			dstImage.Set(px, py, c)

			// Triangle #1 calculations
			t1xWeight += t1xWeightDelta
			t1yWeight += t1yWeightDelta

			// Triangle #2 calculations
			t2xWeight += t2xWeightDelta
			t2yWeight += t2yWeightDelta
		}
	}

	dstTriangle.PaintTriangle(dstImage, options.TriangleSideColor, options.TrianglePointColor)
}

func interpolatePoint(p1, p2 *image.Point, progress float64) *image.Point {
	return &image.Point{
		X: int(float64(p1.X)*(1-progress) + float64(p2.X)*progress),
		Y: int(float64(p1.Y)*(1-progress) + float64(p2.Y)*progress),
	}
}

func interpolateColor(c1 color.Color, c2 color.Color, progress float64) color.Color {
	nc1 := color.NRGBAModel.Convert(c1).(color.NRGBA)
	nc2 := color.NRGBAModel.Convert(c2).(color.NRGBA)
	return color.NRGBA{
		R: toUint8(math.Round(float64(nc1.R)*(1-progress) + float64(nc2.R)*progress)),
		G: toUint8(math.Round(float64(nc1.G)*(1-progress) + float64(nc2.G)*progress)),
		B: toUint8(math.Round(float64(nc1.B)*(1-progress) + float64(nc2.B)*progress)),
		A: toUint8(math.Round(float64(nc1.A)*(1-progress) + float64(nc2.A)*progress)),
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
			P := &image.Point{X: px, Y: py}

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
