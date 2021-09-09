package main

import (
	"errors"
	"fmt"
	"image_generator/models"
	"image_generator/utils"
	"log"
	"math/rand"
	"time"
)

var counter int64

func getRandom(max int) int {
	rand.Seed(time.Now().UnixNano() + counter)
	number := rand.Intn(max)
	counter++
	return number
}
func main() {
	outputFolder := "output"
	inputFolder := "layers"
	err := utils.EnsureDir(outputFolder)
	if err != nil {
		log.Fatal(err)
	}

	traits := models.NewTraits()
	traits.BaseFolder = inputFolder
	traits.Configure()
	traits.AddAll()
	//traits.Print()

	traitKeys := traits.GetTraitKeys()
	for i:=0; i<15; i++ {
		fmt.Println(i)
		imageCreator := NewImageCreator()

		var includeTraits []string
		var excludeTraits []string
		var excludeSingleTraits = make( map[string][]string)

		for _, keyNumber := range traitKeys {
			key := traits.Mapping[keyNumber]
			trait := traits.Traits[key]

			if len(trait.TraitConfig.Exclude) > 0 && len(trait.TraitConfig.Include) > 0 {
				utils.Fatal(errors.New("include and exclude defined"))
			}

			traitTypeNumber := getRandom(100 * 100)
			if keyNumber == 1 {// TODO remove
				for traitTypeNumber > 100 {
					traitTypeNumber = getRandom(100 * 100)
				}
			}
			if trait.MainTraitType == models.TraitSuperRare {
				if traitTypeNumber > traits.Config.SuperRare*100 {
					continue
				}
			} else if trait.MainTraitType == models.TraitRare {
				if traitTypeNumber > traits.Config.Rare*100 {
					continue
				}
			}

			if len(excludeTraits) > 0 && utils.ExistIn(trait.Name, excludeTraits) {
				continue
			}

			if len(includeTraits) > 0 && !utils.ExistIn(trait.Name, includeTraits) {
				continue
			}

			includeTraits = append(includeTraits, trait.TraitConfig.Include...)
			excludeTraits = append(excludeTraits, trait.TraitConfig.Exclude...)
			for keyExclude,value := range trait.TraitConfig.ExcludeSingle {
				excludeSingleTraits[keyExclude] = append(excludeSingleTraits[keyExclude], value...)
			}

			traitTypeToUse := models.TraitNormal

			if trait.MainTraitType != models.TraitSuperRare  && trait.MainTraitType != models.TraitRare {
				randomTraitsTypeMax := getRandom(100 * 100)
				//if keyNumber == 0 || keyNumber == 1 { // TODO remove
				//	for randomTraitsTypeMax > 1000 {
				//		randomTraitsTypeMax = getRandom(100 * 100)
				//	}
				//}
				if randomTraitsTypeMax <= trait.TraitConfig.SuperRare*100 {
					traitTypeToUse = models.TraitSuperRare
				} else if randomTraitsTypeMax <= trait.TraitConfig.Rare*100 {
					traitTypeToUse = models.TraitRare
				}
			}

			traitsToUSe := trait.GetTraitsByType(traitTypeToUse, excludeSingleTraits[trait.Name])
			n := len(traitsToUSe)
			if n == 0 {
				continue
			}
			randomTraitNumber := getRandom(n)

			choosedTrait := traitsToUSe[randomTraitNumber]
			if choosedTrait.Config != nil{
				includeTraits = append(includeTraits, choosedTrait.Config.Include...)
				excludeTraits = append(excludeTraits, choosedTrait.Config.Exclude...)
				for keyExclude,value := range choosedTrait.Config.ExcludeSingle{
					excludeSingleTraits[keyExclude] = append(excludeSingleTraits[keyExclude], value...)
				}
			}
			imageCreator.Add(choosedTrait.BasePath)
		}
		_ = imageCreator.Process()
		imageCreator.WriteTo(fmt.Sprintf(outputFolder+"/%d.png", i))
	}
	//
	//var n = 0
	//for i := 0; i < n; i++ {
	//	imageCreator := NewImageCreator()
	//	imageCreator.Add(
	//		"layers/0_background/gray.png",
	//		"layers/2_specialBehind_r/fire.png",
	//		"layers/5_skin/skin2.png",
	//		"layers/8_hoodie/black_hoodie.png",
	//		"layers/6_hair/brownhair.png",
	//		"layers/7_short/greenshorts.png",
	//		"layers/10_shue/redGreen.png",
	//		"layers/3_tail/3_tail.png",
	//	)
	//	_ = imageCreator.Process()
	//	imageCreator.WriteTo(fmt.Sprintf(outputFolder + "/%d.png", i))
	//}

}
