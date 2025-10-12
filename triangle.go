package warp

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/chran554/go-bresenham"
)

type Triangle struct {
	P1, P2, P3 *image.Point
}

func (t Triangle) String() string {
	return fmt.Sprintf("(%v, %v, %v)", t.P1, t.P2, t.P3)
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

func (t Triangle) PaintTriangle(image draw.Image, lineColor *color.Color, pointColor *color.Color) {
	if lineColor != nil {
		bresenham.Bresenham(t.P1.X, t.P1.Y, t.P2.X, t.P2.Y, func(x, y int) { image.Set(x, y, *lineColor) })
		bresenham.Bresenham(t.P2.X, t.P2.Y, t.P3.X, t.P3.Y, func(x, y int) { image.Set(x, y, *lineColor) })
		bresenham.Bresenham(t.P3.X, t.P3.Y, t.P1.X, t.P1.Y, func(x, y int) { image.Set(x, y, *lineColor) })
	}

	if pointColor != nil {
		paintPoint(image, t.P1.X, t.P1.Y, pointColor)
		paintPoint(image, t.P2.X, t.P2.Y, pointColor)
		paintPoint(image, t.P3.X, t.P3.Y, pointColor)
	}
}

func paintPoint(image draw.Image, x, y int, pointColor *color.Color) {
	image.Set(x-1, y, *pointColor)
	image.Set(x+1, y, *pointColor)
	image.Set(x, y-1, *pointColor)
	image.Set(x, y+1, *pointColor)
}
