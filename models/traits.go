package models

import (
	"encoding/json"
	"image_generator/utils"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
)

func NewTraits() *TraitsManager {
	traits := &TraitsManager{}
	traits.Traits = make(map[string]Trait)
	traits.Mapping = make(map[int]string)
	return traits
}

type TraitsManager struct {
	BaseFolder string `json:"baseFolder"`
	Traits map[string]Trait
	N int `json:"traitCount"`
	Mapping map[int]string
	Config   *TraitManagerConfig   `json:"config"`
}
type TraitManagerConfig struct {
	Normal    int      `json:"normal"`
	Rare      int      `json:"rare"`
	SuperRare int      `json:"superRare"`
}

func (t *TraitsManager) AddAll() {
	folders, err := ioutil.ReadDir(t.BaseFolder)
	if err != nil {
		utils.Fatal(err)
	}
	for _, folder := range folders {
		if folder.IsDir() {
			var trait Trait
			trait.BasePath = t.BaseFolder
			trait.AddAll(folder.Name())
			t.Traits[trait.Name] = trait // todo configure ?? :))

			number := getNumber(folder.Name())
			t.Mapping[number] = trait.Name
		}
	}
}

func (t * TraitsManager) GetTraitKeys() []int {
	var traitKeys []int
	for key := range t.Mapping {
		traitKeys = append(traitKeys, key)
	}
	sort.Ints(traitKeys)
	return traitKeys
}

func (t *TraitsManager) Configure() {
	filePath := t.BaseFolder + "/config.json"
	if !utils.FleExists( filePath) {
		t.Config = &TraitManagerConfig{
			Normal:    100,
			Rare:      25,
			SuperRare: 5,
		}
		return
	}

	var config TraitManagerConfig
	body := utils.ReadAll(filePath)

	err := json.Unmarshal(body, &config)
	if err != nil {
		log.Panic(err)
	}
	t.Config = &config
}

func (t TraitsManager) Print() {
	utils.PrintJson(t.Traits)
}

func getNumber(folderName string) int {
	index := strings.Index(folderName, "_")
	numberStr := folderName[:index]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		utils.Fatal(err)
	}
	return number
}
