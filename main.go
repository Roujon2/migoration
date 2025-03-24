package main

import (
	"log"

	"github.com/Roujon2/migoration/cmd"
)


func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}