package warp

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/chran554/go-bresenham"
)

// Mesh is a mesh of triangles that can be used to warp an image.
// The mesh grid is made of (xDivisions by yDivisions) sub-rectangles of equal size.
// The sub-rectangles are ordered in rows and columns, and each sub-rectangle is made of two triangles.
//
// A mesh with the width and height of 3 by 2 divisions for a rectangle with size (200, 100).
// The number of points in the mesh are (3+1) by (2+1) = 12.
// Note that points in the mesh are shared by the triangles connected to it.
// Any change to a point will affect all connected triangles.
//
//	(0,0)
//	  +--+--+--+
//	  | /| /| /|
//	  |/ |/ |/ |
//	  +--+--+--+
//	  | /| /| /|
//	  |/ |/ |/ |
//	  +--+--+--+
//	           (200, 100)
type Mesh struct {
	width  int
	height int
	points []*image.Point
}

// NewMesh creates a Mesh of triangles that can be used to warp an image.
// The mesh grid is made of xDivisions by yDivisions subdivision-rectangles, that all are of equal size.
// The subdivision rectangles are ordered in yDivision rows and xDivision columns.
//
// A mesh with 3 by 2 divisions for a rectangle with size (200, 100).
// Each subdivision rectangle is made of two triangles.
//
//	(0,0)
//	  +--+--+--+
//	  | /| /| /|
//	  |/ |/ |/ |
//	  +--+--+--+
//	  | /| /| /|
//	  |/ |/ |/ |
//	  +--+--+--+
//	           (200, 100)
func NewMesh(r image.Rectangle, xDivisions, yDivisions int) *Mesh {
	if xDivisions < 1 || yDivisions < 1 {
		return nil
	}

	mesh := &Mesh{
		width:  xDivisions,
		height: yDivisions,
		points: make([]*image.Point, (xDivisions+1)*(yDivisions+1)),
	}

	// xDivisionsPlus := xDivisions + 1
	yDivisionsPlus := xDivisions + 1

	divisionWidth := float64(r.Dx()) / float64(xDivisions)
	divisionHeight := float64(r.Dy()) / float64(yDivisions)

	for x := 0; x <= xDivisions; x++ {
		for y := 0; y <= yDivisions; y++ {
			mesh.points[y*yDivisionsPlus+x] = &image.Point{
				X: r.Min.X + int(math.Round(float64(x)*divisionWidth)),
				Y: r.Min.Y + int(math.Round(float64(y)*divisionHeight)),
			}
		}
	}

	return mesh
}

func WarpMesh(srcImage image.Image, srcMesh *Mesh, dstImage draw.Image, dstMesh *Mesh, interpolation Interpolation) {
	if srcMesh == nil || dstMesh == nil {
		return
	}

	srcWidth, srcHeight := srcMesh.SubdivisionCount()
	dstWidth, dstHeight := dstMesh.SubdivisionCount()

	for y := 0; y < min(srcHeight, dstHeight); y++ {
		for x := 0; x < min(srcWidth, dstWidth); x++ {
			st1, st2 := srcMesh.SubdivisionTriangles(x, y)
			dt1, dt2 := dstMesh.SubdivisionTriangles(x, y)

			Warp(srcImage, st1, dstImage, dt1, interpolation)
			Warp(srcImage, st2, dstImage, dt2, interpolation)
		}
	}
}

// SubdivisionTriangles returns the two triangles that make up a subdivision rectangle of the mesh.
// The subdivision rectangle is defined by the x and y coordinates.
//
// The subdivision rectangle is made of two triangles.
//
// Triangle t1 and t2 are the two triangles that make up a subdivision rectangle of the mesh.
// Triangle t1 has points p1, p2, p3 and triangle t2 has points p2, p4, p3.
//
//	p1     p2
//	+-----+
//	| t1 /|
//	|   / |
//	|  /  |
//	| /   |
//	|/ t2 |
//	+-----+
//	p3     p4
func (m *Mesh) SubdivisionTriangles(subdivisionX, subdivisionY int) (t1, t2 Triangle) {
	if subdivisionX < 0 || subdivisionY < 0 || subdivisionX >= m.width || subdivisionY >= m.height {
		return Triangle{}, Triangle{}
	}

	yDivisionsPlus := m.width + 1
	p1 := m.points[subdivisionY*yDivisionsPlus+subdivisionX]
	p2 := m.points[subdivisionY*yDivisionsPlus+subdivisionX+1]
	p3 := m.points[(subdivisionY+1)*yDivisionsPlus+subdivisionX]
	p4 := m.points[(subdivisionY+1)*yDivisionsPlus+subdivisionX+1]

	// Triangle t1 and t2 are the two triangles that make up a subdivision rectangle of the mesh.
	// Triangle t1 has points p1, p2, p3 and triangle t2 has points p2, p4, p3.
	//  p1     p2
	//  +-----+
	//  | t1 /|
	//  |   / |
	//  |  /  |
	//  | /   |
	//  |/ t2 |
	//  +-----+
	//  p3     p4

	triangle1 := Triangle{p1, p2, p3} // top left triangle (cw winding)
	triangle2 := Triangle{p2, p4, p3} // bottom right triangle (cw winding)

	return triangle1, triangle2
}

func (m *Mesh) SubdivisionCount() (xDivisions, yDivisions int) {
	return m.width, m.height
}

func (m *Mesh) XSubdivisionCount() int {
	return m.width
}

func (m *Mesh) YSubdivisionCount() int {
	return m.height
}

func (m *Mesh) PointCount() (xPoints, yPoints int) {
	return m.width + 1, m.height + 1
}

func (m *Mesh) Point(x, y int) *image.Point {
	if x < 0 || y < 0 || x > m.width || y > m.height {
		return nil
	}

	return m.points[y*(m.width+1)+x]
}

func (m *Mesh) PaintMesh(image draw.Image, lineColor *color.Color, pointColor *color.Color) {
	for yDivision := 0; yDivision < m.height; yDivision++ {
		for xDivision := 0; xDivision < m.width; xDivision++ {
			t1, t2 := m.SubdivisionTriangles(xDivision, yDivision)

			paintTriangle(image, t1, lineColor, pointColor)
			paintTriangle(image, t2, lineColor, pointColor)
		}
	}
}

func paintTriangle(image draw.Image, t Triangle, lineColor *color.Color, pointColor *color.Color) {
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
