package game

import "pokerclientv1/internal/types"

// Table represents the shared state of the poker table.
type Table struct {
	CommunityCards []types.Card
	Pot            int
	CurrentBet     int    // The highest bet amount in the current round
	Round          string // e.g., "Pre-flop", "Flop", "Turn", "River"
}

// NewTable creates a new, empty table.
func NewTable() *Table {
	return &Table{
		CommunityCards: make([]types.Card, 0, 5),
		Pot:            0,
		CurrentBet:     0,
	}
}

// AddCommunityCard adds a card to the community cards.
func (t *Table) AddCommunityCard(card types.Card) {
	t.CommunityCards = append(t.CommunityCards, card)
}

// AddToPot adds an amount to the pot.
func (t *Table) AddToPot(amount int) {
	t.Pot += amount
}

// ResetForNewRound resets the table state for a new betting round (e.g., after flop).
func (t *Table) ResetForNewRound() {
	t.CurrentBet = 0
	// Pot is not reset between betting rounds within the same hand
}

// ResetForNewHand resets the table state for a completely new hand.
func (t *Table) ResetForNewHand() {
	t.CommunityCards = make([]types.Card, 0, 5)
	t.Pot = 0
	t.CurrentBet = 0
	t.Round = ""
}
