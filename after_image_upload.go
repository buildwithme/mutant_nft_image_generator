package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

var n int = 1000

func afterImageUpload() {
	//https://gateway.pinata.cloud/ipfs/Qmdd9HqK4KQgShqBjYDEZg674nBLaVdbgy2LqqE2MYJvAX/1.png
	urls := []string{
		"https://gateway.pinata.cloud/ipfs/QmTmZTrfzGjpTZhagiE56s2qSKweRX4pQKG8KqWZ186eT5",
		"https://gateway.pinata.cloud/ipfs/QmcYsgfEAosccnfvPncg7RyJCi39wFSqiwvBoyD4asZEJg",
	}
	i := 1
	for index, url := range urls {
		for ; i <= (index+1)*n; i++ {
			appendImageURLToMetadataFiles(url, fmt.Sprintf("%s/final/finalMetadata/%d", baseOutput, i),
				fmt.Sprintf("%s/final/finalMetadata/%d", baseOutput, i))
			if i%20 == 0 {
				fmt.Printf("afterImageUpload - %d\n", i)
			}
		}
	}
}

func appendImageURLToMetadataFiles(url, sourceFile, destinationFile string) {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	var m = make(map[string]interface{})
	err = json.Unmarshal(input, &m)
	if err != nil {
		log.Fatal(err)
	}
	m["image"] = strings.ReplaceAll(m["image"].(string), "IPFS_URL", url)

	input, err = json.MarshalIndent(m, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		fmt.Println("Error creating", destinationFile)
		fmt.Println(err)
		return
	}
}
