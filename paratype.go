// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	"fmt"
	"Paratype/context"
	"strings"
)

// Object to represent a communication
type Communication struct {
	path	string
	context	*context.Function
}

// Object to represent the Function as an actor.
// This is probably bad design (we have structs everywhere!)
// but will be fine for now.
type FunctionActor struct {
	function	*context.Function
	channel		chan *Communication
	state		bool
	activeGroup	*sync.WaitGroup
}


// Global:
////////////////////////////////////////////////////////////////
// I am choosing to declare the slice of functions as global.
// By doing this, I am providing channels to all functions in
// their run routines. I am not confident this is the best way
// as of right now, but it will do.
var functions map[*context.Function]*FunctionActor
// hackish. merge Function and FunctionActor?
var functionsArray []*context.Function

// Runtime of a function
func (f *FunctionActor) Run() {
	f.makeActive(false)

	// handling for function composition?

	for message := range f.channel {
		f.makeActive(true)

		// debugging
		fmt.Printf("%v received from path ", f.function.Name)
		pathfuncs := context.PathToFunctions(message.path, functionsArray)
		s := make([]string, len(pathfuncs))
		for i, g := range pathfuncs {
			s[i] = g.Name
		}
		fmt.Printf("%s the context of %v\n", strings.Join(s, "-"), message.context.Name)

		// MERGE
		f.function.Update(message.context)

		// add myself to path
		message.path = context.AddToPath(message.path, f.function)

		// send to all children
		for _, gfuncs := range f.function.Children {
			for g := range gfuncs {
				functions[g].channel <- message
			}
		}

		f.makeActive(false)
	}
}


// Change the state of the function actor. This is used
// for the halting conditions.
func (f *FunctionActor) makeActive(state bool) {
	if state == f.state {
		return
	}

	f.state = state
	if state {
		f.activeGroup.Add(1)
	} else {
		f.activeGroup.Done()
	}
}

// A pseudo constructor for FunctionActors.
func (f *FunctionActor) Initialize(activeGroup *sync.WaitGroup) {
	f.activeGroup = activeGroup
	// Arbitrary buffer size. Note that channels block
	// only when the buffer is full.
	f.channel = make(chan *Communication, 128)
	f.makeActive(true)
}

// sends own context to child
func (f *FunctionActor) InitialSendToChild() {
	comm := new(Communication)
	comm.path = context.FunctionsToPath(f.function)
	comm.context = f.function
	// for function composition, send to inner most children only
	for g := range f.function.Children[0] {
		functions[g].channel <-comm
	}
}


// given a list of functions, run everything!
func RunThings(f ...interface{}) {

	functions = make(map[*context.Function]*FunctionActor)

	// one can pass in multiple Function pointers or a slice of them
	// tests are usually multiple while the parser will generate a slice
	switch f[0].(type) {
	case []*context.Function:
		for _, fun := range f[0].([]*context.Function) {
			fActor := new(FunctionActor)
			fActor.function = fun
			functions[fun] = fActor
			functionsArray = append(functionsArray, fun)
		}

	case *context.Function:
		for _, fun := range f {
			fActor := new(FunctionActor)
			fActor.function = fun.(*context.Function)
			functions[fun.(*context.Function)] = fActor
			functionsArray = append(functionsArray, fun.(*context.Function))
		}
	}

	readyToFinish := new(sync.WaitGroup)

	//fmt.Println("Welcome to Paratype!")

	for _, fActor := range functions {
		fActor.Initialize(readyToFinish)
	}
	// avoid race conditions by having the first communication in channels
	// before starting
	for _, fActor := range functions {
		fActor.InitialSendToChild()
	}
	for _, fActor := range functions {
		//fmt.Printf("\tSpawning Function Actor for %v\n", fActor.function.Name)
		go fActor.Run()
	}

	//fmt.Println("Waiting for halting...")

	// This is actually a race condition. It WOULD be sufficient
	// to both make this check AND check if all channels are
	// empty.
	readyToFinish.Wait()
	for _, fActor := range functions {
		// close channels, otherwise goroutines will hang
		close(fActor.channel)
	}

	//fmt.Println("Done!")
}

// Dummy main function.
func main() {
}
