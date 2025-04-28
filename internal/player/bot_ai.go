// filepath: f:\PokerClientV1\internal\player\bot_ai.go
package player

import (
	"math/rand"
	"pokerclientv1/internal/types"
	"time"
)

// BotAI defines the structure for bot decision logic.
type BotAI struct {
	Difficulty string        // easy, medium, hard
	TurnDelay  time.Duration // How long the bot "thinks" before acting
}

// DecideAction determines the bot's action based on its AI settings.
func (ai *BotAI) DecideAction(hand *types.Hand, table *types.Table, currentBet int, chips int, minRaise int) (action string, amount int) {
	time.Sleep(ai.TurnDelay) // Simulate thinking

	// Current call amount
	callAmount := currentBet

	// Simple random strategy based on difficulty
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch ai.Difficulty {
	case "easy":
		// Easy bot: 60% call, 20% fold, 20% raise (small)
		decision := r.Intn(100)

		if decision < 20 {
			return "fold", 0
		} else if decision < 80 {
			// Call if possible
			if callAmount >= chips {
				return "call", chips // All-in call
			}
			return "call", callAmount
		} else {
			// Small raise between 1-2x min raise
			raiseMultiplier := 1.0 + r.Float64()
			raiseAmount := int(float64(minRaise) * raiseMultiplier)
			totalBet := currentBet + raiseAmount

			if totalBet >= chips {
				return "raise", chips // All-in raise
			}
			return "raise", totalBet
		}

	case "medium":
		// Medium bot: More strategic decisions
		decision := r.Intn(100)

		if decision < 15 {
			return "fold", 0
		} else if decision < 70 {
			if callAmount >= chips {
				return "call", chips // All-in call
			}
			return "call", callAmount
		} else {
			// Medium raises between 1-3x min raise
			raiseMultiplier := 1.0 + 2.0*r.Float64()
			raiseAmount := int(float64(minRaise) * raiseMultiplier)
			totalBet := currentBet + raiseAmount

			if totalBet >= chips {
				return "raise", chips // All-in raise
			}
			return "raise", totalBet
		}

	case "hard":
		// Hard bot: Much more aggressive
		decision := r.Intn(100)

		if decision < 10 {
			return "fold", 0
		} else if decision < 50 {
			if callAmount >= chips {
				return "call", chips // All-in call
			}
			return "call", callAmount
		} else {
			// Larger raises between 2-4x min raise
			raiseMultiplier := 2.0 + 2.0*r.Float64()
			raiseAmount := int(float64(minRaise) * raiseMultiplier)
			totalBet := currentBet + raiseAmount

			if totalBet >= chips {
				return "raise", chips // All-in raise
			}
			return "raise", totalBet
		}

	default: // Default to simple logic
		actionOptions := []string{"fold", "call", "raise"}
		chosenAction := actionOptions[r.Intn(len(actionOptions))]

		switch chosenAction {
		case "fold":
			return "fold", 0
		case "call":
			if currentBet > chips {
				return "call", chips // All-in
			}
			return "call", currentBet
		case "raise":
			// Basic raise logic
			raiseAmount := currentBet + minRaise
			if raiseAmount > chips {
				return "raise", chips // All-in
			}
			return "raise", raiseAmount
		default:
			return "fold", 0
		}
	}
}
