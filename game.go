package blackjack

import (
	"errors"

	"github.com/euller88/deck"
)

// state represents the curret state of a blackjack game
type state int8

const (
	statePlayerTurn state = iota // statePlayerTurn signals that is the player's turn
	stateDealerTurn              // stateDealerTurn signals that is the dealer's turn
	stateHandOver                // stateHandOver signals that the game has ended
)

// Options represents the options an user can give before the start of a game
type Options struct {
	Hands           int     // the number of hands that will be played, at least 1, default 100
	Decks           int     // the number of decks that will be used, at least 1, default 3
	BlackJackPayout float64 // the payout for achieving a natural 21, AKA blackjack, greater than 1, default 1.5
}

// Game is collection of information that a game knows about itself at any given time
type Game struct {
	// unexported fields
	opts Options

	deck  []deck.Card
	state state

	player    []hand
	handIdx   int
	playerBet int
	balance   int

	dealer   []deck.Card
	dealerAi Player
}

// New creates a new game
func New(opts Options) Game {
	if opts.Hands <= 0 {
		opts.Hands = 100
	}
	if opts.Decks <= 0 {
		opts.Decks = 3
	}
	if opts.BlackJackPayout < 1.0 {
		opts.BlackJackPayout = 1.5
	}

	return Game{
		dealerAi: dealer{},
		opts:     opts,
	}
}

type hand struct {
	cards []deck.Card
	bet   int
}

// Move represents the possible actions a player can do
type Move func(*Game) error

// MoveHit is the action when a player hits
func MoveHit(g *Game) error {
	hand := g.currentPlayerHand()
	var card deck.Card
	card, g.deck = draw(g.deck)
	*hand = append(*hand, card)
	if Score(*hand...) > 21 {
		return errBust
	}
	return nil
}

// MoveSplit is the action when a player splits
func MoveSplit(g *Game) error {
	cards := g.currentPlayerHand()
	if len(*cards) != 2 {
		return errors.New("you can only split with two cards in your hand")
	}
	if (*cards)[0].Rank != (*cards)[1].Rank {
		return errors.New("both cards must have the same rank to split")
	}
	g.player = append(g.player, hand{
		cards: []deck.Card{(*cards)[1]},
		bet:   g.player[g.handIdx].bet,
	})
	g.player[g.handIdx].cards = (*cards)[:1]
	return nil
}

// MoveDouble is the action when a player doubles their bet
func MoveDouble(g *Game) error {
	if len(g.player) != 2 {
		return errors.New("can only double on a hand with 2 cards")
	}
	g.playerBet *= 2
	MoveHit(g)
	return MoveStand(g)
}

// MoveStand is the action when a player stands
func MoveStand(g *Game) error {
	if g.state == stateDealerTurn {
		g.state++
		return nil
	}
	if g.state == statePlayerTurn {
		g.handIdx++
		if g.handIdx >= len(g.player) {
			g.state++
		}
		return nil
	}
	return errors.New("invalid state")
}

func draw(cards []deck.Card) (deck.Card, []deck.Card) {
	return cards[0], cards[1:]
}

func bet(gs *Game, pl Player, shuffled bool) {
	bet := pl.Bet(shuffled)
	gs.playerBet = bet
}

// Play is the core loop of execution of a game
func (g *Game) Play(pl Player) int {
	g.deck = nil
	min := 52 * g.opts.Decks / 3
	for i := 0; i < g.opts.Hands; i++ {
		shuffled := false
		if len(g.deck) < min {
			g.deck = deck.New(deck.Deck(g.opts.Decks), deck.Shuffle)
			shuffled = true
		}
		bet(g, pl, shuffled)
		deal(g)
		if Blackjack(g.dealer...) {
			endRound(g, pl)
			continue
		}
		for g.state == statePlayerTurn {
			hand := make([]deck.Card, len(*g.currentPlayerHand()))
			copy(hand, *g.currentPlayerHand())
			move := pl.Play(hand, g.dealer[0])
			err := move(g)
			switch err {
			case errBust:
				MoveStand(g)
			case nil:
				// noop
			default:
				panic(err)
			}
		}
		for g.state == stateDealerTurn {
			hand := make([]deck.Card, len(g.dealer))
			copy(hand, g.dealer)
			move := g.dealerAi.Play(hand, deck.Card{Suit: deck.Joker})
			move(g)
		}
		endRound(g, pl)
	}
	return g.balance
}

// Options returns the current options of a game
func (g Game) Options() Options {
	return g.opts
}

func deal(g *Game) {
	playerHand := make([]deck.Card, 0, 5)
	g.dealer = make([]deck.Card, 0, 5)
	var card deck.Card
	for i := 0; i < 2; i++ {
		card, g.deck = draw(g.deck)
		playerHand = append(playerHand, card)
		card, g.deck = draw(g.deck)
		g.dealer = append(g.dealer, card)
	}
	playerHand = []deck.Card{
		{Rank: deck.Seven},
		{Rank: deck.Seven},
	}
	g.player = []hand{
		{
			cards: playerHand,
			bet:   g.playerBet,
		},
	}
	g.state = statePlayerTurn
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

func endRound(g *Game, pl Player) {
	dScore := Score(g.dealer...)
	dBlackjack := Blackjack(g.dealer...)
	allHands := make([][]deck.Card, len(g.player))
	for hi, hand := range g.player {
		cards := hand.cards
		allHands[hi] = cards
		pScore, pBlackjack := Score(cards...), Blackjack(cards...)
		winnings := hand.bet
		switch {
		case pBlackjack && dBlackjack:
			winnings = 0
		case dBlackjack:
			winnings = -winnings
		case pBlackjack:
			winnings = int(float64(winnings) * g.opts.BlackJackPayout)
		case pScore > 21:
			winnings = -winnings
		case dScore > 21:
			// win
		case pScore > dScore:
			// win
		case dScore > pScore:
			winnings = -winnings
		case dScore == pScore:
			winnings = 0
		}
		g.balance += winnings
	}
	pl.Summary(allHands, g.dealer)
	g.player = nil
	g.dealer = nil
}

func (g *Game) currentPlayerHand() *[]deck.Card {
	switch g.state {
	case statePlayerTurn:
		return &g.player[g.handIdx].cards
	case stateDealerTurn:
		return &g.dealer
	default:
		panic("it isn't currently any player's turn")
	}
}

// Blackjack returns true if a hand is a blackjack
func Blackjack(hand ...deck.Card) bool {
	return len(hand) == 2 && Score(hand...) == 21
}

var (
	errBust = errors.New("hand score exceeded 21")
)
