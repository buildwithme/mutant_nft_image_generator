package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image_generator/models"
	"image_generator/utils"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zenthangplus/goccm"
)

var counterMu sync.Mutex
var counter int64

func getRandom(max int) int {
	counterMu.Lock()
	rand.Seed(time.Now().UnixNano() + counter)
	number := rand.Intn(max)
	counter++
	counterMu.Unlock()
	return number
}
func main() {
	processor()
}

func processor() {
	var n int = 500
	//argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) == 0 {
		argsWithoutProg = []string{strconv.Itoa(n)}
	}

	if len(argsWithoutProg) > 0 {
		var err error
		n, err = strconv.Atoi(argsWithoutProg[0])
		if err != nil {
			log.Fatal(err)
		}
	}

	execute("_3heads", "layers_3heads", n)
}
func execute(outputFolder, inputFolder string, n int) {
	baseOutput := "output"
	err := utils.EnsureDir(baseOutput + "/" + outputFolder)
	if err != nil {
		log.Fatal(err)
	}

	traits := models.NewTraits()
	traits.BaseFolder = inputFolder
	traits.Configure()
	traits.AddAll()

	// Limit x goroutines to run concurrently.
	c := goccm.New(50)

	var mu sync.Mutex
	var unusedTraits = make(map[string]map[int]struct{})

	for key, value := range traits.Traits {
		if _, ok := unusedTraits[key]; !ok {
			unusedTraits[key] = make(map[int]struct{})
		}
		for singleKey := range value.Traits {
			if _, ok := unusedTraits[key]; !ok {
				unusedTraits[key] = make(map[int]struct{})
			}
			unusedTraits[key][singleKey] = struct{}{}
		}
	}

	var counter uint64
	traitKeys := traits.GetTraitKeys()
	for index := 0; index < n; index++ {
		c.Wait()

		go func(i int) {
			defer c.Done()

			mu.Lock()
			pickedUniqueSingleTrait := len(unusedTraits) == 0
			mu.Unlock()

			fmt.Printf("%s - %d\n", inputFolder, i)
			imageCreator := models.NewImageCreator(i)

			for _, keyNumber := range traitKeys {
				key := traits.Mapping[keyNumber]
				trait := traits.Traits[key]

				// fmt.Printf("%s - %d layer [%s]\n", inputFolder, i, key)

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
						rareConfig := traits.Config.Rare
						if key == "Accessory 2" {
							rareConfig = 50
						}
						if traitTypeNumber > rareConfig*100 {
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
				var traitsToUSe []*models.SingleTrait
				var randomTraitNumber int

				if !pickedUniqueSingleTrait {
					traitsToUSe = trait.GetTraitsFiltered(imageCreator.IncludeSingleTraits[trait.Name], imageCreator.ExcludeSingleTraits[trait.Name])
				} else {
					traitsToUSe = trait.GetTraitsByType(traitTypeToUse, imageCreator.IncludeSingleTraits[trait.Name], imageCreator.ExcludeSingleTraits[trait.Name])

				}
				n := len(traitsToUSe)
				if n == 0 {
					continue
				}
				if !pickedUniqueSingleTrait {
					mu.Lock()
					if len(unusedTraits[key]) > 0 {
						for kk := range unusedTraits[key] {
							randomTraitNumber = kk
							delete(unusedTraits[key], kk)
							if len(unusedTraits[key]) == 0 {
								delete(unusedTraits, key)
							}
							pickedUniqueSingleTrait = true
							break
						}
					}
					mu.Unlock()
				} else {
					randomTraitNumber = getRandom(n)
				}

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
			imageCreator.WriteTo(fmt.Sprintf(baseOutput+"/%s/%d.png", outputFolder, i+1))
		}(index)
	}

	time.Sleep(1000)
	c.WaitAllDone()

	// utils.PrintJson(models.SavedTraits)

	m := generateOsMetadata(models.SavedTraits.Data, n)
	// utils.PrintJson(m)
	writeToFile("output/all_metadata_OS", m)

	rez := make(map[string]map[string]int)
	rez = rez
	for i := 0; i < len(m); i++ {
		elem := m[i]["attributes"].([]map[string]interface{})
		for j := 0; j < len(elem); j++ {
			attribute := elem[j]
			attribute = attribute

			traitType := attribute["trait_type"].(string)
			value := attribute["value"].(string)
			if _, ok := rez[traitType]; !ok {
				rez[traitType] = make(map[string]int)
			}
			rez[traitType][value]++
		}
	}
	body, err := json.MarshalIndent(rez, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile("output/rarity.json", body, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func generateOsMetadata(datas map[int]map[int]models.TraitSavedConf, n int) []map[string]interface{} {
	createDir("output/metadata_OS")
	var metadata []map[string]interface{}
	for i := 0; i < n; i++ {
		key := datas[i]
		tokenID := strconv.Itoa(i + 1)

		var meta = make(map[string]interface{})

		meta["name"] = "Reaper " + tokenID
		meta["image"] = "IPFS_URL/" + tokenID + ".png"

		var attributes []map[string]interface{}

		for _, val := range key {

			val.Value = strings.ReplaceAll(val.Value, "_sr.png", "")
			val.Value = strings.ReplaceAll(val.Value, "_r.png", "")
			val.Value = strings.ReplaceAll(val.Value, ".png", "")
			attributes = append(attributes, getNewAttribute(val.TraitType, val.Value)...)
		}

		meta["attributes"] = attributes
		metadata = append(metadata, meta)

		body, err := json.MarshalIndent(meta, "", "\t")
		if err != nil {
			fmt.Println(err)
		}
		err = ioutil.WriteFile("output/metadata_OS/"+tokenID, body, 0777)
		if err != nil {
			fmt.Println("Error creating", "data"+strconv.Itoa(i+1))
			fmt.Println(err)
		}
	}

	return metadata
}

func writeToFile(name string, metadata []map[string]interface{}) {
	body, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(name+".json", body, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}

	createDir(name)
	for i, m := range metadata {

		body, err := json.MarshalIndent(m, "", "\t")
		if err != nil {
			fmt.Println(err)
		}
		err = ioutil.WriteFile(name+"/"+strconv.Itoa(i+1)+".json", body, 0777)
		if err != nil {
			fmt.Println("Error creating", "data"+strconv.Itoa(i+1))
			fmt.Println(err)
			return
		}
	}
}
func createDir(name string) {
	err := os.Mkdir(name, os.ModeDir)
	if err != nil {
		fmt.Println(err)
	}
	err = os.Chmod(name, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
}

func getNewAttribute(traitType, attributeValue string) (v []map[string]interface{}) {
	if len(attributeValue) > 0 {
		v = append(v, map[string]interface{}{
			"trait_type": traitType,
			"value":      attributeValue,
		})
	}
	return
}

func addNewAttribute(traitType, attributeValue string, m map[string]interface{}) {
	if len(attributeValue) > 0 {
		m[traitType] = attributeValue
	}
}

func getInt(value string) int {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatal(err)
	}
	return intValue
}
