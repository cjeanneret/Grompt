package main

import (
	"log"

	"grompt/internal/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
