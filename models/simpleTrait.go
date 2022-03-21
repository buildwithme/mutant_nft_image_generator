package models

import (
	"encoding/json"
	"image_generator/utils"
	"log"
	"strings"
)

func NewSingleTraitFrom(fileName string, basePath string) *SingleTrait {
	var trait SingleTrait
	trait.BasePath = basePath
	trait.AddName(fileName)
	trait.Configure()
	return &trait
}

type SingleTrait struct {
	Name      string             `json:"name"`
	FileName  string             `json:"fileName"`
	BasePath  string             `json:"basePath"`
	TraitType TraitType          `json:"type"`
	Config    *SingleTraitConfig `json:"config"`
}

func (t *SingleTrait) AddName(fileName string) {
	t.FileName = fileName
	t.BasePath = t.BasePath + "/" + fileName

	input := t.FileName
	index := strings.Index(input, "_") + 1

	input = input[index:]
	index = strings.Index(input, "_")
	if index == -1 {
		t.Name = input
	} else {
		t.Name = input[:index]
	}
}

func (t *SingleTrait) Configure() {
	if strings.Contains(t.FileName, "_sr") {
		t.TraitType = TraitSuperRare
		t.Name = strings.ReplaceAll(t.FileName, "_sr.png", "")
		t.Name = strings.ReplaceAll(t.FileName, "_sr.jpg", "")
	} else if strings.Contains(t.FileName, "_r") {
		t.TraitType = TraitRare
		t.Name = strings.ReplaceAll(t.FileName, "_r.png", "")
		t.Name = strings.ReplaceAll(t.FileName, "_r.jpg", "")
	} else {
		t.TraitType = TraitNormal
		t.Name = strings.ReplaceAll(t.FileName, ".png", "")
	}

	t.ConfigFile()
}
func (t *SingleTrait) ConfigFile() {
	filePath := t.GetConfigFileName()
	if !utils.FleExists(filePath) {
		t.Config = &SingleTraitConfig{
			Include:       nil,
			Exclude:       nil,
			IncludeSingle: make(map[string][]string),
			ExcludeSingle: make(map[string][]string),
		}
		return
	}

	var singleTraitConfig SingleTraitConfig
	body := utils.ReadAll(filePath)
	err := json.Unmarshal(body, &singleTraitConfig)
	if err != nil {
		log.Panic(err)
	}
	t.Config = &singleTraitConfig
}

func (t *SingleTrait) GetConfigFileName() string {
	outputName := t.BasePath
	outputName = strings.ReplaceAll(outputName, ".png", "")
	outputName = strings.ReplaceAll(outputName, ".jpg", "")
	return outputName + ".json"
}

type SingleTraitConfig struct {
	Include       []string            `json:"include"`
	Exclude       []string            `json:"exclude"`
	IncludeSingle map[string][]string `json:"includeSingle"`
	ExcludeSingle map[string][]string `json:"excludeSingle"`
}
