package player

import (
	"bufio"
	"fmt"
	"os"
	"pokerclientv1/internal/types"
	"strconv"
	"strings"
)

// HumanPlayer represents a player controlled by user input.
type HumanPlayer struct {
	ID         string
	Chips      int
	Hand       *types.Hand
	Folded     bool
	CurrentBet int // Amount bet in the current round
}

// NewHumanPlayer creates a new human player.
func NewHumanPlayer(id string, startingChips int) *HumanPlayer {
	return &HumanPlayer{
		ID:         id,
		Chips:      startingChips,
		Hand:       &types.Hand{},
		Folded:     false,
		CurrentBet: 0,
	}
}

// Implement all the methods required by the types.Player interface
// Most method implementations remain the same, just update any type references to use types.Hand, types.Table, etc.

func (p *HumanPlayer) GetID() string            { return p.ID }
func (p *HumanPlayer) GetHand() *types.Hand     { return p.Hand }
func (p *HumanPlayer) SetHand(hand *types.Hand) { p.Hand = hand }
func (p *HumanPlayer) AddChips(amount int)      { p.Chips += amount }
func (p *HumanPlayer) GetChips() int            { return p.Chips }
func (p *HumanPlayer) IsFolded() bool           { return p.Folded }
func (p *HumanPlayer) SetFolded(folded bool)    { p.Folded = folded }
func (p *HumanPlayer) GetCurrentBet() int       { return p.CurrentBet }
func (p *HumanPlayer) SetCurrentBet(amount int) { p.CurrentBet = amount }
func (p *HumanPlayer) ResetBet()                { p.CurrentBet = 0 }

// IsHuman returns true for HumanPlayer
func (p *HumanPlayer) IsHuman() bool { return true }

func (p *HumanPlayer) RemoveChips(amount int) error {
	if amount > p.Chips {
		return fmt.Errorf("%s cannot remove %d chips, only has %d", p.ID, amount, p.Chips)
	}
	p.Chips -= amount
	return nil
}

func (p *HumanPlayer) ResetForNewHand() {
	p.Hand = &types.Hand{}
	p.Folded = false
	p.CurrentBet = 0
}

// TakeTurn prompts the human player for their action via the console.
func (p *HumanPlayer) TakeTurn(table *types.Table, currentBet int, minRaise int) (action string, amount int) {
	reader := bufio.NewReader(os.Stdin)
	callAmount := currentBet - p.CurrentBet // Amount needed to call

	for {
		fmt.Printf("%s's turn (Chips: %d, Current Bet: %d). Hand: %s\n", p.ID, p.Chips, p.CurrentBet, p.Hand)
		fmt.Printf("Community Cards: %v | Current High Bet: %d\n", table.CommunityCards, currentBet)

		options := []string{"fold"}
		if p.Chips >= callAmount {
			if callAmount == 0 {
				options = append(options, "check")
			} else {
				options = append(options, fmt.Sprintf("call (%d)", callAmount))
			}
		}
		// Can only raise if they can at least match the current bet and raise by minRaise, or go all-in
		canAffordMinRaise := p.Chips >= callAmount+minRaise
		canGoAllIn := p.Chips > callAmount // Must have more chips than needed to call to raise/go all-in
		if canAffordMinRaise {
			options = append(options, "raise")
		}
		if canGoAllIn {
			options = append(options, "all-in")
		} else if callAmount > 0 && p.Chips < callAmount {
			// If cannot afford call, only option is fold or all-in (which acts as a call here)
			options = []string{"fold", fmt.Sprintf("all-in (%d)", p.Chips)}
		}

		fmt.Printf("Options: [%s]\n", strings.Join(options, ", "))
		fmt.Print("Enter action: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		parts := strings.Fields(input) // Split input by space
		actionCmd := parts[0]

		switch actionCmd {
		case "fold":
			return "fold", 0
		case "check":
			if callAmount == 0 {
				return "check", 0
			}
			fmt.Println("Invalid action: Cannot check, there is a bet to call.")
		case "call":
			if callAmount == 0 {
				fmt.Println("Invalid action: Cannot call, you can check.")
				continue
			}
			if p.Chips >= callAmount {
				return "call", callAmount // Return the amount needed *to add* to the pot
			}
			// If not enough chips to call the full amount, they go all-in
			fmt.Printf("Not enough chips to call %d. Going all-in with %d.\n", callAmount, p.Chips)
			return "call", p.Chips // Go all-in (effectively a call for their remaining chips)
		case "raise":
			if !canGoAllIn {
				fmt.Println("Invalid action: Cannot raise.")
				continue
			}

			var raiseAmount int
			if len(parts) > 1 {
				parsedAmount, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Println("Invalid raise amount. Please enter a number (e.g., 'raise 50').")
					continue
				}
				// The amount entered is the TOTAL amount the player wants to bet in this round
				raiseAmount = parsedAmount
			} else {
				// Ask for amount if not provided
				fmt.Printf("Enter total raise amount (min %d, max %d): ", currentBet+minRaise, p.CurrentBet+p.Chips)
				amountInput, _ := reader.ReadString('\n')
				parsedAmount, err := strconv.Atoi(strings.TrimSpace(amountInput))
				if err != nil {
					fmt.Println("Invalid amount.")
					continue
				}
				raiseAmount = parsedAmount
			} // Validate raise amount
			actualRaise := raiseAmount - currentBet        // The amount *above* the current bet
			totalBetRequired := raiseAmount - p.CurrentBet // Amount to add to pot

			if totalBetRequired > p.Chips {
				fmt.Printf("Invalid raise: You only have %d chips (need %d).\n", p.Chips, totalBetRequired)
				continue
			}
			// Validate minimum raise amount, but allow smaller raises if going all-in
			if actualRaise < minRaise && p.Chips > totalBetRequired {
				fmt.Printf("Invalid raise: Minimum raise amount is %d.\n", minRaise)
				continue
			}
			if raiseAmount <= currentBet {
				fmt.Printf("Invalid raise: Must raise higher than the current bet of %d.\n", currentBet)
				continue
			}

			return "raise", totalBetRequired // Return the amount to *add* to the pot

		case "all-in":
			if !canGoAllIn && !(callAmount > 0 && p.Chips < callAmount) { // Allow all-in if cannot afford call
				fmt.Println("Invalid action: Cannot go all-in.")
				continue
			}
			allInAmount := p.Chips // The amount to add to the pot is all remaining chips
			actionType := "call"   // Default to call if all-in amount is less than or equal to call amount
			if p.CurrentBet+allInAmount > currentBet {
				actionType = "raise" // It's a raise if the total bet exceeds the current highest bet
			}
			fmt.Printf("Going all-in with %d chips.\n", allInAmount)
			return actionType, allInAmount // Return "raise" or "call" depending on context, and the amount added

		default:
			fmt.Println("Invalid action. Please choose from the available options.")
		}
	}
}
