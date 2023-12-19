package deck

import (
	"fmt"
	"testing"
)

func TestEncryptCard(t *testing.T) {
	key := []byte("foobarbazfoobarbazfoobarbazfoobarbazfoobarbaz")
	card := Card{
		Suit: Spades,
		Value: 1,
	}

	encOutput, err := EncryptCard(key, card)
	if err != nil {
		t.Error(err)
	}

	decCard, err := DecryptCard(key, encOutput)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%b\n", encOutput)
	fmt.Printf("%+v\n", decCard)
}