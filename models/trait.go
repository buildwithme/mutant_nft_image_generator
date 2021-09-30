package models

import (
	"encoding/json"
	"image_generator/utils"
	"io/ioutil"
	"log"
	"strings"
)

type Trait struct {
	MainTraitType TraitType      `json:"type"`
	Traits        []*SingleTrait `json:"traits"`
	TraitConfig   *TraitConfig   `json:"config"`
	Name          string         `json:"name"`
	FolderName    string         `json:"folderName"`
	BasePath      string         `json:"basePath"`
}
func (t *Trait) GetTraitsByType(traitTypeToUse TraitType, include, exclude []string) []*SingleTrait{
	var traitsToReturn []*SingleTrait
	for _, traitValue := range t.Traits {
		if traitValue.TraitType == traitTypeToUse && !utils.ExistIn(traitValue.Name, exclude) &&
			(len(include) == 0 || utils.ExistIn(traitValue.Name, include)){
			traitsToReturn = append(traitsToReturn, traitValue)
		}
	}
	if len(traitsToReturn) == 0 {
		switch traitTypeToUse {
		case TraitSuperRare:
			return t.GetTraitsByType(TraitRare, include, exclude)
		case TraitRare:
			return t.GetTraitsByType(TraitNormal, include, exclude)
		}
	}
	return traitsToReturn
}
type TraitConfig struct {
	Normal    int      `json:"normal"`
	Rare      int      `json:"rare"`
	SuperRare int      `json:"superRare"`
	Count     int      `json:"count"` //  number of traits of same type that can be picked
	Include   []string `json:"include"`
	Exclude   []string `json:"exclude"`
	ExcludeSingle   map[string][]string `json:"excludeSingle"`
	IncludeSingle   map[string][]string `json:"includeSingle"`
	Required bool `json:"required"`
}

func NewTraitConfigFrom(path string) *TraitConfig {
	if !utils.FleExists( path) {
		return &TraitConfig{
			Normal:        100,
			Rare:          25,
			SuperRare:     5,
			Count:         0,
			Include:       nil,
			Exclude:       nil,
			ExcludeSingle: make( map[string][]string),
			IncludeSingle: make( map[string][]string),
			Required:      false,
		}
	}

	config := TraitConfig{}

	body := utils.ReadAll(path)

	err := json.Unmarshal(body, &config)
	if err != nil {
		log.Panic(err)
	}
	if config.Count == 0 {
		config.Count++ // TODO ... multiple traits
	}
	return &config
}

//func (t *Trait) NewTraitConfigFrom(configPath) TraitConfig {
//	traitConfig := TraitConfig{}
//
//	body := utils.ReadAll(fmt.Sprintf(t.Base+"/%s/", folderName, configFileName))
//
//	err = json.Unmarshal(body, &rarityConfigValue)
//	if err != nil {
//		log.Panic(err)
//	}
//}

func (t *Trait) AddName(folderName string) {
	t.FolderName = folderName
	t.BasePath = t.BasePath + "/" + folderName

	input := t.FolderName
	index := strings.Index(input, "_") + 1

	input = input[index:]
	index = strings.Index(input, "_")
	if index == -1 {
		t.Name = input
		return
	} else {
		t.Name = input[:index]
	}

	input = input[index:]
	if strings.Contains(input, "_") {
		if strings.Contains(input, "_sr") {
			t.MainTraitType = TraitSuperRare
		} else if strings.Contains(input, "_r") {
			t.MainTraitType = TraitRare
		}
	}
}

func (t *Trait) AddAll(folderName string) {
	t.AddName(folderName)
	files, err := ioutil.ReadDir(t.BasePath)
	if err != nil {
		log.Panic(err)
	}
	t.TraitConfig = NewTraitConfigFrom(t.BasePath + "/" + configFileName)
	for _, file := range files {
		if file.Name() == configFileName {
			continue
		}
		if extension := utils.GetExtension(file.Name()); extension != "png" {
			continue
		}

		t.Traits = append(t.Traits, NewSingleTraitFrom(file.Name(), t.BasePath))
	}
}

func (t Trait) Print() {
	t.PrintTraits()
	t.PrintConfig()
}

func (t Trait) PrintConfig() {
	utils.PrintJson(t.TraitConfig)
}

func (t Trait) PrintTraits() {
	utils.PrintJson(t.Traits)
}
