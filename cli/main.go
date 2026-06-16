package main

import (
	"log"
	"mfeeder/internal/cmd"
)

func main() {

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

}
