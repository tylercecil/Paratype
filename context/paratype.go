package context

import (
	"sync"
	//"fmt"
	//"strings"
)

// send message to function, return whether it was successful
func (f *Function) sendMessage(g *Function, msg *Communication) bool {
	f.KillFlag.Wait()
	if f.Dead == true {
		return false
	}
	g.Channel <- msg
	return true
}

func (f *Function) sendError(errorChannel chan error, err error) bool {
	f.KillFlag.Wait()
	if f.Dead == true {
		return false
	}
	errorChannel <- err
	return true
}

// Runtime for each function actor
// to be used as a goroutine
func (f *Function) Run(Functions *map[*Function]bool, err chan error) {

	f.Depth = len(f.Children)
	f.SendToChildren()

	if f.WaitChildren != nil {
		// function composition:
		// wait for each level of children to return
		for {
			f.WaitChildren.Wait()

			f.Depth--

			if f.Depth == 0 {
				break
			}

			f.SendToChildren()
		}
	}

	// taking advantage of short-circuiting!
	if len(f.Parents) == 0 && f.sendError(err, nil) == false {
		return
	}

	for message := range f.Channel {
		// halting
		if message.LastComm {
			f.NumParentsDone++
		}

		// this is _my_ last communication if my parents are done
		// TODO: should I check whether my buffer is empty?
		message.LastComm = (f.NumParentsDone == len(f.Parents))

		// // debugging
		// fmt.Printf("%v received from path %s the of %v\n",
		// f.Name, PrintablePath(message.Path, *Functions), 
		// message.Context.Name)

		// MERGE
		er := f.Update(message.Context)

		// taking advantage of short-circuiting!
		if er != nil && f.sendError(err, er) == false {
			return
		}

		// send myself to children
		for level, gfuncs := range f.Children {
			for g := range gfuncs {
				msgCopy := new(Communication)

				// copy the message
				*msgCopy = *message

				// in anticipation for function composition with repeats of
				// callees
				msgCopy.Path = AddToPath(msgCopy.Path, g)

				// Necessary for function composition: every leaf notifies the
				// composing function for completion. The composing function 
				// assumes that there is only one leaf for each callee, but 
				// what if one of the callees is also composed? Then, there are
				// more leafs and the waitgroup has to be incremented further.
				if level >= 1 && message.Wait != nil {
					message.Wait.Add(1)
				}

				if f.sendMessage(g, msgCopy) == false {
					return
				}
			}
		}

		// Function composition: decrement the waitgroup if I'm a leaf and I am
		// inheriting from a composed call
		if len(f.Children) == 0 && message.Wait != nil {
			message.Wait.Done()
		}

		// did I just send my last communication?
		// taking advantage of short-circuiting
		if message.LastComm && f.sendError(err, nil) == false{
			return
		}
	}

	// implicit barrier through channel closing
	if f.Implement {
		f.Implementations, f.TypeError = f.Finish()
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
}

// sends own to child
func (f *Function) SendToChildren() {
	if f.Depth > 1 && f.WaitChildren == nil {
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
