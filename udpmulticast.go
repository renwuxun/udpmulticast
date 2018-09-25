package udpmulticast

import (
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"strings"
	"sync"
)

type MultiCastPerr struct {
	group        *net.UDPAddr
	ipv4PConn    *ipv4.PacketConn
	netPConn     net.PacketConn
	recvPacketCB func(b []byte, src net.Addr, fromLocalIp bool)
	err          error
}

var pUdpMultiCastPeer = &sync.Pool{
	New: func() interface{} {
		peer := new(MultiCastPerr)
		return peer
	},
}

var pBuffer = &sync.Pool{
	New: func() interface{} {
		b := make([]byte, 4096, 4096)
		return b
	},
}

func NewMultiCastPeer(groupIp string, groupPort int) *MultiCastPerr {
	peer := pUdpMultiCastPeer.Get().(*MultiCastPerr)

	peer.group = &net.UDPAddr{Port: groupPort, IP: net.ParseIP(groupIp)}
	peer.netPConn, peer.err = net.ListenPacket("udp4", fmt.Sprintf("%s:%d", groupIp, groupPort))
	if nil != peer.err {
		log.Fatalf("%v", peer.err)
	}
	peer.ipv4PConn = ipv4.NewPacketConn(peer.netPConn)

	return peer
}

func (ths *MultiCastPerr) Destroy() {
	ths.ipv4PConn.Close()

	ths.netPConn.Close()
	pUdpMultiCastPeer.Put(ths)
}

func (ths *MultiCastPerr) OnRecv(recvPacketCB func([]byte, net.Addr, bool)) *MultiCastPerr {
	ths.recvPacketCB = recvPacketCB
	return ths
}

func (ths *MultiCastPerr) JoinGroup() *MultiCastPerr {
	InterfacesUpAndMulticast(func(i net.Interface) {
		if err := ths.ipv4PConn.JoinGroup(&i, ths.group); nil != err {
			ths.err = err
		}
	})
	return ths
}

func (ths *MultiCastPerr) Listen() {
	localIPs := LocalIPs()
	for {
		b := pBuffer.Get().([]byte)
		b = b[:cap(b)]
		n, _, src, err := ths.ipv4PConn.ReadFrom(b)
		if nil != err {
			ths.err = err
			pBuffer.Put(b)
			continue
		}
		b = b[:n]
		if nil != ths.recvPacketCB {
			go func() {
				if _, ok := localIPs[strings.Split(src.String(), ":")[0]]; ok {
					ths.recvPacketCB(b, src, true)
				} else {
					ths.recvPacketCB(b, src, false)
				}
				pBuffer.Put(b)
			}()
		} else {
			pBuffer.Put(b)
		}
	}
}

func (ths *MultiCastPerr) Send(b []byte) *MultiCastPerr {
	InterfacesUpAndMulticast(func(i net.Interface) {
		if err := ths.ipv4PConn.SetMulticastInterface(&i); nil != err {
			ths.err = err
		}
		ths.ipv4PConn.SetMulticastTTL(2)
		if _, err := ths.ipv4PConn.WriteTo(b, nil, ths.group); nil != err {
			ths.err = err
		}
	})
	return ths
}

func InterfacesUpAndMulticast(cb func(net.Interface)) {
	ifis, err := net.Interfaces()
	if nil != err {
		return
	}

	for i := range ifis {
		if ifis[i].Flags.String() == "up|broadcast|multicast" {
			cb(ifis[i])
		}
	}
}

func LocalIPs() (ips map[string]struct{}) {
	ips = make(map[string]struct{})
	ips["localhost"] = struct{}{}
	ips["127.0.0.1"] = struct{}{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, address := range addrs {
		ips[strings.Split(address.String(), "/")[0]] = struct{}{}
	}
	return
}
