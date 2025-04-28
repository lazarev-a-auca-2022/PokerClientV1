package player

import (
	"fmt"
	"pokerclientv1/internal/types"
	"time"
)

// BotPlayer rfunc (p *BotPlayer) TakeTurn(table *types.Table, currentBet int, minRaise int) (action string, amount int)presents an AI-controlled player.
type BotPlayer struct {
	ID         string
	Chips      int
	Hand       *types.Hand
	AI         *BotAI
	Folded     bool
	CurrentBet int // Amount bet in the current round
}

// NewBotPlayer creates a new bot player with specified AI settings.
func NewBotPlayer(id string, startingChips int, difficulty string, turnDelay time.Duration) *BotPlayer {
	return &BotPlayer{
		ID:    id,
		Chips: startingChips,
		Hand:  &types.Hand{},
		AI: &BotAI{
			Difficulty: difficulty,
			TurnDelay:  turnDelay,
		},
		Folded:     false,
		CurrentBet: 0,
	}
}

func (p *BotPlayer) GetID() string {
	return p.ID
}

func (p *BotPlayer) GetHand() *types.Hand {
	return p.Hand
}

func (p *BotPlayer) SetHand(hand *types.Hand) {
	p.Hand = hand
}

func (p *BotPlayer) AddChips(amount int) {
	p.Chips += amount
}

func (p *BotPlayer) RemoveChips(amount int) error {
	if amount > p.Chips {
		// Allow removing all chips if going all-in
		// The game engine should handle the case where a player bets more than they have.
		// For now, return an error, but this might need adjustment based on game rules.
		// Alternatively, just remove all chips.
		// Let's allow removing up to p.Chips
		actualAmount := p.Chips
		p.Chips = 0
		return fmt.Errorf("%s cannot remove %d chips, only had %d. Removed %d (all-in)", p.ID, amount, actualAmount, actualAmount)
		// return fmt.Errorf("%s cannot remove %d chips, only has %d", p.ID, amount, p.Chips)
	}
	p.Chips -= amount
	return nil
}

func (p *BotPlayer) GetChips() int {
	return p.Chips
}

func (p *BotPlayer) IsFolded() bool {
	return p.Folded
}

func (p *BotPlayer) SetFolded(folded bool) {
	p.Folded = folded
}

func (p *BotPlayer) GetCurrentBet() int {
	return p.CurrentBet
}

func (p *BotPlayer) SetCurrentBet(amount int) {
	p.CurrentBet = amount
}

func (p *BotPlayer) ResetBet() {
	p.CurrentBet = 0
}

func (p *BotPlayer) ResetForNewHand() {
	p.Hand = &types.Hand{}
	p.Folded = false
	p.CurrentBet = 0
	// Chips carry over
}

// TakeTurn uses the BotAI to decide the action.
func (p *BotPlayer) TakeTurn(table *types.Table, currentBet int, minRaise int) (action string, amount int) {
	// The amount returned by DecideAction is the TOTAL bet for the round.
	// We need to calculate the amount to ADD to the pot.
	callAmount := currentBet - p.CurrentBet
	action, totalBetAmount := p.AI.DecideAction(p.Hand, table, currentBet, p.Chips, minRaise)

	// Adjust the amount based on the action type
	amountToAdd := 0
	switch action {
	case "fold":
		amountToAdd = 0
	case "check": // BotAI currently doesn't return check, but handle for future
		amountToAdd = 0
	case "call":
		// If bot decides to call, the amount should be the difference needed
		amountToAdd = callAmount
		// Handle all-in call (if bot doesn't have enough to cover the full call)
		if amountToAdd > p.Chips {
			amountToAdd = p.Chips
		}
	case "raise":
		// DecideAction returns the total bet amount for the round when raising.
		// Calculate the amount to add to the pot.
		amountToAdd = totalBetAmount - p.CurrentBet
		// Handle all-in raise
		if amountToAdd > p.Chips {
			amountToAdd = p.Chips
			// If going all-in results in a bet less than or equal to the current bet, it's a call.
			if p.CurrentBet+amountToAdd <= currentBet {
				action = "call"
			}
		}
	default:
		action = "fold"
		amountToAdd = 0
	}

	// Ensure bot doesn't bet more chips than it has
	if amountToAdd < 0 {
		// This shouldn't happen with correct logic, but as a safeguard
		fmt.Printf("Warning: Bot %s attempted to bet negative amount (%d). Folding.\n", p.ID, amountToAdd)
		action = "fold"
		amountToAdd = 0
	} else if amountToAdd > p.Chips {
		fmt.Printf("Warning: Bot %s attempting to bet %d but only has %d. Going all-in.\n", p.ID, amountToAdd, p.Chips)
		amountToAdd = p.Chips
		// Re-evaluate if it's a call or raise when going all-in
		if p.CurrentBet+amountToAdd > currentBet {
			action = "raise"
		} else {
			action = "call"
		}
	}

	return action, amountToAdd
}
