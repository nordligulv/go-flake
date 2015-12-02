package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/nordligulv/go-flake"
)

var (
	max     = flag.Int("max", 1, "number of IDs to create")
	hex     = flag.Bool("hex", false, "Show hex representation")
	integer = flag.Bool("integer", false, "Show integer representation")
)

func main() {
	flag.Parse()
	f, err := flake.New(1)
	if err != nil {
		log.Fatal(err)
	}

	if !*hex && !*integer {
		*hex = true
	}

	for i := 0; i < *max; i++ {
		id := f.NextID()

		if *integer {
			fmt.Println(id)
		}

		if *hex {
			fmt.Println(id.String())
		}
	}
}
