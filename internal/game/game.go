package game

import (
	"fmt"
	"pokerclientv1/internal/player"
	"pokerclientv1/internal/types"
	"strings"
	"time"
)

const (
	SmallBlind = 1
	BigBlind   = 2
	MinRaise   = BigBlind // Minimum raise amount must be at least the big blind
)

// Game manages the overall poker game state and flow.
type Game struct {
	Players       []types.Player
	Deck          *Deck
	Table         *types.Table
	Pot           int // Central pot
	DealerPos     int
	CurrentPlayer int
	SmallBlindPos int
	BigBlindPos   int
	UI            types.GameUI  // UI interface for display and logging
	GameSpeed     time.Duration // Delay between steps
	gameOver      bool          // Flag to signal game end
}

// NewGame initializes a new game with players.
func NewGame(players []types.Player, ui types.GameUI, gameSpeed time.Duration) *Game {
	return &Game{
		Players:       players,
		Deck:          NewDeck(),
		Table:         &types.Table{},
		Pot:           0,
		DealerPos:     0,
		CurrentPlayer: 0,
		SmallBlindPos: 0,
		BigBlindPos:   0,
		UI:            ui,
		GameSpeed:     gameSpeed, // Store game speed
		gameOver:      false,
	}
}

// Start begins the main game loop.
func (g *Game) Start() {
	fmt.Println("Starting Poker Game!")
	handNumber := 1
	for !g.gameOver {
		// Check for game end conditions before starting the hand
		if g.checkGameOver() {
			break
		}

		fmt.Printf("\n--- Starting Hand %d ---\n", handNumber)
		g.playHand()

		// Check for game end immediately after the hand (e.g., if human folded and lost)
		if g.gameOver {
			break
		}

		g.removeBrokePlayers() // Remove players with 0 chips

		// Rotate dealer position for the next hand, only among remaining players
		if len(g.Players) > 0 { // Avoid panic if all players are removed
			g.DealerPos = (g.DealerPos + 1) % len(g.Players)
		}

		g.waitWithLoader(g.GameSpeed * 2) // Pause between hands
		handNumber++
	}

	fmt.Println("\n--- Game Over --- ")
	// Display final chip counts if players remain
	if len(g.Players) > 0 {
		fmt.Println("Final Chip Counts:")
		for _, p := range g.Players {
			fmt.Printf("- %s: %d chips\n", p.GetID(), p.GetChips())
		}
	}
}

// getPlayersWithChips returns players who have chips > 0.
func (g *Game) getPlayersWithChips() []types.Player {
	active := []types.Player{}
	for _, p := range g.Players {
		if p.GetChips() > 0 {
			active = append(active, p)
		}
	}
	return active
}

// checkGameOver checks if the game should end.
func (g *Game) checkGameOver() bool {
	if g.gameOver { // Already flagged by player action
		return true
	}

	playersWithChips := g.getPlayersWithChips()
	if len(playersWithChips) <= 1 {
		fmt.Println("Only one player remains!")
		g.gameOver = true
		return true
	}

	// Check if human player exists and has chips
	humanAlive := false
	for _, p := range playersWithChips {
		if p.IsHuman() {
			humanAlive = true
			break
		}
	}
	if !humanAlive {
		// Find the human player in the original list to check if they were ever in the game
		humanWasPresent := false
		for _, p := range g.Players {
			if p.IsHuman() {
				humanWasPresent = true
				break
			}
		}
		if humanWasPresent {
			fmt.Println("You are out of chips!")
			g.gameOver = true
			return true
		}
	}

	return false
}

// removeBrokePlayers removes players with zero chips from the game.
func (g *Game) removeBrokePlayers() {
	remainingPlayers := []types.Player{}
	for _, p := range g.Players {
		if p.GetChips() > 0 {
			remainingPlayers = append(remainingPlayers, p)
		} else {
			if p.IsHuman() {
				// Human is broke, game over is handled in checkGameOver
				// Keep human in the list for final display, but checkGameOver will stop the loop
				remainingPlayers = append(remainingPlayers, p) // Keep human for final display
			} else {
				fmt.Printf("\n>> %s was kicked out due to being poor.\n", p.GetID())
				g.waitWithLoader(g.GameSpeed)
			}
		}
	}
	g.Players = remainingPlayers
	// Adjust dealer position if it's now out of bounds
	if len(g.Players) > 0 {
		g.DealerPos %= len(g.Players)
	} else {
		g.DealerPos = 0
	}
}

