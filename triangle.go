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
	putPixelFn := func(x, y int) {
		image.Set(x, y, porterDuffOverComposition(image.At(x, y), lineColor))
	}

	if lineColor != nil {
		bresenham.Bresenham(t.P1.X, t.P1.Y, t.P2.X, t.P2.Y, putPixelFn)
		bresenham.Bresenham(t.P2.X, t.P2.Y, t.P3.X, t.P3.Y, putPixelFn)
		bresenham.Bresenham(t.P3.X, t.P3.Y, t.P1.X, t.P1.Y, putPixelFn)
	}

	if pointColor != nil {
		paintPoint(image, t.P1.X, t.P1.Y, pointColor)
		paintPoint(image, t.P2.X, t.P2.Y, pointColor)
		paintPoint(image, t.P3.X, t.P3.Y, pointColor)
	}
}

// porterDuffOverComposition is the composed color as the resulting color when you paint with srcColor over dstColor using composition "over".
func porterDuffOverComposition(dstColor color.Color, srcColor *color.Color) color.RGBA {
	dr, dg, db, da := dstColor.RGBA()
	sr, sg, sb, sa := (*srcColor).RGBA()

	r := sr + ((0xffff-sa)*dr)/0xffff
	g := sg + ((0xffff-sa)*dg)/0xffff
	b := sb + ((0xffff-sa)*db)/0xffff
	a := sa + ((0xffff-sa)*da)/0xffff

	cr := uint8(r >> 8)
	cg := uint8(g >> 8)
	cb := uint8(b >> 8)
	ca := uint8(a >> 8)
	c := color.RGBA{R: cr, G: cg, B: cb, A: ca}
	return c
}

func paintPoint(image draw.Image, x, y int, pointColor *color.Color) {
	putPixelFn := func(x, y int) {
		image.Set(x, y, porterDuffOverComposition(image.At(x, y), pointColor))
	}

	putPixelFn(x-1, y)
	putPixelFn(x+1, y)
	putPixelFn(x, y-1)
	putPixelFn(x, y+1)
}
