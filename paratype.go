// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	"time"
	"fmt"
	"Paratype/context"
)

// Object to represent a communication
type Communication struct {
	path	context.Path
	context	*context.Context
}

// Object to represent the Function as an actor.
// This is probably bad design (we have structs everywhere!)
// but will be fine for now.
type FunctionActor struct {
	function	context.Function
	channel		chan *Communication
	state		bool
	activeGroup	*sync.WaitGroup
}

// A Functions main ruitine.
func (f *FunctionActor) Run() {
	f.makeActive(true)
	time.Sleep(time.Duration(5)*time.Second)
	f.makeActive(false)
}


// Change the state of the function actor. This is used
// for the halting conditions.
func (f *FunctionActor) makeActive(state bool) {
	if state == f.state {
		return
	}

	if state {
		f.activeGroup.Add(1)
	} else {
		f.activeGroup.Done()
	}
}

// Dummy main function.
func main() {
	// Make a set of junk functions
	// Run all functions
	// Wait to halt

	readyToFinish := new(sync.WaitGroup)
	var functions [10]FunctionActor

	fmt.Println("Welcome to Paratype!")

	for i, fActor := range functions {
		fmt.Printf("\tSpawning %d Function Actor\n", i)
		fActor.activeGroup = readyToFinish
		go fActor.Run()
	}
	
	fmt.Println("Waiting for halting...")
	readyToFinish.Wait()
	fmt.Println("Done!")

}
