package ui

import (
	"fmt"
	"pokerclientv1/internal/types"
	"strings"
)

// ConsoleUI implements types.GameUI for console-based display
type ConsoleUI struct{}

// NewConsoleUI creates a new console UI instance
func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{}
}

// DisplayGameState prints the current state of the game to the console.
func (ui *ConsoleUI) DisplayGameState(table *types.Table, players []types.Player, pot int, stage string) {
	fmt.Println("\n==================================================")
	fmt.Printf("--- %s --- Pot: %d ---\n", stage, pot)

	// Display Community Cards
	if len(table.CommunityCards) > 0 {
		cardsStr := []string{}
		for _, card := range table.CommunityCards {
			cardsStr = append(cardsStr, card.String())
		}
		fmt.Printf("Community Cards: [ %s ]\n", strings.Join(cardsStr, " "))
	} else {
		fmt.Println("Community Cards: [ ]")
	}

	fmt.Println("--- Players ---")
	for _, p := range players {
		status := ""
		if p.IsFolded() {
			status = " (Folded)"
		} else if p.GetChips() == 0 && p.GetCurrentBet() > 0 {
			status = " (All-In)"
		}
		// Don't show bot hands
		handStr := "[ ? ? ]" // Default hidden hand
		// Note: We can't type assert here anymore since we're using the interface
		// We'll need to add a method to the Player interface to check if it's human
		handStr = p.GetHand().String()

		fmt.Printf("- %s: Chips: %d | Bet: %d | Hand: %s%s\n",
			p.GetID(),
			p.GetChips(),
			p.GetCurrentBet(),
			handStr,
			status)
	}
	fmt.Println("==================================================")
}

// LogAction prints a message describing a player's action.
func (ui *ConsoleUI) LogAction(playerID string, action string, amount int) {
	if amount > 0 {
		fmt.Printf(">> %s %s (%d)\n", playerID, action, amount)
	} else {
		fmt.Printf(">> %s %s\n", playerID, action)
	}
}
