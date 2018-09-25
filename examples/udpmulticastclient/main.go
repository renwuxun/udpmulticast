package main

import (
	"../../../udpmulticast"
)

func main() {
	peer := udpmulticast.NewMultiCastPeer("224.0.1.100", 9527)

	peer.JoinGroup()
	peer.Send([]byte("hello!"))
	peer.Destroy()
}
