package context

import (
	"sync"
	"fmt"
	"strings"
)

func (f *Function) Run(Functions *map[*Function]bool) {

	// handling for function composition?
	if len(f.Parents) == 0 {
		f.makeActive(false)
	}

	for message := range f.Channel {
		f.makeActive(true)

		// debugging
		pathfuncs := PathToFunctions(message.Path, *Functions)
		s := make([]string, len(pathfuncs))
		for i, g := range pathfuncs {
			s[i] = g.Name
		}
		fmt.Printf("%v received from path %s the of %v\n",
			f.Name, strings.Join(s, "-"), message.Context.Name)

		// MERGE
		f.Update(message.Context)

		// add myself to path
		message.Path = AddToPath(message.Path, f)

		// send to all children
		for _, gfuncs := range f.Children {
			for g := range gfuncs {
				g.Channel <- message
			}
		}

		f.makeActive(false)
	}
}


// Change the state of the function actor. This is used
// for the halting conditions.
func (f *Function) makeActive(state bool) {
	if state == f.State {
		return
	}

	f.State = state
	if state {
		f.ActiveGroup.Add(1)
	} else {
		f.ActiveGroup.Done()
	}
}

// A pseudo constructor for Functions.
func (f *Function) Initialize(activeGroup *sync.WaitGroup) {
	f.ActiveGroup = activeGroup
	// Arbitrary buffer size. Note that Channels block
	// only when the buffer is full.
	f.Channel = make(chan *Communication, 128)
	f.makeActive(true)
}

// sends own to child
func (f *Function) InitialSendToChild() {
	comm := new(Communication)
	comm.Path = FunctionsToPath(f)
	comm.Context = f
	// for function composition, send to inner most children only
	for g := range f.Children[0] {
		g.Channel <-comm
	}
}
