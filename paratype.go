// The Main package for Paratype type analysis software.
package main

import (
	"sync"
	"time"
	"fmt"
	"math/rand"
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
	Function	context.Function
	Channel		chan *Communication

	// Temporary channel for testing
	Tmp		chan string

	state		bool
	activeGroup	*sync.WaitGroup
}


// Global:
////////////////////////////////////////////////////////////////
// I am choosing to declare the slice of functions as global.
// By doing this, I am providing channels to all functions in
// their run routines. I am not confident this is the best way
// as of right now, but it will do.
const functionCount = 10
var functions [functionCount]FunctionActor

// A Functions main ruitine.
func (f *FunctionActor) Run() {
	f.makeActive(true)
	time.Sleep(time.Duration(5)*time.Second)
	id := rand.Int() % functionCount
	f.SendMessage("A message!", id)
	f.makeActive(false)
	for {
		message := <-f.Tmp
		f.makeActive(true)
		f.HandleMessage(message)
		f.makeActive(false)
	}
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

// A psuedo constructor for FunctionActors.
func (f *FunctionActor) Initialize(activeGroup *sync.WaitGroup) {
	f.activeGroup = activeGroup
	// Arbitrary buffer size. Note that channels block
	// only when the buffer is full.
	f.Channel = make(chan *Communication, 128)
	f.Tmp = make(chan string, 128)
}

// The outline of the function that sends messages to other actors.
// In the future the types of the arguments need to be changed. This 
// is only a proof of concept.
func (f *FunctionActor) SendMessage(message string, functionID int) {
	functions[functionID].Tmp <- message
}

// Function to handle messages from functions.
func (f *FunctionActor) HandleMessage(message string) {
	fmt.Printf("Function %s recieived a message:\n\t%s",
		f.Function.Name,
		message)
}
	

// Dummy main function.
func main() {
	// Make a set of junk functions
	// Run all functions
	// Wait to halt

	readyToFinish := new(sync.WaitGroup)

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
	fmt.Println("Done!")

}
