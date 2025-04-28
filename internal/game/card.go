package game

import "fmt"

// Suit represents the suit of a card.
type Suit int

const (
	Spade Suit = iota
	Heart
	Diamond
	Club
)

func (s Suit) String() string {
	return [...]string{"Spades", "Hearts", "Diamonds", "Clubs"}[s]
}

// Rank represents the rank of a card.
type Rank int

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
	Ace // Ace can be high or low, handle in evaluation
)

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
