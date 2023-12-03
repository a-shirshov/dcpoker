package main

import (
	"fmt"
	"time"

	"github.com/a-shirshov/dcpoker/deck"
	"github.com/a-shirshov/dcpoker/p2p"
)

func main() {
	cfg := p2p.ServerConfig{
		Version: "DCPoker v.0.1-alpha",
		ListenAddr: ":3000",
	}

	server := p2p.NewServer(cfg)
	go server.Start()

	time.Sleep(1 * time.Second)

	remoteCfg := p2p.ServerConfig{
		Version: "DCPoker v.0.1-alpha",
		ListenAddr: ":4000",
	}
	remoteServer := p2p.NewServer(remoteCfg)
	go remoteServer.Start()
	remoteServer.Connect(":3000")
	fmt.Println(deck.New())
	
}