// getActivePlayers returns players who haven't folded and have chips.
func (g *Game) getActivePlayers() []types.Player {
	active := []types.Player{}
	for _, p := range g.Players {
		// Consider players active if they haven't folded in the *current* hand
		// and have chips. Folded status is reset each hand.
		if !p.IsFolded() && p.GetChips() > 0 {
			active = append(active, p)
		}
	}
	return active
}

// getPlayersInHand returns players who haven't folded in the current hand.
func (g *Game) getPlayersInHand() []types.Player {
	active := []types.Player{}
	for _, p := range g.Players {
		if !p.IsFolded() {
			active = append(active, p)
		}
	}
	return active
}

// playHand executes a single hand of poker.
func (g *Game) playHand() {
	// 0. Clear screen at the start of the hand
	g.UI.ClearScreen()

	// Reset gameOver flag potentially set by human player exiting previous hand's betting
	// g.gameOver = false // Let checkGameOver handle this

	// Check if enough players to play
	if len(g.getPlayersWithChips()) < 2 {
		g.gameOver = true
		return
	}

	// 1. Reset table and player states for the new hand
	g.resetForNewHand()

	// 2. Shuffle the deck
	g.Deck.Shuffle()

	// 3. Determine blind positions
	g.determineBlinds()

	// 4. Post blinds
	g.postBlinds()

	// 5. Deal initial hands (2 cards each for Texas Hold'em)
	g.dealHands(2)
	g.waitWithLoader(g.GameSpeed)

	// 6. Pre-flop betting round
	g.Table.Round = "Pre-flop"
	g.UI.DisplayGameState(g.Table, g.Players, g.Pot, "Pre-flop Betting")
	if !g.runBettingRound((g.BigBlindPos + 1) % len(g.Players)) {
		g.awardPotUncontested()
		return // Hand ends early
	}
	if g.gameOver {
		return
	} // Check if player exited during betting

	// 7. Flop
	g.dealCommunityCards("Flop", 3)
	g.waitWithLoader(g.GameSpeed)
	g.UI.DisplayGameState(g.Table, g.Players, g.Pot, "Flop Betting")
	if !g.runBettingRound(g.SmallBlindPos) {
		g.awardPotUncontested()
		return // Hand ends early
	}
	if g.gameOver {
		return
	} // Check if player exited during betting

	// 8. Turn
	g.dealCommunityCards("Turn", 1)
	g.waitWithLoader(g.GameSpeed)
	g.UI.DisplayGameState(g.Table, g.Players, g.Pot, "Turn Betting")
	if !g.runBettingRound(g.SmallBlindPos) {
		g.awardPotUncontested()
		return // Hand ends early
	}
	if g.gameOver {
		return
	} // Check if player exited during betting

	// 9. River
	g.dealCommunityCards("River", 1)
	g.waitWithLoader(g.GameSpeed)
	g.UI.DisplayGameState(g.Table, g.Players, g.Pot, "River Betting")
	if !g.runBettingRound(g.SmallBlindPos) {
		g.awardPotUncontested()
		return // Hand ends early
	}
	if g.gameOver {
		return
	} // Check if player exited during betting

	// 10. Showdown
	g.waitWithLoader(g.GameSpeed)
	g.showdown()
}

// resetForNewHand prepares the game state for a new hand.
func (g *Game) resetForNewHand() {
	g.Deck = NewDeck() // Get a fresh deck
	g.Table.ResetForNewHand()
	g.Pot = 0
	for _, p := range g.Players {
		p.ResetForNewHand()
	}
}

// determineBlinds sets the small and big blind positions based on the dealer.
func (g *Game) determineBlinds() {
	numPlayers := len(g.Players)
	g.SmallBlindPos = (g.DealerPos + 1) % numPlayers
	g.BigBlindPos = (g.DealerPos + 2) % numPlayers
	// Handle heads-up case (2 players)
	if numPlayers == 2 {
		g.SmallBlindPos = g.DealerPos
		g.BigBlindPos = (g.DealerPos + 1) % numPlayers
	}
	fmt.Printf("Dealer: %s | Small Blind: %s | Big Blind: %s\n",
		g.Players[g.DealerPos].GetID(),
		g.Players[g.SmallBlindPos].GetID(),
		g.Players[g.BigBlindPos].GetID())
}

