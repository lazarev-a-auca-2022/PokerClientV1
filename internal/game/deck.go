package game

import (
	"errors"
	"math/rand"
	"time"

	"pokerclientv1/internal/types"
)

// Deck represents a deck of playing cards
type Deck struct {
	cards []types.Card
}

// NewDeck creates and returns a new deck of 52 cards
func NewDeck() *Deck {
	deck := &Deck{
		cards: make([]types.Card, 0, 52),
	}

	// Create all combinations of suits and ranks
	for suit := types.Spade; suit <= types.Club; suit++ {
		for rank := types.Two; rank <= types.Ace; rank++ {
			deck.cards = append(deck.cards, types.Card{
				Suit: suit,
				Rank: rank,
			})
		}
	}

	return deck
}

// Shuffle randomizes the order of cards in the deck
func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

// Deal removes and returns the top card from the deck
func (d *Deck) Deal() (types.Card, error) {
	if len(d.cards) == 0 {
		return types.Card{}, errors.New("no cards left in deck")
	}

	card := d.cards[len(d.cards)-1]
	d.cards = d.cards[:len(d.cards)-1]
	return card, nil
}

// DealMultiple deals multiple cards from the deck
func (d *Deck) DealMultiple(numCards int) ([]types.Card, error) {
	if len(d.cards) < numCards {
		return nil, errors.New("not enough cards left in deck")
	}

	cards := make([]types.Card, numCards)
	for i := 0; i < numCards; i++ {
		card, err := d.Deal()
		if err != nil {
			return nil, err
		}
		cards[i] = card
	}
	return cards, nil
}

// CardsLeft returns the number of cards remaining in the deck
func (d *Deck) CardsLeft() int {
	return len(d.cards)
}

// Reset resets the deck to a full 52-card deck
func (d *Deck) Reset() {
	*d = *NewDeck()
}
