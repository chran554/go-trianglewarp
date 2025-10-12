func bilinearInterpolation(img image.Image, srcX, srcY float64) color.Color {
    x0, y0 := int(math.Floor(srcX)), int(math.Floor(srcY))
    x1, y1 := x0+1, y0+1

    // Ensure coordinates stay within image bounds
    bounds := img.Bounds()
    x0 = clamp(x0, bounds.Min.X, bounds.Max.X-1)
    y0 = clamp(y0, bounds.Min.Y, bounds.Max.Y-1)
    x1 = clamp(x1, bounds.Min.X, bounds.Max.X-1)
    y1 = clamp(y1, bounds.Min.Y, bounds.Max.Y-1)

    // Calculate fractional weights
    fx := srcX - float64(x0)
    fy := srcY - float64(y0)
    fx1, fy1 := 1-fx, 1-fy

    // Fetch the color values from neighboring pixels
    c00 := color.RGBAModel.Convert(img.At(x0, y0)).(color.RGBA)
    c10 := color.RGBAModel.Convert(img.At(x1, y0)).(color.RGBA)
    c01 := color.RGBAModel.Convert(img.At(x0, y1)).(color.RGBA)
    c11 := color.RGBAModel.Convert(img.At(x1, y1)).(color.RGBA)

    // Compute the weighted sum of the RGB channels
    r := fx1*fy1*float64(c00.R) +
        fx*fy1*float64(c10.R) +
        fx1*fy*float64(c01.R) +
        fx*fy*float64(c11.R)

    g := fx1*fy1*float64(c00.G) +
        fx*fy1*float64(c10.G) +
        fx1*fy*float64(c01.G) +
        fx*fy*float64(c11.G)

    b := fx1*fy1*float64(c00.B) +
        fx*fy1*float64(c10.B) +
        fx1*fy*float64(c01.B) +
        fx*fy*float64(c11.B)

    a := fx1*fy1*float64(c00.A) +
        fx*fy1*float64(c10.A) +
        fx1*fy*float64(c01.A) +
        fx*fy*float64(c11.A)

    // Return the interpolated color
    return color.RGBA{
        R: uint8(clampFloat(r, 0, 255)),
        G: uint8(clampFloat(g, 0, 255)),
        B: uint8(clampFloat(b, 0, 255)),
        A: uint8(clampFloat(a, 0, 255)),
    }
}

func clamp(value, min, max int) int {
    if value < min {
        return min
    }
    if value > max {
        return max
    }
    return value
}

func clampFloat(value, min, max float64) float64 {
    if value < min {
        return min
    }
    if value > max {
        return max
    }
    return value
}