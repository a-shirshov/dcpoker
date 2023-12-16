package p2p

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-shirshov/dcpoker/deck"
	"github.com/sirupsen/logrus"
)

type GameStatus int32

func (g GameStatus) String() string {
	switch g{
	case GameStatusWaitingForCards:
		return "WAITING FOR CARDS"
	case GameStatusReceivingCards:
		return "RECEIVING CARDS"
	case GameStatusDealing:
		return "DEALING"
	case GameStatusPreFlop:
		return "PREFLOP"
	case GameStatusFlop:
		return "FLOP"
	case GameStatusTurn:
		return "TURN"
	case GameStatusRiver:
		return "RIVER"
	default:
		return "unknown"
	}
}

const (
	GameStatusWaitingForCards GameStatus = iota
	GameStatusReceivingCards
	GameStatusDealing 
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type Player struct {
	Status GameStatus
}

type GameState struct {
	listenAddr string
	broadcastch chan any
	isDealer bool
	gameStatus GameStatus

	playersWaitingForCards int32
	playerLock sync.RWMutex
	players map[string]*Player
}

func NewGameState(addr string,broadcastch chan any) *GameState {
	g := &GameState{
		listenAddr: addr,
		broadcastch: broadcastch,
		isDealer: false,
		gameStatus: GameStatusWaitingForCards,
		players: make(map[string]*Player),
	}

	go g.loop()

	return g
}

func (g *GameState) SetStatus(s GameStatus) {
	if g.gameStatus != s {
		atomic.StoreInt32((*int32)(&g.gameStatus), int32(s))
	}
}

func (g *GameState) AddPlayerWaitingForCards() {
	atomic.AddInt32(&g.playersWaitingForCards, 1)
}

func (g *GameState) CheckNeedDealCards() {
	playersWaiting := atomic.LoadInt32(&g.playersWaitingForCards)

	if playersWaiting == int32(len(g.players)) && 
	g.isDealer && 
	g.gameStatus == GameStatusWaitingForCards {
		logrus.WithFields(logrus.Fields{
			"address": g.listenAddr,
		}).Info("need to deal cards")

		g.DealCards()
	}
}

func (g *GameState) InitiateShuffleAndDeal() {
	g.SetStatus(GameStatusReceivingCards)
	
}

func (g *GameState) DealCards() {
	deck := deck.New()
	cards := MessageCards {
		Deck: deck,
	}

	g.broadcastch <- cards
}

func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
	player, ok := g.players[addr]
	if !ok {
		panic("player could not be found, although it should exist")
	}

	player.Status = status

	g.CheckNeedDealCards()
}

func (g *GameState) LenPlayersConnectedWithLock() int {
	g.playerLock.RLock()
	defer g.playerLock.RUnlock()

	return len(g.players)
}

func (g *GameState) AddPlayer(addr string, status GameStatus) {
	g.playerLock.Lock()
	defer g.playerLock.Unlock()

	if status == GameStatusWaitingForCards {
		g.AddPlayerWaitingForCards()
	}

	g.players[addr] = &Player{}

	g.SetPlayerStatus(addr, status)

	logrus.WithFields(logrus.Fields{
		"addr": addr,
		"status": status,
	}).Info("new player joined")
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			logrus.WithFields(logrus.Fields{
				"players connected": g.LenPlayersConnectedWithLock(),
				"status": g.gameStatus,
			}).Info()
		}
	}
}