// postBlinds forces the blind players to make their bets.
func (g *Game) postBlinds() {
	sbPlayer := g.Players[g.SmallBlindPos]
	bbPlayer := g.Players[g.BigBlindPos]

	sbAmount := g.forceBet(sbPlayer, SmallBlind)
	g.UI.LogAction(sbPlayer.GetID(), "posts small blind", sbAmount)

	bbAmount := g.forceBet(bbPlayer, BigBlind)
	g.UI.LogAction(bbPlayer.GetID(), "posts big blind", bbAmount)

	g.Table.CurrentBet = BigBlind // Initial bet to match is the Big Blind
}

// forceBet makes a player bet a specific amount, handling all-in cases.
func (g *Game) forceBet(p types.Player, amount int) int {
	betAmount := amount
	if p.GetChips() < amount {
		betAmount = p.GetChips() // All-in
		fmt.Printf("%s is all-in for the blind.\n", p.GetID())
	}
	p.RemoveChips(betAmount)
	p.SetCurrentBet(betAmount)
	g.Pot += betAmount
	return betAmount
}

// dealHands deals the initial private cards to each player.
func (g *Game) dealHands(numCards int) {
	fmt.Println("Dealing hands...")
	for i := 0; i < numCards; i++ {
		for _, p := range g.Players {
			if p.GetChips() > 0 { // Only deal to players with chips
				card, err := g.Deck.Deal()
				if err != nil {
					fmt.Printf("Error dealing card: %v\n", err)
					return // Or handle error more gracefully
				}
				p.GetHand().AddCard(card)
			}
		}
	}
	// Show human player their hand (if applicable)
	for _, p := range g.Players {
		if human, ok := p.(*player.HumanPlayer); ok {
			fmt.Printf("Your hand (%s): %s\n", human.GetID(), human.GetHand())
		}
	}
}

// dealCommunityCards deals cards to the table (Flop, Turn, River).
func (g *Game) dealCommunityCards(roundName string, numCards int) {
	fmt.Printf("--- Dealing %s ---\n", roundName)
	// Burn a card (optional, standard practice)
	_, err := g.Deck.Deal()
	if err != nil {
		fmt.Printf("Error burning card: %v\n", err)
		return
	}

	cards, err := g.Deck.DealMultiple(numCards)
	if err != nil {
		fmt.Printf("Error dealing %s cards: %v\n", roundName, err)
		return
	}
	for _, card := range cards {
		g.Table.AddCommunityCard(card)
	}
	g.Table.Round = roundName
	// Reset betting state for the new round
	g.Table.CurrentBet = 0
	for _, p := range g.Players {
		p.ResetBet()
	}
}

