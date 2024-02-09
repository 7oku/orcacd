package main

import (
	"sync"
)

var config = NewConfig()

func main() {
	printBanner()

	stacks, dsession := initialize(config)

	if len(stacks) == 0 {
		logOrcacd.Fatalf("No Stacks to work on. Did you define a repo?")
	}

	var wg sync.WaitGroup
	for _, stack := range stacks {
		wg.Add(1)
		go stack.RunPuller(dsession, config, &wg)
	}
	wg.Wait()

}
