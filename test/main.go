package main

import (
	"fmt"
	"os"

	"github.com/Eslam-Nawara/foreman"
)

func main() {
	foreman, err := foreman.New("../test-procfiles/Procfile", false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = foreman.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
