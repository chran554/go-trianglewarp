package main

import (
	"fmt"

	"github.com/chran554/go-bresenham"
	warp "github.com/chran554/go-trianglewarp"

	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <image file>\n", os.Args[0])
	}

	imageFilename := os.Args[1]

	srcImage, _ := loadImage(imageFilename)
	srcImage2, _ := loadImage("images/src-image-2.png")

	warpTriangle(srcImage)

	// warpMeshedImage(srcImage, true)
	warpMeshedImage2(srcImage, srcImage2, true)
}

func warpMeshedImage2(srcImage draw.Image, srcImage2 draw.Image, showMesh bool) {
	//lineColor := color.Color(color.RGBA{R: 255, A: 64})
	//pointColor := color.Color(color.White)

	srcWidth := srcImage.Bounds().Dx()
	srcHeight := srcImage.Bounds().Dy()

	dstImage := image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))

	xDivisions := 12
	yDivisions := 8

	srcMesh := warp.NewMesh(image.Rect(0, 0, srcWidth, srcHeight), xDivisions, yDivisions)
	dstMesh := warp.NewMesh(image.Rect(0, 0, srcWidth, srcHeight), xDivisions, yDivisions)

	// Make destination mesh somewhat irregular/distorted
	deviation := min(srcWidth/srcMesh.XSubdivisionCount(), srcHeight/srcMesh.YSubdivisionCount()) / 2
	for x := 0; x < dstMesh.XSubdivisionCount(); x++ {
		for y := 0; y < dstMesh.YSubdivisionCount(); y++ {
			t1, t2 := dstMesh.SubdivisionTriangles(x, y)

			t1.P1.X = rnd(t1.P1.X, deviation)
			t1.P1.Y = rnd(t1.P1.Y, deviation)
			t1.P2.X = rnd(t1.P2.X, deviation)
			t1.P2.Y = rnd(t1.P2.Y, deviation)
			t1.P3.X = rnd(t1.P3.X, deviation)
			t1.P3.Y = rnd(t1.P3.Y, deviation)

			t2.P1.X = rnd(t2.P1.X, deviation)
			t2.P1.Y = rnd(t2.P1.Y, deviation)
			t2.P2.X = rnd(t2.P2.X, deviation)
			t2.P2.Y = rnd(t2.P2.Y, deviation)
			t2.P3.X = rnd(t2.P3.X, deviation)
			t2.P3.Y = rnd(t2.P3.Y, deviation)
		}
	}

	//warp.Warp(srcImage, srcTriangle1, dstImage, dstTriangle1, warp.InterpolationBiLinear)
	//warp.Warp(srcImage, srcTriangle2, dstImage, dstTriangle2, warp.InterpolationBiLinear)

	frameCount := 20
	for i := 0; i <= frameCount; i++ {
		progress := float64(i) / float64(frameCount)
		MorphImages(srcImage, srcMesh, srcImage2, dstMesh, dstImage, progress, nil)

		//srcMesh.PaintMesh(srcImage, &lineColor, &pointColor)
		//dstMesh.PaintMesh(dstImage, &lineColor, &pointColor)

		_ = saveImage(dstImage, fmt.Sprintf("output/mesh-warp-dst-%03d.png", i))
	}
}

func MorphImages(srcImage1 draw.Image, srcMesh1 *warp.Mesh, srcImage2 draw.Image, srcMesh2 *warp.Mesh, dstImage *image.RGBA, progress float64, options *warp.Options) {
	for x := 0; x < srcMesh2.XSubdivisionCount(); x++ {
		for y := 0; y < srcMesh2.YSubdivisionCount(); y++ {
			srcTriangle1_1, srcTriangle1_2 := srcMesh1.SubdivisionTriangles(x, y)
			srcTriangle2_1, srcTriangle2_2 := srcMesh2.SubdivisionTriangles(x, y)

			warp.Morph(srcImage1, srcTriangle1_1, srcImage2, srcTriangle2_1, dstImage, progress, options)
			warp.Morph(srcImage1, srcTriangle1_2, srcImage2, srcTriangle2_2, dstImage, progress, options)
		}
	}
}

func rnd(value, maxDeviation int) int {
	return value + rand.N(maxDeviation/2) - (maxDeviation / 4)
}

