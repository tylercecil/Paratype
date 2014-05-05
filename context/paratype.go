package context

import (
	"sync"
	//"fmt"
	//"strings"
)

// Runtime for each function actor
// to be used as a goroutine
func (f *Function) Run(Functions *map[*Function]bool, err chan error) {


	if f.WaitChildren != nil {
		// function composition:
		// wait for each level of children to return
		for {
			//fmt.Printf("%v waiting %v\n", f.Name, f.Depth)
			f.WaitChildren.Wait()

			f.Depth--

			if f.Depth == 0 {
				break
			} else {
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
					f.KillFlag.Wait()
					if f.Dead == true {
						return
					}
					g.Channel <-comm
				}
			}
		}
	}

	f.KillFlag.Wait()
	if f.Dead == true {
		return
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
			/*pathfuncs := PathToFunctions(message.Path, *Functions)
			s := make([]string, len(pathfuncs))
			for i, g := range pathfuncs {
				s[i] = g.Name
			}
			fmt.Printf("%v received from path %s the of %v\n",
				f.Name, strings.Join(s, "-"), message.Context.Name)*/


		depth := len(f.Children)

		var WaitChildren *sync.WaitGroup
		if depth >= 1 {
			WaitChildren = new(sync.WaitGroup)
		}

		for g := range f.Children[depth - 1] {
			comm := new(Communication)
			*comm = *message
			comm.Path = AddToPath(comm.Path, g)
			comm.Depth = depth - 1
			if depth >= 1 {
				comm.Wait = WaitChildren
				//fmt.Printf("adding %v %v\n", g, WaitChildren)
				WaitChildren.Add(1)
			}
			f.KillFlag.Wait()
			if f.Dead == true {
				return
			}

			g.Channel <-comm
		}


		if WaitChildren != nil {
			for {
				//fmt.Printf("%v waiting %v %v\n", f.Name, depth, WaitChildren)
				WaitChildren.Wait()

				depth--

				if depth == 0 {
					if message.Wait != nil {
						/*fmt.Printf("%v decrementing %v %v\n", f.Name,
						message.Wait, message.Path)*/
						message.Wait.Done()
					}
					break
				} else {
					for g := range f.Children[depth-1] {
						comm := new(Communication)
						*comm = *message
						comm.Path = AddToPath(comm.Path, g)
						comm.Depth = depth - 1
						if depth >= 1 {
							comm.Wait = WaitChildren
							WaitChildren.Add(1)
						}
						f.KillFlag.Wait()
						if f.Dead == true {
							return
						}
						g.Channel <-comm
					}
				}
			}
		}


		// add myself to path
		/*for _, gfuncs := range f.Children {
			for g := range gfuncs {
				msgCopy := new(Communication)
				msgCopy = message
				msgCopy.Path = AddToPath(msgCopy.Path, g)
				g.Channel <- msgCopy
			}
		}*/

		if len(f.Children) == 0 && message.Wait != nil {
			/*fmt.Printf("%v decrementing %v no child %v\n", f.Name,
			message.Wait, message.Path)*/

			//fmt.Printf("%+v %+v\n", PathToFunctionsmessage.Path, f.Name)
			message.Wait.Done()
		}

		// did I just send my last communication?
		if message.LastComm {
			err <- nil
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
