package main

import (
	"errors"
	"fmt"
	"image_generator/models"
	"image_generator/utils"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/zenthangplus/goccm"
)

var counter int64

func getRandom(max int) int {
	rand.Seed(time.Now().UnixNano() + counter)
	number := rand.Intn(max)
	counter++
	return number
}

func main() {
	//argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]

	n, err := strconv.Atoi(argsWithoutProg[0])
	if err != nil {
		log.Fatal(err)
	}

	execute("images", "layers_2", n)
	//execute("astronaut_layers", "_astronaut_layers", 100)
	//execute("robot_layers", "_robot_layers", 100)
	//execute("aliens_layers", "_aliens_layers", 100)
	//execute("flame_layers", "_flame_layers", 100)
}
func execute(outputFolder, inputFolder string, n int) {
	err := utils.EnsureDir(outputFolder)
	if err != nil {
		log.Fatal(err)
	}

	traits := models.NewTraits()
	traits.BaseFolder = inputFolder
	traits.Configure()
	traits.AddAll()

	// Limit 3 goroutines to run concurrently.
	c := goccm.New(10)

	var counter uint64
	traitKeys := traits.GetTraitKeys()
	for index := 0; index < n; index++ {
		c.Wait()

		go func(i int) {
			defer c.Done()

			fmt.Printf("%s - %d\n", inputFolder, i)
			imageCreator := models.NewImageCreator(i)

			for _, keyNumber := range traitKeys {
				key := traits.Mapping[keyNumber]
				trait := traits.Traits[key]

				if trait.TraitConfig != nil && len(trait.TraitConfig.Exclude) > 0 && len(trait.TraitConfig.Include) > 0 {
					utils.Fatal(errors.New("include and exclude defined"))
				}

				if !trait.TraitConfig.Required {
					traitTypeNumber := getRandom(100 * 100)
					if trait.MainTraitType == models.TraitSuperRare {
						if traitTypeNumber > traits.Config.SuperRare*100 {
							continue
						}
					} else if trait.MainTraitType == models.TraitRare {
						if traitTypeNumber > traits.Config.Rare*100 {
							continue
						}
					}
				}

				if len(imageCreator.ExcludeTraits) > 0 && utils.ExistIn(trait.Name, imageCreator.ExcludeTraits) {
					continue
				}

				if len(imageCreator.IncludeTraits) > 0 && !utils.ExistIn(trait.Name, imageCreator.IncludeTraits) {
					continue
				}

				imageCreator.IncludeTraits = append(imageCreator.IncludeTraits, trait.TraitConfig.Include...)
				imageCreator.ExcludeTraits = append(imageCreator.ExcludeTraits, trait.TraitConfig.Exclude...)
				for keyExclude, value := range trait.TraitConfig.ExcludeSingle {
					if !utils.ExistIn(keyExclude, imageCreator.ExcludeTraits) {
						imageCreator.ExcludeTraits = append(imageCreator.ExcludeTraits, keyExclude)
					}
					imageCreator.ExcludeSingleTraits[keyExclude] = append(imageCreator.ExcludeSingleTraits[keyExclude], value...)
				}
				for keyInclude, value := range trait.TraitConfig.IncludeSingle {
					if !utils.ExistIn(keyInclude, imageCreator.IncludeTraits) {
						imageCreator.IncludeTraits = append(imageCreator.IncludeTraits, keyInclude)
					}
					imageCreator.IncludeSingleTraits[keyInclude] = append(imageCreator.IncludeSingleTraits[keyInclude], value...)
				}

				var traitTypeToUse models.TraitType

				if trait.MainTraitType == models.TraitNormal {
					randomTraitsTypeMax := getRandom(100 * 100)
					if randomTraitsTypeMax <= trait.TraitConfig.SuperRare*100 {
						traitTypeToUse = models.TraitSuperRare
					} else if randomTraitsTypeMax <= trait.TraitConfig.Rare*100 {
						traitTypeToUse = models.TraitRare
					} else {
						traitTypeToUse = models.TraitNormal
					}
				} else if trait.MainTraitType == models.TraitRare {
					randomTraitsTypeMax := getRandom(100 * 100)
					if randomTraitsTypeMax <= trait.TraitConfig.SuperRare*100 {
						traitTypeToUse = models.TraitSuperRare
					} else {
						traitTypeToUse = models.TraitRare
					}
				} else {
					traitTypeToUse = models.TraitSuperRare
				}

				traitsToUSe := trait.GetTraitsByType(traitTypeToUse, imageCreator.IncludeSingleTraits[trait.Name], imageCreator.ExcludeSingleTraits[trait.Name])
				n := len(traitsToUSe)
				if n == 0 {
					continue
				}
				randomTraitNumber := getRandom(n)

				choosedTrait := traitsToUSe[randomTraitNumber]
				imageCreator.Add(trait, choosedTrait)
				if choosedTrait.Config != nil {
					imageCreator.IncludeTraits = append(imageCreator.IncludeTraits, choosedTrait.Config.Include...)
					imageCreator.ExcludeTraits = append(imageCreator.ExcludeTraits, choosedTrait.Config.Exclude...)
					for keyExclude, value := range choosedTrait.Config.ExcludeSingle {
						imageCreator.ExcludeSingleTraits[keyExclude] = append(imageCreator.ExcludeSingleTraits[keyExclude], value...)
					}
					for keyInclude, value := range choosedTrait.Config.IncludeSingle {
						imageCreator.IncludeSingleTraits[keyInclude] = append(imageCreator.IncludeSingleTraits[keyInclude], value...)
					}
				}
			}
			_ = imageCreator.Process()

			if !imageCreator.IsHashValid() {
				return
			}
			atomic.AddUint64(&counter, 1)
			imageCreator.WriteTo(fmt.Sprintf("output/%s/%d.png", outputFolder, i))
		}(index)
	}

	time.Sleep(1000)
	c.WaitAllDone()

	utils.PrintJson(models.SavedTraits)
}
