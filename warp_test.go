package warp

import (
	"fmt"
	"image"
	"testing"
)

type triangleTestScenario struct {
	name        string
	srcTriangle Triangle
	dstTriangle Triangle
}

func BenchmarkWarp2(b *testing.B) {
	srcImage, dstImage := testImages()
	testScenarios := triangleTestScenarios()

	for _, testScenario := range testScenarios {
		b.Run(fmt.Sprintf("scenario %s", testScenario.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				warp2(srcImage, testScenario.srcTriangle, dstImage, testScenario.dstTriangle)
			}
		})
	}
}

func BenchmarkWarp3(b *testing.B) {
	srcImage, dstImage := testImages()
	testScenarios := triangleTestScenarios()

	for _, testScenario := range testScenarios {
		b.Run(fmt.Sprintf("scenario %s", testScenario.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				warp3(srcImage, testScenario.srcTriangle, dstImage, testScenario.dstTriangle)
			}
		})
	}
}

func BenchmarkWarp4(b *testing.B) {
	srcImage, dstImage := testImages()
	testScenarios := triangleTestScenarios()

	for _, testScenario := range testScenarios {
		b.Run(fmt.Sprintf("scenario %s", testScenario.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				warp4(srcImage, testScenario.srcTriangle, dstImage, testScenario.dstTriangle)
			}
		})
	}
}

func BenchmarkWarp5(b *testing.B) {
	srcImage, dstImage := testImages()
	testScenarios := triangleTestScenarios()

	for _, testScenario := range testScenarios {
		b.Run(fmt.Sprintf("scenario %s", testScenario.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				warp5(srcImage, testScenario.srcTriangle, dstImage, testScenario.dstTriangle)
			}
		})
	}
}

func BenchmarkWarp6(b *testing.B) {
	srcImage, dstImage := testImages()
	testScenarios := triangleTestScenarios()

	for _, testScenario := range testScenarios {
		b.Run(fmt.Sprintf("scenario %s", testScenario.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				warp6(srcImage, testScenario.srcTriangle, dstImage, testScenario.dstTriangle)
			}
		})
	}
}

func BenchmarkWarp7(b *testing.B) {
	srcImage, dstImage := testImages()
	testScenarios := triangleTestScenarios()

	for _, testScenario := range testScenarios {
		b.Run(fmt.Sprintf("scenario %s", testScenario.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Warp(srcImage, testScenario.srcTriangle, dstImage, testScenario.dstTriangle, &Options{Interpolation: InterpolationNearestNeighbour})
			}
		})
	}
}

func testImages() (srcImage *image.RGBA, dstImage *image.RGBA) {
	imageWidth := 500
	imageHeight := 500

	srcImage = image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	dstImage = image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	return srcImage, dstImage
}

func triangleTestScenarios() []triangleTestScenario {
	srcTriangle1 := Triangle{&image.Point{X: 200, Y: 200}, &image.Point{X: 300, Y: 200}, &image.Point{X: 200, Y: 300}}
	dstTriangle1 := Triangle{&image.Point{X: 50, Y: 50}, &image.Point{X: 450, Y: 75}, &image.Point{X: 75, Y: 425}}

	srcTriangle2 := Triangle{&image.Point{X: 100, Y: 100}, &image.Point{X: 400, Y: 100}, &image.Point{X: 100, Y: 400}}
	dstTriangle2 := Triangle{&image.Point{X: 200, Y: 200}, &image.Point{X: 450, Y: 150}, &image.Point{X: 110, Y: 375}}

	var testScenarios = []triangleTestScenario{
		{name: "triangle 1", srcTriangle: srcTriangle1, dstTriangle: dstTriangle1},
		{name: "triangle 2", srcTriangle: srcTriangle2, dstTriangle: dstTriangle2},
	}

	return testScenarios
}
