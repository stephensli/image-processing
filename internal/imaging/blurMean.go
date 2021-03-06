package imaging

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/stephensli/image-processing/internal/helpers"
)

type BlurActionMean struct {
	BlurAction
}

// getUpdatedPixel will iterate within the given kernel value range and
// determine the total mean value within the selected kernel size. The returned value being
// the new pixel which could be placed within the center of the kernel of the new image
func (b *BlurActionMean) getUpdatedPixel(x, y int, pixels [][]Pixel) color.RGBA {
	startIdx := helpers.Max(x-int(math.Floor(float64(b.KernelSize)/2))-1, 0)
	endIdx := helpers.Min(startIdx+b.KernelSize-1, len(pixels))

	startYIdx := helpers.Max(y-int(math.Floor(float64(b.KernelSize)/2))-1, 0)
	endYIdx := helpers.Min(startYIdx+b.KernelSize-1, len(pixels[0]))

	kernelInnerSize := 0

	result := Pixel{
		R: 0,
		G: 0,
		B: 0,
		A: 0,
	}

	for i := startIdx; i < endIdx; i++ {
		for j := startYIdx; j < endYIdx; j++ {
			// we increment it here since edges will  not have NxN items,
			//so it's easier to have an adjustable value. Otherwise, on the
			// edges it can seem darker.
			kernelInnerSize += 1

			result.R += pixels[i][j].R
			result.G += pixels[i][j].G
			result.B += pixels[i][j].B
			result.A += pixels[i][j].A
		}
	}

	return color.RGBA{
		R: uint8(math.Min(float64(result.R/kernelInnerSize), 255)),
		G: uint8(math.Min(float64(result.G/kernelInnerSize), 255)),
		B: uint8(math.Min(float64(result.B/kernelInnerSize), 255)),
		A: uint8(math.Min(float64(result.A/kernelInnerSize), 255)),
	}
}

func (b *BlurActionMean) Blur() (image.Image, error) {
	var targetImg *image.RGBA

	pixels, pixelError := b.validateAndGetImagePixels()

	for iter := 0; iter < b.Iterations; iter++ {
		targetImg = image.NewRGBA(b.Image.Bounds())

		if pixelError != nil {
			return nil, pixelError
		}

		var wg sync.WaitGroup

		// iterate over each pixel within the image and determine the new pixel value.
		// Once the new pixel value is determined, update the new target image pixel
		// location.
		for i := 0; i < len(pixels); i++ {
			for j := 0; j < len(pixels[i]); j++ {
				wg.Add(1)

				go func(i, j int, pixels [][]Pixel) {
					newPixel := b.getUpdatedPixel(i, j, pixels)
					targetImg.SetRGBA(j, i, newPixel)

					wg.Done()
				}(i, j, pixels)
			}
		}

		wg.Wait()

		b.Image = targetImg
		pixels, pixelError = b.validateAndGetImagePixels()
	}

	return targetImg, nil
}
