package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

type GameVariant uint8

func (gv GameVariant) String() string {
	switch gv {
	case TexasHoldem:
		return "TEXAS HOLDEM"
	case Other:
		return "Other"
	default:
		return "unknown"
	}
}

const (
	TexasHoldem GameVariant = iota
	Other
)

type ServerConfig struct {
	ID int
	Version string
	ListenAddr string
	GameVariant GameVariant
}

type Server struct {
	ServerConfig

	transport *TCPTransport
	mu sync.RWMutex
	peers map[net.Addr] *Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
	gameState *GameState
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		ServerConfig: cfg,
		peers: make(map[net.Addr]*Peer),
		addPeer: make(chan *Peer),
		delPeer: make(chan *Peer),
		msgCh: make(chan *Message),
		gameState: NewGameState(),
	}

	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr
	
	tr.AddPeer = s.addPeer
	tr.DelPeer = s.delPeer

	return s
}

func (s *Server) Start() {
	go s.loop()

	logrus.WithFields(logrus.Fields{
		"port": s.ListenAddr,
		"variant": s.GameVariant,
		"gameStatus": s.gameState.gameStatus,
	}).Info("started new game server")

	s.transport.ListenAndAccept()
}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		GameVariant: s.GameVariant,
		Version: s.Version,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	peer := &Peer{
		conn: conn,
		outbound: true,
	}

	s.addPeer <- peer

	return s.SendHandshake(peer)
}

func(s *Server) loop() {
	for {
		select {
		case peer := <-s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
				"local": peer.conn.LocalAddr(),
			}).Info("player disconnected")

			delete(s.peers, peer.conn.RemoteAddr())
		case peer := <-s.addPeer:
			if err := s.handshake(peer); err != nil {
				logrus.Errorf("handshake failed with incoming player %s", err)
				peer.conn.Close()
				delete(s.peers, peer.conn.RemoteAddr())
				continue
			}

			go peer.readLoop(s.msgCh)

			if !peer.outbound {
				if err := s.SendHandshake(peer); err != nil {
					logrus.Errorf("failed to dend handshake with peer %s", err)
					peer.conn.Close()
					delete(s.peers, peer.conn.RemoteAddr())
					continue
				}
			}

			logrus.WithFields(logrus.Fields{
				"id": s.ID,
				"addr": peer.conn.RemoteAddr(),
				"local": peer.conn.LocalAddr(),
			}).Info("handshake successful new player connected")

			s.peers[peer.conn.RemoteAddr()] = peer
			
		case msg := <-s.msgCh:
			if err := s.HandleMessage(msg); err != nil {
				panic(err)
			}
		}
	}
}

func (s *Server) handshake(p *Peer) error {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return err
	}
	
	if s.GameVariant != hs.GameVariant {
		return fmt.Errorf("gamevariant does not match %s", hs.GameVariant)
	}
	if s.Version != hs.Version {
		return fmt.Errorf("version does not match %s", hs.Version)
	}

	logrus.WithFields(logrus.Fields{
		"ID": s.ID,
		"peer": p.conn.RemoteAddr(),
		"version": hs.Version,
		"variant": hs.GameVariant,
	}).Info("received handshake")
	return nil
}

func (s *Server) HandleMessage(msg *Message) error {
	fmt.Printf("%+v\n", msg)
	return nil
}