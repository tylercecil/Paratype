// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	"fmt"
	"Paratype/context"
	"runtime"
)

var Functions map[*context.Function]bool

// given a list of functions, run everything!
func RunThings(f ...interface{}) {
	runtime.GOMAXPROCS(2)
	Functions = make(map[*context.Function]bool)

	// one can pass in multiple Function pointers or a slice of them
	// tests are usually multiple while the parser will generate a slice
	switch f[0].(type) {
	case []*context.Function:
		for _, fun := range f[0].([]*context.Function) {
			Functions[fun] = true
		}

	case *context.Function:
		for _, fun := range f {
			Functions[fun.(*context.Function)] = true
		}
	}

	readyToFinish := new(sync.WaitGroup)

	fmt.Println("Welcome to Paratype!")

	for fActor := range Functions {
		fActor.Initialize(readyToFinish)
	}
	// avoid race conditions by having the first communication in Channels
	// before starting
	for fActor := range Functions {
		fActor.InitialSendToChild()
	}
	for fActor := range Functions {
		fmt.Printf("\tSpawning Function Actor for %v\n", fActor.Name)
		go fActor.Run(&Functions)
	}

	fmt.Println("Waiting for halting...")

	// This is actually a race condition. It WOULD be sufficient
	// to both make this check AND check if all Channels are
	// empty.
	readyToFinish.Wait()
	for fActor := range Functions {
		// close Channels, otherwise goroutines will hang
		if len(fActor.Channel) != 0 {
			break
		}
		close(fActor.Channel)
	}

	fmt.Println("Done!")
}


// Dummy main function.
func main() {
}
