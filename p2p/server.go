package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"

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
	Version string
	ListenAddr string
	GameVariant GameVariant
}

type Server struct {
	ServerConfig

	transport *TCPTransport
	peerLock sync.RWMutex
	peers map[string] *Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
	broadcastch chan BroadcastTo
	gameState *GameState
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		ServerConfig: cfg,
		peers: make(map[string]*Peer),
		addPeer: make(chan *Peer, 100),
		delPeer: make(chan *Peer),
		msgCh: make(chan *Message, 100),
		broadcastch: make(chan BroadcastTo, 100),
	}

	s.gameState = NewGameState(s.ListenAddr, s.broadcastch)


	if s.ListenAddr == ":3000" {
		s.gameState.isDealer = true
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
	}).Info("started new game server")

	s.transport.ListenAndAccept()
}

func (s *Server) sendPeerList(p *Peer) error {

	peerList := MessagePeerList{
		Peers: []string{},
	}

	peers := s.Peers()
	for i := 0; i < len(peers); i++ {
		if peers[i] != p.listenAddr {
			peerList.Peers = append(peerList.Peers, peers[i])
		}
	}

	if len(peerList.Peers) == 0 {
		return nil
	}

	msg := NewMessage(s.ListenAddr, peerList)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) AddPeer(p *Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.listenAddr] = p
}

func (s *Server) Peers() []string {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()

	peers := make([]string, len(s.peers))
	it := 0
	for _, peer := range s.peers {
		peers[it] = peer.listenAddr
		it++
	}

	return peers
}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		GameVariant: s.GameVariant,
		Version: s.Version,
		GameStatus: s.gameState.gameStatus,
		ListenAddr: s.ListenAddr,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) isInPeerList(addr string) bool {
	peers := s.Peers()
	for i := 0; i < len(peers); i++ {
		if peers[i] == addr {
			return true
		}
	}
	return false
}

func (s *Server) Connect(addr string) error {
	if s.isInPeerList(addr) {
		return nil
	}

	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
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
		case msg := <-s.broadcastch:
			if err := s.Broadcast(msg); err != nil {
				logrus.Errorf("broadcast error: %s", err)
			}
		case peer := <-s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
				"local": peer.conn.LocalAddr(),
			}).Info("player disconnected")

			delete(s.peers, peer.conn.RemoteAddr().String())
		case peer := <-s.addPeer:
			if err := s.handleNewPeer(peer); err != nil {
				logrus.Errorf("handle peer error: %s", err)
			}
			
		case msg := <-s.msgCh:
			go func() {
				if err := s.handleMessage(msg); err != nil {
					logrus.Errorf("handle msg error: %s", err)
				}
			}()
		}
	}
}

func (s *Server) handleNewPeer(peer *Peer) error {
	hs, err := s.handshake(peer)
	if err != nil {
		peer.conn.Close()
		delete(s.peers, peer.conn.RemoteAddr().String())
		return fmt.Errorf("handshake failed with incoming player %s", err)
	}

	go peer.ReadLoop(s.msgCh)

	if !peer.outbound {
		if err := s.SendHandshake(peer); err != nil {
			peer.conn.Close()
			delete(s.peers, peer.conn.RemoteAddr().String())
			return fmt.Errorf("failed to dend handshake with peer %s", err)
		}

		go func() {
			if err := s.sendPeerList(peer); err != nil {
				logrus.Errorf("peerlist error: %s", err)
			}
		}()
		
	}

	logrus.WithFields(logrus.Fields{
		"peer": peer.conn.RemoteAddr(),
		"version": hs.Version,
		"variant": hs.GameVariant,
		"gameStatus": hs.GameStatus,
		"listenAddr": peer.listenAddr,
		"we": s.ListenAddr,
	}).Info("handshake successful new player connected")

	s.AddPeer(peer)

	s.gameState.AddPlayer(peer.listenAddr, hs.GameStatus)
	return nil
}

func (s *Server) Broadcast(broadcastMsg BroadcastTo) error {
	msg := NewMessage(s.ListenAddr, broadcastMsg.Payload)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	fmt.Printf("%v\n", broadcastMsg)

	for _, addr := range broadcastMsg.To {
		peer, ok := s.peers[addr]
		if ok {
			go func(peer *Peer) {
				if err := peer.Send(buf.Bytes()); err != nil {
					logrus.Errorf("broadcast to peer error: %s", err)
				}
				logrus.WithFields(logrus.Fields{
					"we": s.ListenAddr,
					"peer:": peer.listenAddr,
				}).Info("sending msg to peer")
			}(peer)
		}
		
	}

	// for _, peer := range s.peers {
	// 	go func(peer *Peer) {
	// 		if err := peer.Send(buf.Bytes()); err != nil {
	// 			logrus.Errorf("broadcast to peer error: %s", err)
	// 		}
	// 	}(peer)
	// }
	return nil
}

func (s *Server) handshake(p *Peer) (*Handshake, error) {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return nil, err
	}
	
	if s.GameVariant != hs.GameVariant {
		return nil, fmt.Errorf("gamevariant does not match %s", hs.GameVariant)
	}
	if s.Version != hs.Version {
		return nil, fmt.Errorf("version does not match %s", hs.Version)
	}

	p.listenAddr = hs.ListenAddr

	return hs, nil
}

func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessagePeerList:
		return s.handlePeerList(v)
	case MessageEncDeck:
		return s.handleEncDeck(msg.From, v)
	}
	return nil
}

func (s *Server) handleEncDeck(from string, msg MessageEncDeck) error {
	logrus.WithFields(logrus.Fields{
		"we": s.ListenAddr,
		"from": from,
	}).Info("recv enc deck")

	return s.gameState.ShuffleAndEncrypt(from, msg.Deck)
}

func (s *Server) handlePeerList(l MessagePeerList) error {
	logrus.WithFields(logrus.Fields{
		"we": s.ListenAddr,
		"list:": l.Peers,
	})//.Info("received PeerList")
	for i := 0; i < len(l.Peers); i++ {
		if err := s.Connect(l.Peers[i]); err != nil {
			logrus.Errorf("failed to dial peer: %s", err)
			continue
		}
	}

	return nil
}

func init () {
	gob.Register(MessagePeerList{})
	gob.Register(MessageEncDeck{})
}