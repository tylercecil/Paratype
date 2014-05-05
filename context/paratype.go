package context

import (
	"sync"
	"fmt"
	"strings"
)

func (f *Function) Run(Functions *map[*Function]bool, err chan error) {

	for message := range f.Channel {
		// halting
		if message.LastComm {
			f.NumParentsDone++
		}

		if f.NumParentsDone < len(f.Parents) {
			message.LastComm = false
		} else {
			message.LastComm = true
		}

		// debugging
		pathfuncs := PathToFunctions(message.Path, *Functions)
		s := make([]string, len(pathfuncs))
		for i, g := range pathfuncs {
			s[i] = g.Name
		}
		fmt.Printf("%v received from path %s the of %v\n",
			f.Name, strings.Join(s, "-"), message.Context.Name)

		// MERGE
		er := f.Update(message.Context)
		if er != nil {
			err <- er
			return
		}

		// add myself to path
		message.Path = AddToPath(message.Path, f)
		for _, gfuncs := range f.Children {
			msgCopy := new(Communication)
			msgCopy = message
			for g := range gfuncs {
				g.Channel <- msgCopy
			}
		}

		// did I just send my last communication?
		if message.LastComm {
			err <- nil
		}
	}

	// implicit barrier through channel closing
	if f.Implement {
		f.Finish()
		f.ImplementationWait.Done()
	}

	return
}

// A pseudo constructor for Functions.
func (f *Function) Initialize(implWait *sync.WaitGroup) {
	// Arbitrary buffer size. Note that Channels block
	// only when the buffer is full.
	f.Channel = make(chan *Communication, 128)
	f.ImplementationWait = implWait
	//implWait.Add(1)
}

// sends own to child
func (f *Function) InitialSendToChild() {
	comm := new(Communication)
	comm.Path = FunctionsToPath(f)
	comm.Context = f
	// for function composition, send to inner most children only
	deepest := len(f.Children)
	comm.Depth = deepest - 1

	if len(f.Parents) == 0 {
		comm.LastComm = true
	} else {
		comm.LastComm = false
	}
	for g := range f.Children[deepest - 1] {
		g.Channel <-comm
	}
}
