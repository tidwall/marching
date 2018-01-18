package marching

import (
	"image"
)

// TerrariumValues ...
func TerrariumValues(img *image.RGBA) []float64 {
	width, height := img.Bounds().Size().X, img.Bounds().Size().Y
	values := make([]float64, width*height)
	for i, j := 0, 0; i < len(img.Pix); i, j = i+4, j+1 {
		values[j] = (float64(img.Pix[i+0])*256 + float64(img.Pix[i+1]) +
			float64(img.Pix[i+2])/256) - 32768
	}
	return values
}
