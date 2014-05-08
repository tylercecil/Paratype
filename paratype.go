// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	"fmt"
	"Paratype/context"
	"Paratype/paraparse"
	"runtime"
	"os"
	"flag"
	"time"
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

		//context.PrintAll(fActor)
	}

	for fActor := range Functions {
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
			}
			killFlag.Done()

			for f := range Functions {
				defer close(f.Channel)
			}
			defer close(err)
			break;
		}
	}

	return s
}


// Takes the number of processors and a list of functions
// Runs paratype and will print all implementations of functions to screen
//
func RunParatype(n int, out string, hprint bool, f ...interface{}) {
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
			fmt.Println(e)
		}
	} else {

		/*for _, f := range funcs {
			context.PrintAll(f)
		}*/

		noprint := false
		for _, f := range funcs {
			// type errros that arose during collection of implementations
			// (means that there is a missing implementation)
			if f.TypeError != nil {
				fmt.Println(f.TypeError.Error())
				noprint = true
			}
		}

		if noprint == false && hprint == true {
			fi, err := os.Create(out)
			if err != nil {
				panic(err)
			}
			defer fi.Close()
			// we have working implementations! print 'em.
			for _, f := range funcs {
				for _, typemap := range f.Implementations {
					f.PrintImplementation(typemap, fi)
				}
			}
		}

	}
}

// Dummy main function.
func main() {
	procsPtr := flag.Int("procs", 4, "Number of processors.")
	printPtr := flag.Bool("print", false, "Should I print?")
	inFilePtr := flag.String("infile", "", "File to operate on.")
	outFilePtr := flag.String("outfile", "", "File to output to.")
	timePtr := flag.Bool("time", false, "Should I time?")

	flag.Parse()

	if *inFilePtr == "" {
		fmt.Println("ERROR: OH NO! Provide an input file!")
		return
	}

	if *printPtr == true && *outFilePtr == "" {
		fmt.Println("ERROR: OH NO! Provide a file to write results to!")
		return
	}
	begin := time.Now()
	flist, err := paraparse.Setup(*inFilePtr, true)
	end := time.Now()

	if *timePtr == true {
		fmt.Printf("SETUP: %d ", end.Sub(begin).Nanoseconds())
	}
	if err != nil {
		fmt.Printf("%+v", err)
		return
	}
	begin = time.Now()
	RunParatype(*procsPtr, *outFilePtr, *printPtr, flist)
	end = time.Now()
	if *timePtr == true {
		fmt.Printf("COMPLETION: %d\n", end.Sub(begin).Nanoseconds())
	}
}
