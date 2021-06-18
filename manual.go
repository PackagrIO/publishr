package main

import (
	"github.com/packagrio/go-common/utils/git"
	"log"
)


func main() {

	result, err := git.GitFindNearestTagName("/source")
	if err != nil {
		log.Fatalf("Error found: %v", err)
	} else {
		log.Printf("Success!, %v", result)
	}
}