func warpMeshedImage(srcImage image.Image, showMesh bool) {
	srcWidth := srcImage.Bounds().Dx()
	srcHeight := srcImage.Bounds().Dy()

	dstImage := image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))

	divisions := 8
	divisionsPlus := divisions + 1
	divisionWidth := srcWidth / divisions
	divisionHeight := srcHeight / divisions

	meshPoints := make([]*image.Point, (divisions+1)*(divisions+1))
	warpPoints := make([]*image.Point, (divisions+1)*(divisions+1))

	// Setup points for both source and destination meshes
	for x := 0; x <= divisions; x++ {
		for y := 0; y <= divisions; y++ {
			dx := rand.N(divisionWidth/2) - (divisionWidth / 4)
			dy := rand.N(divisionHeight/2) - (divisionHeight / 4)

			meshPoints[y*divisionsPlus+x] = &image.Point{X: x * divisionWidth, Y: y * divisionHeight}
			warpPoints[y*divisionsPlus+x] = &image.Point{X: x*divisionWidth + dx, Y: y*divisionHeight + dy}
		}
	}

	// Warp the source image to the destination by using the predefined mesh points
	for x := 0; x < divisions; x++ {
		for y := 0; y < divisions; y++ {
			sp1 := meshPoints[y*divisionsPlus+x]
			sp2 := meshPoints[y*divisionsPlus+x+1]
			sp3 := meshPoints[(y+1)*divisionsPlus+x]
			sp4 := meshPoints[(y+1)*divisionsPlus+x+1]

			dp1 := warpPoints[y*divisionsPlus+x]
			dp2 := warpPoints[y*divisionsPlus+x+1]
			dp3 := warpPoints[(y+1)*divisionsPlus+x]
			dp4 := warpPoints[(y+1)*divisionsPlus+x+1]

			srcTriangle1 := warp.Triangle{sp1, sp2, sp3}
			srcTriangle2 := warp.Triangle{sp2, sp3, sp4}

			dstTriangle1 := warp.Triangle{dp1, dp2, dp3}
			dstTriangle2 := warp.Triangle{dp2, dp3, dp4}

			warp.Warp(srcImage, srcTriangle1, dstImage, dstTriangle1, nil)
			warp.Warp(srcImage, srcTriangle2, dstImage, dstTriangle2, nil)

			if showMesh {
				drawTriangle(dstImage, dstTriangle1, color.NRGBA{128, 128, 128, 255})
				drawTriangle(dstImage, dstTriangle2, color.NRGBA{128, 128, 128, 255})

				drawDot(dstImage, dp1, color.White)
				drawDot(dstImage, dp2, color.White)
				drawDot(dstImage, dp3, color.White)
				drawDot(dstImage, dp4, color.White)
			}
		}
	}

	_ = saveImage(dstImage, "output/mesh-warp.png")
}

func warpTriangle(srcImage image.Image) {
	srcWidth := srcImage.Bounds().Dx()
	srcHeight := srcImage.Bounds().Dy()

	dstImage := image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))

	srcTriangle := warp.Triangle{&image.Point{X: 200, Y: 250}, &image.Point{X: 400, Y: 250}, &image.Point{X: 300, Y: 400}}
	dstTriangle := warp.Triangle{&image.Point{X: 350, Y: 450}, &image.Point{X: 500, Y: 350}, &image.Point{X: 460, Y: 500}}
	dstTriangle2 := warp.Triangle{&image.Point{X: 50, Y: 450}, &image.Point{X: 160, Y: 500}, &image.Point{X: 200, Y: 350}}

	copyArea(srcImage, 0, 0, srcWidth-1, srcHeight-1, dstImage)
	warp.Warp(srcImage, srcTriangle, dstImage, dstTriangle, nil)
	warp.Warp(srcImage, srcTriangle, dstImage, dstTriangle2, nil)
	drawTriangle(dstImage, srcTriangle, color.White)
	drawTriangle(dstImage, dstTriangle, color.White)
	drawTriangle(dstImage, dstTriangle2, color.White)
	_ = saveImage(dstImage, "output/triangle-warp.png")
}

// copyArea just copies the src image to the destination image pixel by pixel.
func copyArea(src image.Image, sx0 int, sy0 int, sx1 int, sy1 int, dst draw.Image) {
	for y := sy0; y <= min(dst.Bounds().Dy()-1, sy1-sy0); y++ {
		for x := sx0; x <= min(dst.Bounds().Dx()-1, sx1-sx0); x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
}

func drawTriangle(img *image.RGBA, t warp.Triangle, c color.Color) {
	bresenham.Bresenham(t.P1.X, t.P1.Y, t.P2.X, t.P2.Y, func(x, y int) { img.Set(x, y, c) })
	bresenham.Bresenham(t.P2.X, t.P2.Y, t.P3.X, t.P3.Y, func(x, y int) { img.Set(x, y, c) })
	bresenham.Bresenham(t.P3.X, t.P3.Y, t.P1.X, t.P1.Y, func(x, y int) { img.Set(x, y, c) })
}

func drawDot(img *image.RGBA, p *image.Point, c color.Color) {
	img.Set(p.X, p.Y, c)
	img.Set(p.X+1, p.Y, c)
	img.Set(p.X-1, p.Y, c)
	img.Set(p.X, p.Y+1, c)
	img.Set(p.X, p.Y-1, c)
}

func saveImage(img image.Image, filename string) error {
	f, err := createFileWithDirPath(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		return err
	}
	return nil
}

func loadImage(filename string) (draw.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba, err
	}
	if nrgba, ok := img.(*image.NRGBA); ok {
		return nrgba, err
	}
	return nil, err
}

func createFileWithDirPath(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}
