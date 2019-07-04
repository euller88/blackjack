package blackjack

import (
	"fmt"

	"github.com/euller88/deck"
)

// state represents the curret state of a blackjack game
type state int8

const (
	statePlayerTurn state = iota // statePlayerTurn signals that is the player's turn
	stateDealerTurn              // stateDealerTurn signals that is the dealer's turn
	stateHandOver                // stateHandOver signals that the game has ended
)

// Game is collection of information that a game knows about itself at any given time
type Game struct {
	// unexported fields
	deck     []deck.Card
	state    state
	player   []deck.Card
	dealer   []deck.Card
	dealerAi Player
	balance  int
}

// New creates a new game
func New() Game {
	return Game{
		dealerAi: dealer{},
	}
}

// Move represents the possible actions a player can do
type Move func(*Game)

// MoveHit draws a card from the deck and give it to the current player
func MoveHit(gs *Game) {
	hand := gs.currentPlayerHand()
	var card deck.Card
	card, gs.deck = draw(gs.deck)
	*hand = append(*hand, card)
	if Score(*hand...) > 21 {
		MoveStand(gs)
	}
}

// MoveStand signals to the game state that the current player is not accepting any new cards
func MoveStand(gs *Game) {
	gs.state++
}

func draw(cards []deck.Card) (deck.Card, []deck.Card) {
	return cards[0], cards[1:]
}

// Play is the core loop of execution of a game
func (gs *Game) Play(pl Player) int {
	gs.deck = deck.New(deck.Deck(3), deck.Shuffle)

	for i := 0; i < 10; i++ {
		deal(gs)

		for gs.state == statePlayerTurn {
			hand := make([]deck.Card, len(gs.player))
			copy(hand, gs.player)
			move := pl.Play(hand, gs.dealer[0])
			move(gs)
		}

		for gs.state == stateDealerTurn {
			hand := make([]deck.Card, len(gs.dealer))
			copy(hand, gs.dealer)
			move := gs.dealerAi.Play(hand, deck.Card{Suit: deck.Joker})
			move(gs)
		}

		endHand(gs, pl)
	}

	return 0
}

func deal(gs *Game) {
	gs.player = make([]deck.Card, 0, 5)
	gs.dealer = make([]deck.Card, 0, 5)

	var card deck.Card

	for i := 0; i < 2; i++ {
		card, gs.deck = draw(gs.deck)
		gs.player = append(gs.player, card)

		card, gs.deck = draw(gs.deck)
		gs.dealer = append(gs.dealer, card)
	}

	gs.state = statePlayerTurn
}

// Score returns and change the value of any ace to 11 if is needed
func Score(h ...deck.Card) int {
	m := minScore(h...)

	if m > 11 {
		return m
	}

	for _, c := range h {
		if c.Rank == deck.Ace {
			return m + 10
		}
	}

	return m
}

// Soft indicates if the current set of cards has that value because an Ace is being counted as its value is 11
func Soft(h ...deck.Card) bool {
	m, s := minScore(h...), Score(h...)
	return m != s
}

// MinScore determines how much a is worth in points if all aces were considered to be worth 1
func minScore(h ...deck.Card) int {
	score := 0
	for _, card := range h {
		score += min(int(card.Rank), 10)
	}
	return score
}

// min is just a minimum function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func endHand(gs *Game, pl Player) {
	pScore, dScore := Score(gs.player...), Score(gs.dealer...)
	switch {
	case pScore > 21:
		fmt.Print("You busted\n\n")
		gs.balance--
	case dScore > 21:
		fmt.Print("Dealer busted\n\n")
		gs.balance++
	case pScore > dScore:
		fmt.Print("You win!\n\n")
		gs.balance++
	case dScore > pScore:
		fmt.Print("You lose\n\n")
		gs.balance--
	case dScore == pScore:
		fmt.Print("Draw\n\n")
	}

	pl.Summary([][]deck.Card{gs.player}, gs.dealer)

	gs.player = nil
	gs.dealer = nil
}

func (gs *Game) currentPlayerHand() *[]deck.Card {
	switch gs.state {
	case statePlayerTurn:
		return &gs.player
	case stateDealerTurn:
		return &gs.dealer
	default:
		panic("it isn't a player turn")
	}
}
