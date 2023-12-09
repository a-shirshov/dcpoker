package p2p

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type GameStatus uint32

func (g GameStatus) String() string {
	switch g{
	case GameStatusWaitingForCards:
		return "WAITING FOR CARDS"
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
	isDealer bool
	gameStatus GameStatus

	playersWaitingForCards int32
	playerLock sync.RWMutex
	players map[string]*Player
}

func (g *GameState) AddPlayerWaitingForCards() {
	atomic.AddInt32(&g.playersWaitingForCards, 1)
}

func (g *GameState) CheckNeedDealCards() {
	playersWaiting := atomic.LoadInt32(&g.playersWaitingForCards)

	if playersWaiting == int32(len(g.players)) && 
	g.isDealer && 
	g.gameStatus == GameStatusWaitingForCards {
		logrus.Println("dealing cards")
	}
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

	//g.CheckNeedDealCards()

	logrus.WithFields(logrus.Fields{
		"addr": addr,
		"status": status,
	}).Info("new player joined")
}

func NewGameState() *GameState {
	g := &GameState{
		isDealer: false,
		gameStatus: GameStatusWaitingForCards,
		players: make(map[string]*Player),
	}

	go g.loop()

	return g
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
