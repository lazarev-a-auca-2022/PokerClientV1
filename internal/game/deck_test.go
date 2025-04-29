package game

import (
	"pokerclientv1/internal/types"
	"testing"
)

// TestNewDeck checks if a new deck has the correct number of cards (52).
func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	expectedDeckSize := 52
	if len(deck.cards) != expectedDeckSize {
		t.Errorf("NewDeck() created a deck with %d cards, want %d", len(deck.cards), expectedDeckSize)
	}

	// Optional: Check for uniqueness (no duplicate cards)
	seenCards := make(map[types.Card]bool)
	for _, card := range deck.cards {
		if seenCards[card] {
			t.Errorf("NewDeck() created a deck with duplicate card: %s", card.String())
		}
		seenCards[card] = true
	}
}

// TestShuffle checks if the deck is actually shuffled (order changes).
func TestShuffle(t *testing.T) {
	deck1 := NewDeck()
	deck2 := NewDeck()

	deck2.Shuffle() // Shuffle the second deck

	if len(deck1.cards) != len(deck2.cards) {
		t.Fatalf("Shuffle() changed the number of cards. Before: %d, After: %d", len(deck1.cards), len(deck2.cards))
	}

	// Check if the order is different. It's statistically highly unlikely
	// that a shuffled deck has the exact same order as a new one.
	sameOrder := true
	for i := range deck1.cards {
		if deck1.cards[i] != deck2.cards[i] {
			sameOrder = false
			break
		}
	}
	if sameOrder {
		t.Errorf("Shuffle() did not change the order of the cards. This is statistically improbable.")
	}

	// Optional: Check if all original cards are still present after shuffle
	seenCards := make(map[types.Card]int)
	for _, card := range deck1.cards {
		seenCards[card]++
	}
	for _, card := range deck2.cards {
		seenCards[card]--
	}
	for card, count := range seenCards {
		if count != 0 {
			t.Errorf("Shuffle() resulted in missing or extra card: %s (count difference: %d)", card.String(), count)
		}
	}
}

// TestDeal checks dealing a single card.
func TestDeal(t *testing.T) {
	deck := NewDeck()
	initialSize := len(deck.cards)

	card, err := deck.Deal()
	if err != nil {
		t.Fatalf("Deal() returned an unexpected error: %v", err)
	}

	if len(deck.cards) != initialSize-1 {
		t.Errorf("Deal() did not reduce deck size correctly. Got %d, want %d", len(deck.cards), initialSize-1)
	}

	// Check if the dealt card is no longer in the deck
	found := false
	for _, remainingCard := range deck.cards {
		if remainingCard == card {
			found = true
			break
		}
	}
	if found {
		t.Errorf("Deal() did not remove the dealt card (%s) from the deck", card.String())
	}

	// Test dealing from an empty deck
	deck.cards = []types.Card{} // Empty the deck
	_, err = deck.Deal()
	if err == nil {
		t.Errorf("Deal() from empty deck did not return an error")
	}
}

// TestDealMultiple checks dealing multiple cards.
func TestDealMultiple(t *testing.T) {
	deck := NewDeck()
	initialSize := len(deck.cards)
	numToDeal := 5

	cards, err := deck.DealMultiple(numToDeal)
	if err != nil {
		t.Fatalf("DealMultiple(%d) returned an unexpected error: %v", numToDeal, err)
	}

	if len(cards) != numToDeal {
		t.Errorf("DealMultiple(%d) dealt %d cards, want %d", numToDeal, len(cards), numToDeal)
	}

	if len(deck.cards) != initialSize-numToDeal {
		t.Errorf("DealMultiple(%d) did not reduce deck size correctly. Got %d, want %d", numToDeal, len(deck.cards), initialSize-numToDeal)
	}

	// Test dealing more cards than available
	deck = NewDeck()
	numToDeal = 53
	_, err = deck.DealMultiple(numToDeal)
	if err == nil {
		t.Errorf("DealMultiple(%d) with insufficient cards did not return an error", numToDeal)
	}
}
