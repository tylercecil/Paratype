// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	"fmt"
	"Paratype/context"
	"runtime"
)

var Functions map[*context.Function]bool

var NumThreadsActive int

// Given a list of functions, will spawn function actors and resolve types
// Returns a list of type errors collected
// 
func RunActors(f ...interface{}) []error {
	Functions = make(map[*context.Function]bool)

	// one can pass in multiple Function pointers or a slice of them
	// tests are usually multiple while the parser will generate a slice; i.e.
	// RunActors(f, g, h)
	// and
	// RunActors([]*context.Function{f, g, h}) 
	// are equivalent
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

	// wait group that waits for implementation collection to finish
	implementationWait := new(sync.WaitGroup)

	// wait group to assist with halting when a type error is detected
	killFlag := new(sync.WaitGroup)

	// channel to send errors back to here from function actors
	err := make(chan error, len(Functions))

	for fActor := range Functions {
		fActor.Initialize(implementationWait, killFlag)
	}

	// avoid race conditions by having the first communication in Channels
	// before starting
	for fActor := range Functions {
		fActor.InitialSendToChild()
	}

	for fActor := range Functions {
		//fmt.Printf("\tSpawning Function Actor for %v\n", fActor.Name)
		implementationWait.Add(1)
		go fActor.Run(&Functions, err)
		NumThreadsActive++
	}

	var s []error
	// listen to error channel
	for er := range err {
		// every function will send something through the error channel back to
		// RunActors: it will send nil if it finished without type errors and
		// it will send the type error if one arose.
		if er != nil {
			fmt.Printf("%v\n", er.Error())
			s = append(s, er)
		} else {
			// one goroutine finished
			NumThreadsActive--
		}

		if NumThreadsActive == 0 {
			// all goroutines finished without type errors

			// close all channels
			for f := range Functions {
				f.Implement = true
				close(f.Channel)
			}

			implementationWait.Wait()

			close(err)

			break;
		} else if er != nil {
			// type errors arose when goroutines ran

			// set all goroutines to stop when next possible
			killFlag.Add(1)
			for f := range Functions {
				// disable implementation collection
				f.Implement = false
				f.Dead = true
				defer close(f.Channel)
			}
			close(err)
			killFlag.Done()
			break;
		}
	}

	return s
}


// Takes the number of processors and a list of functions
// Runs paratype and will print all implementations of functions to screen
// 
func RunParatype(n int, f ...interface{}) {
	runtime.GOMAXPROCS(n)
	var funcs []*context.Function

	// f may either be an array of Function pointers or just many of them; i.e. 
	// RunParatype(4, f, g, h)
	// and
	// RunParatype(4, []*context.Function{f, g, h}) 
	// are equivalent
	switch f[0].(type) {
	case []*context.Function:
		for _, fun := range f[0].([]*context.Function) {
			funcs = append(funcs, fun)
		}

	case *context.Function:
		for _, fun := range f {
			funcs = append(funcs, fun.(*context.Function))
		}
	}

	// run actors, collect type errors
	errors := RunActors(funcs)
	if len(errors) > 0 {
		// print type errors if there are any
		for _, e := range errors {
			fmt.Printf("%+v\n", e)
		}
	} else {
		fmt.Printf("\n=== Implementations ===\n\n")

		noprint := false
		for _, f := range funcs {
			// type errros that arose during collection of implementations
			// (means that there is a missing implementation)
			if f.TypeError != nil {
				fmt.Println(f.Name, f.TypeError.Error())
				noprint = true
			}
		}

		if noprint == false {
			// we have working implementations! print 'em.
			for _, f := range funcs {
				for _, typemap := range f.Implementations {
					f.PrintImplementation(typemap)
				}
			}
		}

		fmt.Printf("\n")
	}
}

// Dummy main function.
func main() {
}
