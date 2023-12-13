package deck

import (
	"fmt"
	"math/rand"
	"strconv"
)

type Suit int

func (s Suit) String() string {
	switch s {
	case Spades:
		return "SPADES"
	case Hearts:
		return "HEARTS"
	case Clubs:
		return "CLUBS"
	case Diamonds:
		return "DIAMONDS"
	default:
		panic("invalid card suit")
	}
}

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

type Card struct {
	Suit Suit
	Value int
}

func (c Card) String() string {
	value := strconv.Itoa(c.Value)
	if c.Value == 1 {
		value = "ACE"
	}
	return fmt.Sprintf("%s of %s %s", value, c.Suit, suitToUnicode(c.Suit))
}

func NewCard(s Suit, v int) Card {
	if v > 13 {
		panic("big value")
	}
	return Card{
		Suit: s,
		Value: v,
	}
}

type Deck [52]Card 

func New() Deck {
	nSuits := 4
	nCards := 13
	d := [52]Card{}

	for i := 0; i < nSuits; i++ {
		for j := 0; j < nCards; j++ {
			d[nCards*i+j] = NewCard(Suit(i), j+1)
		}
	}

	return shuffle(d)
}

func shuffle(d Deck) Deck {
	rand.Shuffle(len(d), func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})

	return d
}

func suitToUnicode(s Suit) string {
	switch s {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Clubs:
		return "♦"
	case Diamonds:
		return "♣"
	default:
		panic("invalid card suit")
	}
}