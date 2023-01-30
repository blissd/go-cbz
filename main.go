package main

import (
	"archive/zip"
	"fmt"
	"log"
)

func main() {

	input, err := zip.OpenReader("comic.cbz")
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}

	for _, file := range input.File {
		fmt.Println("name: ", file.Name)
	}
}
