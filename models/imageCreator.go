package models

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"image_generator/utils"
	"os"
	"sort"
	"strings"
)

var cacheCounter int
var cacheImage = make(map[string]int)
var imageIndex = make(map[int]image.Image)


var currentHashes = make(map[string]bool)

type TraitSavedConf struct {
	Value string `json:"value"`
	TraitType string `json:"trait_type"`
	Path string `json:"path"`
}

var TraitSaved = make(map[int]map[int]TraitSavedConf)
var traitSavedCount  =1

func NewImageCreator() *ImageCreator {
	img := &ImageCreator{}
	img.ExcludeSingleTraits  = make(map[string][]string)
	img.IncludeSingleTraits  = make(map[string][]string)
	return img
}

type ImageCreator struct {
	Paths  []string
	images []int
	final  *image.RGBA

	IncludeTraits []string
	ExcludeTraits []string
	ExcludeSingleTraits map[string][]string
	IncludeSingleTraits map[string][]string
}

func (c *ImageCreator) Add(trait Trait, choosedType *SingleTrait) {
	imagePath := choosedType.BasePath

	c.Paths = append(c.Paths, imagePath)

	var counter int
	if indexImage, ok := cacheImage[imagePath]; ok {
		counter = indexImage
	} else {
		sourceImage := getImage(imagePath)
		sourceImageFinal := resize.Resize(5100, 5100, sourceImage, resize.Lanczos3)
		cacheImage[imagePath] = cacheCounter
		imageIndex[cacheCounter] = sourceImageFinal
		counter = cacheCounter
		cacheCounter++
	}
	if TraitSaved[traitSavedCount] == nil {
		TraitSaved[traitSavedCount] = make(map[int]TraitSavedConf)
	}

	TraitSaved[traitSavedCount][counter] = TraitSavedConf{
		Value: choosedType.Name,
		TraitType: trait.Name,
		Path: imagePath,
	}
	c.images = append(c.images, counter)
}

func (c *ImageCreator) Process() *image.RGBA {

	for i, imageSourceIndex := range c.images {
		traitConfig := TraitSaved[traitSavedCount]
		if len(c.ExcludeTraits) > 0 && utils.ExistIn(traitConfig[imageSourceIndex].TraitType, c.ExcludeTraits) {
			delete(traitConfig, imageSourceIndex)
			continue
		}
		if len(c.IncludeTraits) > 0 && !utils.ExistIn(traitConfig[imageSourceIndex].TraitType, c.IncludeTraits) {
			delete(traitConfig, imageSourceIndex)
			continue
		}

		if len(c.ExcludeSingleTraits) > 0 && utils.ExistSingleIn(traitConfig[imageSourceIndex].TraitType, traitConfig[imageSourceIndex].Value, c.ExcludeSingleTraits) {
			delete(traitConfig, imageSourceIndex)
			continue
		}
		if len(c.IncludeSingleTraits) > 0 && !utils.ExistSingleIn(traitConfig[imageSourceIndex].TraitType, traitConfig[imageSourceIndex].Value, c.IncludeSingleTraits) {
			delete(traitConfig, imageSourceIndex)
			continue
		}
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

func (c ImageCreator) IsHashValid() bool {

	var trait = TraitSaved[traitSavedCount];
	var paths []string

	var keys []int
	for key := range trait {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	for _, i := range keys {
		paths = append(paths, trait[i].Path)
	}

	fullJointPath:= strings.Join(paths,",")

	hash := fmt.Sprintf("%x",sha256.Sum256([]byte(fullJointPath)))
	fmt.Println(hash)
	if _, ok := currentHashes[hash]; ok {
		delete(TraitSaved, traitSavedCount)
		return false
	}
	currentHashes[hash] = true
	return true
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
