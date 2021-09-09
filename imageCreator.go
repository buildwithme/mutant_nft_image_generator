package main

import (
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"image_generator/utils"
	"os"
)

func NewImageCreator() *ImageCreator {
	return &ImageCreator{}
}

type ImageCreator struct {
	Paths []string
	images []image.Image
	final  *image.RGBA
}

func (c *ImageCreator) Add(imagePaths ...string) {
	c.Paths = append(c.Paths, imagePaths...)
	for _, imagePath := range imagePaths {
		sourceImage := getImage(imagePath)
		// resize to width 1000 using Lanczos resampling
		// and preserve aspect ratio
		sourceImageFinal := resize.Resize(1096, 1096, sourceImage, resize.Lanczos3)
		c.images = append(c.images, sourceImageFinal)
	}
}

func (c *ImageCreator) Process() *image.RGBA {
	for i, imageSource := range c.images {
		drawType := draw.Src
		if i != 0 {
			drawType = draw.Over
		}
		if c.final == nil {
			c.final = image.NewRGBA(imageSource.Bounds())
		}
		draw.Draw(c.final, imageSource.Bounds(), imageSource, image.Point{}, drawType)
	}
	return c.final
}

func (c ImageCreator) WriteTo(outputPath string) {
	if c.final == nil {
		utils.Fatal(errors.New("Final image is nil"))
	}
	finalImageOutput, err := os.Create(outputPath)
	if err != nil {
		utils.Fatal(errors.New(fmt.Sprintf("failed to create: %s", err)))
	}
	jpeg.Encode(finalImageOutput, c.final, &jpeg.Options{jpeg.DefaultQuality})
	finalImageOutput.Close()
}

func getImage(imagePath string) image.Image {
	imageSource, err := os.Open(imagePath)
	if err != nil {
		utils.Fatalf("failed to open: %s", err)
	}

	imageResult, err := png.Decode(imageSource)
	if err != nil {
		utils.Fatalf("failed to decode: %s", err)
	}

	return imageResult
}

