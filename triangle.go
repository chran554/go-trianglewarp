package warp

import "image"

type Triangle struct {
	P1, P2, P3 *image.Point
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
