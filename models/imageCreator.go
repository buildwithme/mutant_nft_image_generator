package models

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

var cacheCounter int
var cacheImage = make(map[string]int)
var imageIndex = make(map[int]image.Image)

type TraitSavedConf struct {
	Value string `json:"value"`
	TraitType string `json:"trait_type"`
	Path string `json:"path"`
}

var TraitSaved = make(map[int][]TraitSavedConf)
var traitSavedCount  =1

func NewImageCreator() *ImageCreator {
	return &ImageCreator{}
}

type ImageCreator struct {
	Paths  []string
	images []int
	final  *image.RGBA
}

func (c *ImageCreator) Add(trait Trait, choosedType *SingleTrait) {
	imagePath := choosedType.BasePath

	TraitSaved[traitSavedCount] = append(TraitSaved[traitSavedCount], TraitSavedConf{
		Value: choosedType.Name,
		TraitType: trait.Name,
		Path: choosedType.BasePath,
	})

	c.Paths = append(c.Paths, imagePath)

	if indexImage, ok := cacheImage[imagePath]; ok {
		c.images = append(c.images, indexImage)
		return
	}
	sourceImage := getImage(imagePath)
	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	sourceImageFinal := resize.Resize(1096, 1096, sourceImage, resize.Lanczos3)

	cacheImage[imagePath] = cacheCounter
	imageIndex[cacheCounter] = sourceImageFinal
	c.images = append(c.images, cacheCounter)
	cacheCounter++

}

func (c *ImageCreator) Process() *image.RGBA {
	for i, imageSourceIndex := range c.images {
		imageSource := imageIndex[imageSourceIndex]
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
	traitSavedCount++
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
