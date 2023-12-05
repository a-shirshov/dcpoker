package main

import (
	"time"

	"github.com/a-shirshov/dcpoker/p2p"
)

func main() {
	cfg := p2p.ServerConfig{
		ID: 1,
		Version: "DCPoker v.0.1-alpha",
		ListenAddr: ":3000",
		GameVariant: p2p.TexasHoldem,
	}

	server := p2p.NewServer(cfg)
	go server.Start()

	time.Sleep(1 * time.Second)

	remoteCfg := p2p.ServerConfig{
		ID: 2,
		Version: "DCPoker v.0.1-alpha",
		ListenAddr: ":4000",
		GameVariant: p2p.TexasHoldem,
	}
	remoteServer := p2p.NewServer(remoteCfg)
	go remoteServer.Start()

	remoteServer.Connect(":3000")
	time.Sleep(5 * time.Second)
}
