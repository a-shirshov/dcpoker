package p2p

import (
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Version string
	ListenAddr string
}

type Server struct {
	ServerConfig

	handler Handler
	transport *TCPTransport
	mu sync.RWMutex
	peers map[net.Addr] *Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		handler: &DefaultHandler{},
		ServerConfig: cfg,
		peers: make(map[net.Addr]*Peer),
		addPeer: make(chan *Peer),
		delPeer: make(chan *Peer),
		msgCh: make(chan *Message),
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
		"type": "Texas Holdem",
	}).Info("started new game server")
	s.transport.ListenAndAccept()
}

func (s *Server) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	peer := &Peer{
		conn: conn,
	}

	s.addPeer <- peer

	return peer.Send([]byte(fmt.Sprintf("%s%s",s.Version,"\n")))
}

func(s *Server) loop() {
	for {
		select {
		case peer := <-s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
			}).Info("player disconnected")
			delete(s.peers, peer.conn.RemoteAddr())
		case peer := <-s.addPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
			}).Info("new player connected")
			s.peers[peer.conn.RemoteAddr()] = peer
			go peer.readLoop(s.msgCh)
		case msg := <-s.msgCh:
			if err := s.handler.HandleMessage(msg); err != nil {
				panic(err)
			}
		}
	}
}