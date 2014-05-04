// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	"fmt"
	"Paratype/context"
	"runtime"
)

var Functions map[*context.Function]bool

// should return implementations -- PARALLEL COMMUNICATION NEEDED
// given a list of functions, run everything!
func RunThings(f ...interface{}) []error {
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
	err := make(chan error, len(Functions))

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
		go fActor.Run(&Functions, err)
	}

	fmt.Println("Waiting for halting...")

	// RACE CONDITION
	// This is actually a race condition. It WOULD be sufficient
	// to both make this check AND check if all Channels are
	// empty.
ShittyGoto:
	for fActor := range Functions {
		// close Channels, otherwise goroutines will hang
		if len(fActor.Channel) > 0 {
			break ShittyGoto
		}
		close(fActor.Channel)
	}

	readyToFinish.Wait()

	fmt.Println("Done!", len(err))

	// collect error messages
	var s []error
	if len(err) > 0 {
		s = make([]error, len(err))
		for i := 0; len(err) > 0; i++ {
			m := <-err
			s[0] = m
		}
	}

	close(err)
	return s
}


func RunThem(n int, f ...interface{}) {
	runtime.GOMAXPROCS(n)
	var funcs []*context.Function
	for _, fun := range f {
		funcs = append(funcs, fun.(*context.Function))
	}
	errors := RunThings(funcs)
	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Println(e.Error())
		}
	} else {
		fmt.Printf("\n===implementations===\n\n")
		for _, fun := range f {
			fun.(*context.Function).Finish()
		}
		fmt.Printf("\n")
	}
}



// Dummy main function.
func main() {
}
