package main

import (
	"time"

	"github.com/a-shirshov/dcpoker/p2p"
)

func makeServerAndStart(addr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version: "DCPoker v.0.1-alpha",
		ListenAddr: addr,
		GameVariant: p2p.TexasHoldem,
	}

	server := p2p.NewServer(cfg)
	go server.Start()
	time.Sleep(1 * time.Second)
	return server
}

func main() {
	playerA := makeServerAndStart(":3000")
	playerB := makeServerAndStart(":4000")
	playerC := makeServerAndStart(":5000")
	playerD := makeServerAndStart(":6000")

	time.Sleep(1 * time.Second)
	playerB.Connect(playerA.ListenAddr)
	time.Sleep(1 * time.Second)
	playerC.Connect(playerB.ListenAddr)
	time.Sleep(1 * time.Second)
	playerD.Connect(playerC.ListenAddr)
	//_ = playerA
	//_ = playerB
	select {}
}


