package main

import (
	"bufio"
	"fmt"
	"os"
	"pokerclientv1/internal/game"
	"pokerclientv1/internal/player"
	"pokerclientv1/internal/types"
	"pokerclientv1/internal/ui"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("Welcome to Poker Client V1!")
	reader := bufio.NewReader(os.Stdin)

	// Get game settings from user
	numBots := promptForInt(reader, "Enter the number of bot opponents: ", 1, 5) // Limit bots for simplicity
	startingChips := promptForInt(reader, "Enter the starting chip amount for each player: ", 100, 10000)
	gameSpeedChoice := promptForGameSpeed(reader, "Select game speed (instant, fast, default, slow): ")
	gameSpeed := getSpeedDuration(gameSpeedChoice)

	// Initialize the UI
	consoleUI := ui.NewConsoleUI()

	// Create players
	players := []types.Player{}
	humanPlayer := player.NewHumanPlayer("Player 1", startingChips)
	players = append(players, humanPlayer)

	for i := 0; i < numBots; i++ {
		botID := fmt.Sprintf("Bot %d", i+1)
		difficulty := promptForDifficulty(reader, fmt.Sprintf("Enter difficulty for %s (easy, medium, hard): ", botID))
		// Use a default turn delay for now
		botPlayer := player.NewBotPlayer(botID, startingChips, difficulty, 500*time.Millisecond)
		players = append(players, botPlayer)
	}

	// Create and start the game
	pokerGame := game.NewGame(players, consoleUI, gameSpeed) // Pass game speed
	pokerGame.Start()

	fmt.Println("Thank you for playing!")
}

// Helper function to prompt for integer input
func promptForInt(reader *bufio.Reader, prompt string, min int, max int) int {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		val, err := strconv.Atoi(input)
		if err == nil && val >= min && val <= max {
			return val
		}
		fmt.Printf("Invalid input. Please enter a number between %d and %d.\n", min, max)
	}
}

// Helper function to prompt for difficulty
func promptForDifficulty(reader *bufio.Reader, prompt string) string {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "easy" || input == "medium" || input == "hard" {
			return input
		}
		fmt.Println("Invalid input. Please enter 'easy', 'medium', or 'hard'.")
	}
}

// Helper function to prompt for game speed
func promptForGameSpeed(reader *bufio.Reader, prompt string) string {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		speeds := map[string]bool{"instant": true, "fast": true, "default": true, "slow": true}
		if speeds[input] {
			return input
		}
		fmt.Println("Invalid input. Please enter 'instant', 'fast', 'default', or 'slow'.")
	}
}

// Helper function to get duration from speed choice
func getSpeedDuration(speed string) time.Duration {
	switch speed {
	case "instant":
		return 0 // No delay
	case "fast":
		return 500 * time.Millisecond
	case "default":
		return 1 * time.Second
	case "slow":
		return 2 * time.Second
	default:
		return 1 * time.Second // Default speed
	}
}
