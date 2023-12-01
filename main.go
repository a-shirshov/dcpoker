package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/a-shirshov/dcpoker/deck"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	d := deck.New()
	fmt.Println(d)
	fmt.Println()
}
