// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	//"time"
	"fmt"
	//"math/rand"
	"Paratype/context"
)

// Object to represent a communication
type Communication struct {
	Path	string
	Context	*context.Function
}

// Object to represent the Function as an actor.
// This is probably bad design (we have structs everywhere!)
// but will be fine for now.
type FunctionActor struct {
	Function	*context.Function
	Channel		chan *Communication

	// Temporary channel for testing
	Tmp			chan string

	state		bool
	activeGroup	*sync.WaitGroup
}


// Global:
////////////////////////////////////////////////////////////////
// I am choosing to declare the slice of functions as global.
// By doing this, I am providing channels to all functions in
// their run routines. I am not confident this is the best way
// as of right now, but it will do.
//const functionCount = 10
var functions map[*context.Function]*FunctionActor

// A Functions main ruitine.
func (f *FunctionActor) Run() {
	f.makeActive(false)
	for message := range f.Channel {

		f.makeActive(true)
		f.Function.Update(message.Context)

		for _, gfuncs := range f.Function.Children {
			for g := range gfuncs {
				functions[g].Channel <- message
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

// A psuedo constructor for FunctionActors.
func (f *FunctionActor) Initialize(activeGroup *sync.WaitGroup) {
	f.activeGroup = activeGroup
	// Arbitrary buffer size. Note that channels block
	// only when the buffer is full.
	f.Channel = make(chan *Communication, 128)
	f.Tmp = make(chan string, 128)
	f.makeActive(true)
}

func (f *FunctionActor) SendToChild() {
	comm := new(Communication)
	comm.Path = context.ConvertPath(f.Function)
	comm.Context = f.Function
	for _, gfuncs := range f.Function.Children {
		for g := range gfuncs {
			functions[g].Channel <-comm
		}
	}
}

// not ready
func RunThings(f ...interface{}) {

	functions = make(map[*context.Function]*FunctionActor)
	for _, fun := range f {
		fActor := new(FunctionActor)
		fActor.Function = fun.(*context.Function)
		functions[fun.(*context.Function)] = fActor
	}

	readyToFinish := new(sync.WaitGroup)

	fmt.Println("Welcome to Paratype!")

	for _, fActor := range functions {
		fmt.Printf("\tSpawning Function Actor for %v\n", fActor.Function.Name)
		fActor.Initialize(readyToFinish)
	}
	for _, fActor := range functions {
		fActor.SendToChild()
	}
	for _, fActor := range functions {
		go fActor.Run()
	}

	fmt.Println("Waiting for halting...")
	// This is actually a race condition. It WOULD be sufficient
	// to both make this check AND check if all channels are
	// empty.

	readyToFinish.Wait()
	for _, fActor := range functions {
		close(fActor.Channel)
		close(fActor.Tmp)
	}

	fmt.Println("Done!")

}

// Dummy main function.
func main() {
	// Make a set of junk functions
	// Run all functions
	// Wait to halt

	//RunThings()

	/*readyToFinish := new(sync.WaitGroup)

	fmt.Println("Welcome to Paratype!")

	for i, fActor := range functions {
		fmt.Printf("\tSpawning %d Function Actor\n", i)
		fActor.Initialize(readyToFinish)
		go fActor.Run()
	}

	fmt.Println("Waiting for halting...")
	// This is actually a race condition. It WOULD be sufficient
	// to both make this check AND check if all channels are
	// empty.
	readyToFinish.Wait()
	fmt.Println("Done!")*/

}