// runBettingRound manages the betting actions for a single round.
// Returns true if the hand should continue, false if only one player remains or player exits.
func (g *Game) runBettingRound(startPos int) bool {
	numPlayers := len(g.Players)
	lastRaiser := -1 // Index of the last player who raised
	playersActed := 0
	playersInRound := g.getPlayersInHand() // Players active at the start of this round
	numToAct := len(playersInRound)

	// Determine the initial player to act
	currentPlayerIndex := startPos
	for g.Players[currentPlayerIndex].IsFolded() || g.Players[currentPlayerIndex].GetChips() == 0 {
		currentPlayerIndex = (currentPlayerIndex + 1) % numPlayers
	}

	// The player who needs to act last is initially the one before the startPos
	// (usually the Big Blind in pre-flop, or player before dealer in post-flop)
	// This changes if someone raises.
	if g.Table.Round == "Pre-flop" {
		// actTarget = g.BigBlindPos // Big blind acts last pre-flop unless there's a raise
		// The logic now relies on checking if the action returns to the lastRaiser
	}

	for playersActed < numToAct {
		// Check if only one player is left in the hand (not just with chips)
		if len(g.getPlayersInHand()) <= 1 {
			return false // Hand ends, pot awarded uncontested later
		}

		currentPlayer := g.Players[currentPlayerIndex]

		// Skip folded players or players with no chips (already all-in)
		if currentPlayer.IsFolded() || (currentPlayer.GetChips() == 0 && currentPlayer.GetCurrentBet() > 0) {
			currentPlayerIndex = (currentPlayerIndex + 1) % numPlayers
			// Need to increment playersActed if skipping someone who already acted before raise
			// This logic gets complex with raises. Simpler to check the exit condition below.
			if lastRaiser == currentPlayerIndex {
				break // Action came back around
			}
			continue
		}

		// Check if the action has come back around to the last raiser
		if lastRaiser != -1 && lastRaiser == currentPlayerIndex {
			break // Betting round is over
		}

		// Get player action
		minRaiseAmount := MinRaise // Base minimum raise
		// TODO: Calculate min raise based on previous raises in the round if necessary
		action, amount := currentPlayer.TakeTurn(g.Table, g.Table.CurrentBet, minRaiseAmount)

		// Check for player exit
		if action == "exit" {
			fmt.Printf("\n%s has chosen to leave the table.\n", currentPlayer.GetID())
			g.gameOver = true
			return false // Signal game end
		}

		// Process action
		betAmount := 0
		switch action {
		case "fold":
			currentPlayer.SetFolded(true)
			g.UI.LogAction(currentPlayer.GetID(), "folds", 0)
		case "check":
			if g.Table.CurrentBet > currentPlayer.GetCurrentBet() {
				// This should be caught by TakeTurn, but double-check
				fmt.Printf("Error: %s cannot check, current bet is %d\n", currentPlayer.GetID(), g.Table.CurrentBet)
				// Force fold for now, or re-prompt human
				currentPlayer.SetFolded(true)
				g.UI.LogAction(currentPlayer.GetID(), "folds (error)", 0)
			} else {
				g.UI.LogAction(currentPlayer.GetID(), "checks", 0)
			}
		case "call":
			betAmount = amount
			if betAmount > currentPlayer.GetChips() {
				betAmount = currentPlayer.GetChips() // All-in call
			}
			callAmountNeeded := g.Table.CurrentBet - currentPlayer.GetCurrentBet()
			if betAmount != callAmountNeeded && currentPlayer.GetChips() >= callAmountNeeded {
				// Discrepancy, likely from TakeTurn logic vs game state
				fmt.Printf("Warning: Call amount mismatch for %s. Expected %d, got %d. Adjusting.\n", currentPlayer.GetID(), callAmountNeeded, betAmount)
				betAmount = callAmountNeeded
			}
			currentPlayer.RemoveChips(betAmount)
			currentPlayer.SetCurrentBet(currentPlayer.GetCurrentBet() + betAmount)
			g.Pot += betAmount
			g.UI.LogAction(currentPlayer.GetID(), "calls", betAmount)
		case "raise":
			betAmount = amount // Amount to ADD to the pot
			if betAmount > currentPlayer.GetChips() {
				betAmount = currentPlayer.GetChips() // All-in raise
			}

			totalPlayerBet := currentPlayer.GetCurrentBet() + betAmount
			actualRaiseAmount := totalPlayerBet - g.Table.CurrentBet

			// Validate raise amount (minimum raise, etc.) - Should be partially done in TakeTurn
			if totalPlayerBet <= g.Table.CurrentBet {
				fmt.Printf("Error: %s raise amount %d is not greater than current bet %d. Treating as call.\n", currentPlayer.GetID(), totalPlayerBet, g.Table.CurrentBet)
				// Treat as call
				callAmountNeeded := g.Table.CurrentBet - currentPlayer.GetCurrentBet()
				if callAmountNeeded < 0 {
					callAmountNeeded = 0
				}
				if callAmountNeeded > currentPlayer.GetChips() {
					callAmountNeeded = currentPlayer.GetChips()
				}
				betAmount = callAmountNeeded
				action = "call"
				currentPlayer.RemoveChips(betAmount)
				currentPlayer.SetCurrentBet(currentPlayer.GetCurrentBet() + betAmount)
				g.Pot += betAmount
				g.UI.LogAction(currentPlayer.GetID(), "calls (invalid raise)", betAmount)

			} else if actualRaiseAmount < MinRaise && currentPlayer.GetChips() > betAmount {
				// Invalid raise size (not all-in)
				fmt.Printf("Error: %s raise amount %d (total %d) is less than minimum raise %d. Forcing min raise or fold.\n", currentPlayer.GetID(), actualRaiseAmount, totalPlayerBet, MinRaise)
				// TODO: Handle this more gracefully - maybe force min raise if possible?
				// For now, treat as fold
				currentPlayer.SetFolded(true)
				g.UI.LogAction(currentPlayer.GetID(), "folds (invalid raise size)", 0)
				betAmount = 0
			} else {
				// Valid raise
				currentPlayer.RemoveChips(betAmount)
				currentPlayer.SetCurrentBet(totalPlayerBet)
				g.Pot += betAmount
				g.Table.CurrentBet = totalPlayerBet  // Update the high bet
				lastRaiser = currentPlayerIndex      // This player is the new last raiser
				playersActed = 0                     // Reset count since the bet changed
				numToAct = len(g.getPlayersInHand()) // Re-evaluate number of players to act
				g.UI.LogAction(currentPlayer.GetID(), fmt.Sprintf("raises to %d", totalPlayerBet), betAmount)
			}
		}

		// Check if player went all-in
		if currentPlayer.GetChips() == 0 && action != "fold" {
			fmt.Printf("%s is all-in!\n", currentPlayer.GetID())
		}

		// Only increment playersActed if the player wasn't skipped and didn't raise
		if action != "raise" {
			playersActed++
		}

		currentPlayerIndex = (currentPlayerIndex + 1) % numPlayers

		// Update UI after each action
		g.UI.DisplayGameState(g.Table, g.Players, g.Pot, g.Table.Round+" Betting")
		g.waitWithLoader(g.GameSpeed / 4) // Short pause after each action

	}

	// End of betting round cleanup
	g.waitWithLoader(g.GameSpeed / 2) // Short pause after betting round
	fmt.Println("Betting round finished.")
	fmt.Printf("Pot: %d\n", g.Pot)
	// Return true if more than one player is still in the hand
	return len(g.getPlayersInHand()) > 1
}

