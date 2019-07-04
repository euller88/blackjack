package blackjack

import (
	"fmt"

	"github.com/euller88/deck"
)

// Player is an abstraction of what a player in a blackjack game may be, a human inputing commands or an AI executing tasks
type Player interface {
	Bet() int
	Play(hand []deck.Card, dealer deck.Card) Move
	Summary(hand [][]deck.Card, dealer []deck.Card)
}

// HumanPlayer returns a Player "object" ready to receive human inputs
func HumanPlayer() Player {
	return humanPlayer{}
}

// humanPlayer is the implementation of the Player interface to receive interactions from a real person, whatever it may be
type humanPlayer struct{}

// Bet returns whatever the player whats to put at stake in the game
func (pl humanPlayer) Bet() int {
	return 1
}

// Play is the core function of that takes an input and decides what kind of move a player should make
func (pl humanPlayer) Play(hand []deck.Card, dealer deck.Card) Move {
	var input string
	for {
		fmt.Println("Player:", hand)
		fmt.Println("Dealer:", dealer)
		fmt.Println("What will you do? (h)it, (s)tand")
		fmt.Scanf("%s\n", &input)

		switch input {
		case "h":
			return MoveHit
		case "s":
			return MoveStand
		default:
			fmt.Println("Invalid option: ", input)
		}
	}
}

// Summary gives the final state of a player
func (pl humanPlayer) Summary(hand [][]deck.Card, dealer []deck.Card) {
	fmt.Println("==FINAL HANDS==")
	fmt.Println("Player:", hand)
	fmt.Println("Dealer:", dealer)
}

// Dealer is the implementation of the Player interface that represents a dealer in a blackjack game
type dealer struct{}

// Bet is the implementation of the Bet method of the Player interface, this actually does nothing
func (dl dealer) Bet() int {
	// noop
	return 0
}

// Play is the implementation of the Play method of the Player interface
func (dl dealer) Play(hand []deck.Card, dealer deck.Card) Move {
	ds := Score(hand...)
	if ds <= 16 || (ds == 17 && Soft(hand...)) {
		return MoveHit
	}
	return MoveStand
}

// Summary is the implementation of the Summary method of the Player interface, this actually does nothing
func (dl dealer) Summary(hand [][]deck.Card, dealer []deck.Card) {
	// noop
}
