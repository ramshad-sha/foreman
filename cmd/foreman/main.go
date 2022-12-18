package main

import (
	"flag"
	"fmt"

	"github.com/Eslam-Nawara/foreman"
)

func main() {

	verbosePtr := flag.Bool("v", false, "run the program verbosely")
	procfilePtr := flag.String("f", "Procfile", "specify the procfile path")
	flag.Parse()

	f, err := foreman.New(*procfilePtr, *verbosePtr)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = f.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

}
