package main

import (
	"fmt"
	"pokerclientv1/internal/game"
	"pokerclientv1/internal/player"
	"pokerclientv1/internal/types"
	"pokerclientv1/internal/ui"
	"time"
)

const (
	startingChips = 1000 // Starting chips for each player
)

func main() {
	fmt.Println("Welcome to Poker Client V1!")

	// Initialize the UI
	consoleUI := ui.NewConsoleUI()

	// Create players
	humanPlayer := player.NewHumanPlayer("Player 1", startingChips)
	botPlayer := player.NewBotPlayer("Bot 1", startingChips, "easy", 1*time.Second) // Simple bot for now

	players := []types.Player{
		humanPlayer,
		botPlayer,
	}

	// Create and start the game
	pokerGame := game.NewGame(players, consoleUI)
	pokerGame.Start()

	fmt.Println("Thank you for playing!")
}
