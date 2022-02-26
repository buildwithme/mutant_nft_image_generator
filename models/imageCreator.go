package models

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"image_generator/utils"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

type Cache struct {
	mu           sync.Mutex
	cacheImage   map[string]int
	imageIndex   map[int]image.Image
	cacheCounter int
}

func NewCache() *Cache {
	return &Cache{
		cacheImage: make(map[string]int),
		imageIndex: make(map[int]image.Image),
	}
}

var cache *Cache = NewCache()

type Hashes struct {
	mu            sync.Mutex
	CurrentHashes map[string]bool
}

var SavedHashes *Hashes = &Hashes{
	CurrentHashes: make(map[string]bool),
}

type TraitSavedConf struct {
	Value     string `json:"value"`
	TraitType string `json:"trait_type"`
	Path      string `json:"path"`
}

// SafeCounter is safe to use concurrently.
type TraitSaved struct {
	mu              sync.Mutex
	Data            map[int]map[int]TraitSavedConf
	TraitSavedCount int
}

var SavedTraits *TraitSaved = NewSavedTraits()

func NewSavedTraits() *TraitSaved {
	var d = TraitSaved{}
	d.Data = make(map[int]map[int]TraitSavedConf)
	d.TraitSavedCount = 1
	return &d
}
func (s *TraitSaved) Add() {

}
func (s *TraitSaved) Exists(id int) bool {
	return s.Data[id] != nil
}
func (s *TraitSaved) Lock() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("==================err")
			fmt.Println(err)
		}
	}()
	s.mu.Lock()
}
func (s *TraitSaved) Unlock() {
	s.mu.Unlock()
}

func NewImageCreator(id int) *ImageCreator {
	img := &ImageCreator{}
	img.id = id
	img.ExcludeSingleTraits = make(map[string][]string)
	img.IncludeSingleTraits = make(map[string][]string)
	return img
}

type ImageCreator struct {
	id     int
	Paths  []string
	images []int
	final  *image.RGBA

	IncludeTraits       []string
	ExcludeTraits       []string
	ExcludeSingleTraits map[string][]string
	IncludeSingleTraits map[string][]string
}

func (c *ImageCreator) Add(trait Trait, choosedType *SingleTrait) {
	imagePath := choosedType.BasePath

	c.Paths = append(c.Paths, imagePath)

	var counter int
	cache.mu.Lock()
	if indexImage, ok := cache.cacheImage[imagePath]; ok {
		counter = indexImage
	} else {
		sourceImage := getImage(imagePath)
		sourceImageFinal := resize.Resize(2048, 2048, sourceImage, resize.Lanczos3)
		cache.cacheImage[imagePath] = cache.cacheCounter
		cache.imageIndex[cache.cacheCounter] = sourceImageFinal
		counter = cache.cacheCounter
		cache.cacheCounter++
	}
	cache.mu.Unlock()

	SavedTraits.mu.Lock()
	if !SavedTraits.Exists(c.id) {
		SavedTraits.Data[c.id] = make(map[int]TraitSavedConf)
	}

	SavedTraits.Data[c.id][counter] = TraitSavedConf{
		Value:     choosedType.Name,
		TraitType: trait.Name,
		Path:      imagePath,
	}
	SavedTraits.mu.Unlock()
	c.images = append(c.images, counter)
}

func (c *ImageCreator) Process() *image.RGBA {

	for i, imageSourceIndex := range c.images {
		SavedTraits.mu.Lock()
		traitConfig := SavedTraits.Data[c.id]
		SavedTraits.mu.Unlock()
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
		cache.mu.Lock()
		imageSource := cache.imageIndex[imageSourceIndex]
		cache.mu.Unlock()
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
	SavedTraits.mu.Lock()
	var trait = SavedTraits.Data[c.id]
	SavedTraits.mu.Unlock()
	var paths []string

	var keys []int
	for key := range trait {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	for _, i := range keys {
		paths = append(paths, trait[i].Path)
	}

	fullJointPath := strings.Join(paths, ",")

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(fullJointPath)))
	fmt.Println(hash)

	SavedHashes.mu.Lock()
	if _, ok := SavedHashes.CurrentHashes[hash]; ok {
		SavedHashes.mu.Unlock()
		SavedTraits.mu.Lock()
		delete(SavedTraits.Data, c.id)
		SavedTraits.mu.Unlock()
		return false
	} else {
		SavedHashes.CurrentHashes[hash] = true
		SavedHashes.mu.Unlock()
	}

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
	SavedTraits.mu.Lock()
	SavedTraits.TraitSavedCount++
	SavedTraits.mu.Unlock()
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