// showdown determines the winner(s) among the remaining players.
func (g *Game) showdown() {
	fmt.Println("--- Showdown ---")
	remainingPlayers := g.getPlayersInHand()

	if len(remainingPlayers) == 0 {
		fmt.Println("No players left for showdown?") // Should not happen
		return
	}

	if len(remainingPlayers) == 1 {
		g.awardPotUncontested()
		return
	}

	fmt.Println("Remaining players:")
	for _, p := range remainingPlayers {
		fmt.Printf("- %s: %s (Chips: %d)\n", p.GetID(), p.GetHand(), p.GetChips())
	}
	fmt.Printf("Community Cards: %v\n", g.Table.CommunityCards)

	// --- Hand Evaluation Logic ---
	// This is where the complex part of comparing poker hands goes.
	// For now, we'll just declare the first player as the winner.
	// TODO: Implement proper hand evaluation (Phase 1/5 refinement)
	winner := remainingPlayers[0]
	fmt.Printf("\n!!! Winner (Placeholder): %s !!!\n", winner.GetID())

	// Award pot
	g.awardPot(winner)
}

// awardPot gives the main pot to the winner.
// TODO: Handle side pots for all-in situations.
func (g *Game) awardPot(winner types.Player) {
	fmt.Printf("%s wins the pot of %d chips!\n", winner.GetID(), g.Pot)
	winner.AddChips(g.Pot)
	g.Pot = 0 // Reset pot
}

// awardPotUncontested gives the pot to the last remaining player.
func (g *Game) awardPotUncontested() {
	remaining := g.getPlayersInHand()
	if len(remaining) == 1 {
		winner := remaining[0]
		fmt.Printf("%s wins the pot of %d chips uncontested!\n", winner.GetID(), g.Pot)
		winner.AddChips(g.Pot)
		g.Pot = 0
	} else {
		fmt.Println("Error: Tried to award pot uncontested with multiple players remaining.")
	}
}

// waitWithLoader pauses execution for a duration and shows a simple loader.
func (g *Game) waitWithLoader(duration time.Duration) {
	if duration <= 0 {
		return // No delay for instant speed
	}
	loaderChars := []string{".   ", "..  ", "... ", "...."}
	startTime := time.Now()
	charIndex := 0
	for time.Since(startTime) < duration {
		// Print loader character and carriage return to overwrite
		fmt.Printf("\r%s", loaderChars[charIndex%len(loaderChars)])
		charIndex++
		time.Sleep(200 * time.Millisecond) // Update loader every 200ms
	}
	// Clear the loader line
	fmt.Printf("\r%s\r", strings.Repeat(" ", len(loaderChars[0])))
}
