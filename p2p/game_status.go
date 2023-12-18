package p2p

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusShuffleAndDeal:
		return "SHUFFLE AND DEAL"
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
	GameStatusShuffleAndDeal
	GameStatusReceivingCards
	GameStatusDealing 
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)