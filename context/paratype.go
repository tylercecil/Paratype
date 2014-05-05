package context

import (
	"sync"
//	"fmt"
//	"strings"
)

func (f *Function) Run(Functions *map[*Function]bool, err chan error) {
	if f.WaitChildren != nil {

		f.WaitChildren.Wait()

		f.Depth--

		if f.Depth >= 1 {
			for g := range f.Children[f.Depth-1] {
				comm := new(Communication)
				comm.Path = FunctionsToPath(f, g)
				comm.Context = f
				comm.Depth = f.Depth - 1
				comm.LastComm = (len(f.Parents) == 0)
				if f.Depth > 1 {
					comm.Wait = f.WaitChildren
					f.WaitChildren.Add(1)
				}
				g.Channel <-comm
			}
		}

		// send to next layer of children
	}

	if len(f.Parents) == 0 {
		err <- nil
	}

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

		// // debugging
		// pathfuncs := PathToFunctions(message.Path, *Functions)
		// s := make([]string, len(pathfuncs))
		// for i, g := range pathfuncs {
		// 	s[i] = g.Name
		// }
		// fmt.Printf("%v received from path %s the of %v\n",
		// 	f.Name, strings.Join(s, "-"), message.Context.Name)

		// MERGE
		er := f.Update(message.Context)
		f.KillFlag.Wait()
		if f.Dead == true {
			break;
		}
		if er != nil {
			err <- er
			return
		}

		// add myself to path
		for _, gfuncs := range f.Children {
			for g := range gfuncs {
				message.Path = AddToPath(message.Path, g)
				msgCopy := new(Communication)
				msgCopy = message
				g.Channel <- msgCopy
			}
		}

		if len(f.Children) == 0 && message.Wait != nil {
			message.Wait.Done()
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
func (f *Function) Initialize(implWait *sync.WaitGroup, killFlag *sync.WaitGroup) {
	// Arbitrary buffer size. Note that Channels block
	// only when the buffer is full.
	f.Channel = make(chan *Communication, 128)
	f.ImplementationWait = implWait
	f.KillFlag = killFlag
	//implWait.Add(1)
}

// sends own to child
func (f *Function) InitialSendToChild() {
	f.Depth = len(f.Children)

	if f.Depth > 1 {
		f.WaitChildren = new(sync.WaitGroup)
	}

	for g := range f.Children[f.Depth - 1] {
		comm := new(Communication)
		comm.Path = FunctionsToPath(f, g)
		comm.Context = f
		comm.Depth = f.Depth - 1
		comm.LastComm = (len(f.Parents) == 0)
		if f.Depth > 1 {
			comm.Wait = f.WaitChildren
			f.WaitChildren.Add(1)
		}
		g.Channel <-comm
	}
}
