package game

import "pokerclientv1/internal/types"

// Hand represents a player's hand of cards.
type Hand struct {
	Cards []types.Card
}

// AddCard adds a card to the hand.
func (h *Hand) AddCard(card types.Card) {
	h.Cards = append(h.Cards, card)
}

// String returns a string representation of the hand.
func (h *Hand) String() string {
	s := ""
	for i, card := range h.Cards {
		s += card.String()
		if i < len(h.Cards)-1 {
			s += " "
		}
	}
	return s
}

// TODO: Implement hand evaluation logic (e.g., GetStrength, Compare)
// This will involve determining the best poker hand (pair, flush, straight, etc.)
// from the player's cards and any community cards.
