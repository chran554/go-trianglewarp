package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math/rand/v2"
	morph "morphimage"
	"os"
	"path/filepath"

	"github.com/chran554/go-bresenham"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <image file>\n", os.Args[0])
	}

	imageFilename := os.Args[1]

	srcImage, _ := loadImage(imageFilename)

	morphTriangle(srcImage)

	morphMeshedImage(srcImage, true)
}

func morphMeshedImage(srcImage image.Image, showMesh bool) {
	srcWidth := srcImage.Bounds().Dx()
	srcHeight := srcImage.Bounds().Dy()

	dstImage := image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))

	divisions := 8
	divisionsPlus := divisions + 1
	divisionWidth := srcWidth / divisions
	divisionHeight := srcHeight / divisions

	meshPoints := make([]image.Point, (divisions+1)*(divisions+1))
	morphPoints := make([]image.Point, (divisions+1)*(divisions+1))

	// Setup points for both source and destination meshes
	for x := 0; x <= divisions; x++ {
		for y := 0; y <= divisions; y++ {
			dx := rand.N(divisionWidth/2) - (divisionWidth / 4)
			dy := rand.N(divisionHeight/2) - (divisionHeight / 4)

			meshPoints[y*divisionsPlus+x] = image.Point{X: x * divisionWidth, Y: y * divisionHeight}
			morphPoints[y*divisionsPlus+x] = image.Point{X: x*divisionWidth + dx, Y: y*divisionHeight + dy}
		}
	}

	// Morph the source image to the destination by using the predefined mesh points
	for x := 0; x < divisions; x++ {
		for y := 0; y < divisions; y++ {
			sp1 := meshPoints[y*divisionsPlus+x]
			sp2 := meshPoints[y*divisionsPlus+x+1]
			sp3 := meshPoints[(y+1)*divisionsPlus+x]
			sp4 := meshPoints[(y+1)*divisionsPlus+x+1]

			dp1 := morphPoints[y*divisionsPlus+x]
			dp2 := morphPoints[y*divisionsPlus+x+1]
			dp3 := morphPoints[(y+1)*divisionsPlus+x]
			dp4 := morphPoints[(y+1)*divisionsPlus+x+1]

			srcTriangle1 := morph.Triangle{sp1, sp2, sp3}
			srcTriangle2 := morph.Triangle{sp2, sp3, sp4}

			dstTriangle1 := morph.Triangle{dp1, dp2, dp3}
			dstTriangle2 := morph.Triangle{dp2, dp3, dp4}

			morph.Morph(srcImage, srcTriangle1, dstImage, dstTriangle1, morph.InterpolationBiLinear)
			morph.Morph(srcImage, srcTriangle2, dstImage, dstTriangle2, morph.InterpolationBiLinear)

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

	_ = saveImage(dstImage, "output/mesh-morph.png")
}

func morphTriangle(srcImage image.Image) {
	srcWidth := srcImage.Bounds().Dx()
	srcHeight := srcImage.Bounds().Dy()

	dstImage := image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))

	srcTriangle := morph.Triangle{image.Point{X: 200, Y: 250}, image.Point{X: 400, Y: 250}, image.Point{X: 300, Y: 400}}
	dstTriangle := morph.Triangle{image.Point{X: 350, Y: 450}, image.Point{X: 500, Y: 350}, image.Point{X: 460, Y: 500}}
	dstTriangle2 := morph.Triangle{image.Point{X: 50, Y: 450}, image.Point{X: 160, Y: 500}, image.Point{X: 200, Y: 350}}

	copyArea(srcImage, 0, 0, srcWidth-1, srcHeight-1, dstImage)
	morph.Morph(srcImage, srcTriangle, dstImage, dstTriangle, morph.InterpolationBiLinear)
	morph.Morph(srcImage, srcTriangle, dstImage, dstTriangle2, morph.InterpolationBiLinear)
	drawTriangle(dstImage, srcTriangle, color.White)
	drawTriangle(dstImage, dstTriangle, color.White)
	drawTriangle(dstImage, dstTriangle2, color.White)
	_ = saveImage(dstImage, "output/triangle-morph.png")
}

// copyArea just copies the src image to the destination image pixel by pixel.
func copyArea(src image.Image, sx0 int, sy0 int, sx1 int, sy1 int, dst draw.Image) {
	for y := sy0; y <= min(dst.Bounds().Dy()-1, sy1-sy0); y++ {
		for x := sx0; x <= min(dst.Bounds().Dx()-1, sx1-sx0); x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
}

func drawTriangle(img *image.RGBA, t morph.Triangle, c color.Color) {
	bresenham.Bresenham(t.P1.X, t.P1.Y, t.P2.X, t.P2.Y, func(x, y int) { img.Set(x, y, c) })
	bresenham.Bresenham(t.P2.X, t.P2.Y, t.P3.X, t.P3.Y, func(x, y int) { img.Set(x, y, c) })
	bresenham.Bresenham(t.P3.X, t.P3.Y, t.P1.X, t.P1.Y, func(x, y int) { img.Set(x, y, c) })
}

func drawDot(img *image.RGBA, p image.Point, c color.Color) {
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

func loadImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func createFileWithDirPath(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}
