package main

import (
	"../../../udpmulticast"
	"fmt"
	"net"
)

func main() {
	peer := udpmulticast.NewMultiCastPeer("224.0.1.100", 9527)

	peer.OnRecv(func(b []byte, src net.Addr, fromLocalIp bool) {
		fmt.Printf("[recv:%s, from:%v, is from local:%v]\n", b, src.String(), fromLocalIp)
	})
	peer.JoinGroup()
	peer.Listen()
	peer.Destroy()
}
