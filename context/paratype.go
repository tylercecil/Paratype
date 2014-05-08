package context

import "sync"

// send message to function, return whether it was successful
func (f *Function) sendMessage(g *Function, msg *Communication) bool {
	f.KillFlag.Wait()
	if f.Dead == true {
		return false
	}
	g.Channel <- msg
	return true
}

// send error message back to main thread
func (f *Function) sendError(errorChannel chan error, err error) bool {
	f.KillFlag.Wait()
	if f.Dead == true {
		return false
	}
	errorChannel <- err
	return true
}

// Runtime for each function actor
// To be used as a goroutine
func (f *Function) Run(Functions *map[*Function]bool, err chan error) {

	// Function composition
	// Send to children at f.Depth first, wait for them to finish resolving
	// types and then decrease f.Depth, repeating until Depth is 1
	f.Depth = len(f.Children)
	if f.SendToChildren() == false {
		return
	}

	if f.WaitChildren != nil {
		// function composition:
		// wait for each level of children to return
		for {
			f.WaitChildren.Wait()

			f.Depth--

			if f.Depth == 0 {
				break
			}

			if f.SendToChildren() == false {
				return
			}
		}
	}

	// If I have no parents, send a message back to main thread saying that I'm
	// done. Halt if I couldn't send that message.
	// taking advantage of short-circuiting!
	if len(f.Parents) == 0 && f.sendError(err, nil) == false {
		return
	}

	// Receive messages
	for message := range f.Channel {
		// Halting: keeping track of how many parents are done sending
		if message.LastComm {
			f.NumParentsDone++
		}

		// This is _my_ last communication if all my parents are done
		message.LastComm = (f.NumParentsDone == len(f.Parents))

		// MERGE: add information to the function that is contained in
		// message.Context
		er := f.Update(message.Context)

		// send an error back if there was one and abort
		// taking advantage of short-circuiting!
		if er != nil && f.sendError(err, er) == false {
			return
		}

		// send the previously received context to all my children
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

// sends own to children at current depth
// returns false if unsuccessful and function actor should abort
func (f *Function) SendToChildren() bool {
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
		if f.sendMessage(g, comm) == false {
			return false
		}
	}
	return true
}
