package types

import (
	"fmt"
	"strings"
)

// GameUI defines the interface for game display and logging
type GameUI interface {
	DisplayGameState(table *Table, players []Player, pot int, stage string)
	LogAction(playerID string, action string, amount int)
}

// Player defines the interface for any player (human or bot)
type Player interface {
	GetID() string
	GetHand() *Hand
	TakeTurn(table *Table, currentBet int, minRaise int) (action string, amount int)
	AddChips(amount int)
	RemoveChips(amount int) error
	GetChips() int
	SetHand(hand *Hand)
	IsFolded() bool
	SetFolded(folded bool)
	ResetForNewHand()
	GetCurrentBet() int
	SetCurrentBet(amount int)
	ResetBet()
}

// Table represents the shared state of the poker table
type Table struct {
	CommunityCards []Card
	CurrentBet     int
	Round          string
}

// Hand represents a player's hand of cards
type Hand struct {
	Cards []Card
}

// Card represents a playing card
type Card struct {
	Suit Suit
	Rank Rank
}

// Suit represents the suit of a card
type Suit int

// Rank represents the rank of a card
type Rank int

// Constants for Suit
const (
	Spade Suit = iota
	Heart
	Diamond
	Club
)

// Constants for Rank
const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

// String methods for Card, Suit, and Rank
func (s Suit) String() string {
	// Use Unicode symbols for suits
	return [...]string{"♠", "♥", "♦", "♣"}[s]
}

func (r Rank) String() string {
	switch r {
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ace:
		return "A"
	default:
		return fmt.Sprintf("%d", r)
	}
}

func (c Card) String() string {
	// Combine rank and Unicode suit symbol
	return fmt.Sprintf("%s%s", c.Rank.String(), c.Suit.String()) // Corrected: use c.Suit
}

func (h *Hand) String() string {
	if h == nil || len(h.Cards) == 0 {
		return "[ ]"
	}
	cards := make([]string, len(h.Cards))
	for i, card := range h.Cards {
		cards[i] = card.String()
	}
	return fmt.Sprintf("[ %s ]", strings.Join(cards, " "))
}

func (h *Hand) AddCard(card Card) {
	if h.Cards == nil {
		h.Cards = make([]Card, 0, 2)
	}
	h.Cards = append(h.Cards, card)
}

func (t *Table) ResetForNewHand() {
	t.CommunityCards = make([]Card, 0, 5)
	t.CurrentBet = 0
	t.Round = ""
}

func (t *Table) AddCommunityCard(card Card) {
	t.CommunityCards = append(t.CommunityCards, card)
}
