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

	implementationWait := new(sync.WaitGroup)
	killFlag := new(sync.WaitGroup)
	err := make(chan error, len(Functions))

	fmt.Println("Welcome to Paratype!")

	for fActor := range Functions {
		fActor.Initialize(implementationWait, killFlag)
	}
	// avoid race conditions by having the first communication in Channels
	// before starting
	for fActor := range Functions {
		fActor.InitialSendToChild()
	}

	for fActor := range Functions {
		fmt.Printf("\tSpawning Function Actor for %v\n", fActor.Name)
		if len(fActor.Parents) > 0 {
			implementationWait.Add(1)
			go fActor.Run(&Functions, err)
			NumThreadsActive++
		}
	}

	var s []error
	// errors
	for er := range err {
		if er != nil {
			fmt.Printf("%v\n", er.Error())
			s = append(s, er)
		} else {
			NumThreadsActive--
		}

		if NumThreadsActive == 0 {
			// close all channels
			for f := range Functions {
				f.Implement = true
				close(f.Channel)
			}

			implementationWait.Wait()

			close(err)

			break;
		} else if er != nil {
			killFlag.Add(1)
			for f := range Functions {
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


func RunThem(n int, f ...interface{}) {
	runtime.GOMAXPROCS(n)
	var funcs []*context.Function

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

	errors := RunThings(funcs)
	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Printf("%+v\n", e)
		}
	} else {
		fmt.Printf("\n===implementations===\n\n")
		/*switch f[0].(type) {
		case []*context.Function:
			for _, fun := range f[0].([]*context.Function) {
				//fun.Finish()
			}

		case *context.Function:
			for _, fun := range f {
				//fun.(*context.Function).Finish()
			}
		}*/

		fmt.Printf("\n")
	}
}



// Dummy main function.
func main() {
}
