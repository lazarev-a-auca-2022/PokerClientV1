package game

import (
	"fmt"
	"pokerclientv1/internal/types"
	"testing"
	"time"
)

// --- Mock Implementations ---

// MockPlayer implements the types.Player interface for testing.
type MockPlayer struct {
	ID         string
	Chips      int
	Hand       *types.Hand
	Folded     bool
	CurrentBet int
	IsHumanVal bool
	// ActionQueue allows predefining actions for TakeTurn
	ActionQueue []struct {
		Action string
		Amount int
	}
	TurnCount int
}

func NewMockPlayer(id string, chips int, isHuman bool) *MockPlayer {
	return &MockPlayer{
		ID:         id,
		Chips:      chips,
		Hand:       &types.Hand{Cards: make([]types.Card, 0)},
		IsHumanVal: isHuman,
	}
}

func (mp *MockPlayer) GetID() string            { return mp.ID }
func (mp *MockPlayer) GetHand() *types.Hand     { return mp.Hand }
func (mp *MockPlayer) SetHand(hand *types.Hand) { mp.Hand = hand }
func (mp *MockPlayer) AddChips(amount int)      { mp.Chips += amount }
func (mp *MockPlayer) GetChips() int            { return mp.Chips }
func (mp *MockPlayer) IsFolded() bool           { return mp.Folded }
func (mp *MockPlayer) SetFolded(folded bool)    { mp.Folded = folded }
func (mp *MockPlayer) GetCurrentBet() int       { return mp.CurrentBet }
func (mp *MockPlayer) SetCurrentBet(amount int) { mp.CurrentBet = amount }
func (mp *MockPlayer) ResetBet()                { mp.CurrentBet = 0 }
func (mp *MockPlayer) IsHuman() bool            { return mp.IsHumanVal }
func (mp *MockPlayer) RemoveChips(amount int) error {
	if amount > mp.Chips {
		// Simulate all-in if trying to remove more than available
		// Return error only if amount is negative or zero?
		// For testing, let's assume game logic handles all-in correctly
		// and just remove what's possible.
		actualRemove := mp.Chips
		mp.Chips = 0
		return fmt.Errorf("mock player %s all-in, tried to remove %d, removed %d", mp.ID, amount, actualRemove)
	}
	mp.Chips -= amount
	return nil
}
func (mp *MockPlayer) ResetForNewHand() {
	mp.Hand = &types.Hand{Cards: make([]types.Card, 0)}
	mp.Folded = false
	mp.CurrentBet = 0
	mp.TurnCount = 0 // Reset turn count for action queue
}

// TakeTurn returns the next action from the queue.
func (mp *MockPlayer) TakeTurn(table *types.Table, currentBet int, minRaise int) (action string, amount int) {
	if mp.TurnCount >= len(mp.ActionQueue) {
		// Default action if queue is empty (e.g., fold)
		fmt.Printf("Warning: MockPlayer %s ran out of actions, defaulting to fold\n", mp.ID)
		return "fold", 0
	}
	a := mp.ActionQueue[mp.TurnCount]
	mp.TurnCount++
	return a.Action, a.Amount
}

// MockUI implements the types.GameUI interface for testing.
type MockUI struct {
	DisplayedStates []string // Store descriptions of displayed states
	LoggedActions   []string // Store logged actions
	Cleared         bool     // Track if ClearScreen was called
}

func (mu *MockUI) DisplayGameState(table *types.Table, players []types.Player, pot int, stage string) {
	mu.DisplayedStates = append(mu.DisplayedStates, fmt.Sprintf("Stage: %s, Pot: %d", stage, pot))
}
func (mu *MockUI) LogAction(playerID string, action string, amount int) {
	if amount > 0 {
		mu.LoggedActions = append(mu.LoggedActions, fmt.Sprintf("%s %s (%d)", playerID, action, amount))
	} else {
		mu.LoggedActions = append(mu.LoggedActions, fmt.Sprintf("%s %s", playerID, action))
	}
}
func (mu *MockUI) ClearScreen() {
	mu.Cleared = true
}

// --- Test Functions ---

// TestNewGame checks basic game initialization.
func TestNewGame(t *testing.T) {
	mockP1 := NewMockPlayer("P1", 100, true)
	mockP2 := NewMockPlayer("P2", 100, false)
	mockUI := &MockUI{}
	gameSpeed := 0 * time.Millisecond // Instant for tests

	game := NewGame([]types.Player{mockP1, mockP2}, mockUI, gameSpeed)

	if game == nil {
		t.Fatal("NewGame() returned nil")
	}
	if len(game.Players) != 2 {
		t.Errorf("NewGame() created game with %d players, want 2", len(game.Players))
	}
	if game.Pot != 0 {
		t.Errorf("NewGame() initial pot is %d, want 0", game.Pot)
	}
	if game.Table == nil {
		t.Errorf("NewGame() did not initialize Table")
	}
	if game.Deck == nil || len(game.Deck.cards) != 52 {
		t.Errorf("NewGame() did not initialize Deck correctly")
	}
	if game.UI != mockUI {
		t.Errorf("NewGame() did not set UI correctly")
	}
	if game.GameSpeed != gameSpeed {
		t.Errorf("NewGame() did not set GameSpeed correctly")
	}
}

// TestDetermineBlinds checks blind positions for 2 and 3 players.
func TestDetermineBlinds(t *testing.T) {
	mockP1 := NewMockPlayer("P1", 100, true)
	mockP2 := NewMockPlayer("P2", 100, false)
	mockP3 := NewMockPlayer("P3", 100, false)
	mockUI := &MockUI{}
	gameSpeed := 0 * time.Millisecond

	// Test 2 players (Heads-up)
	game2p := NewGame([]types.Player{mockP1, mockP2}, mockUI, gameSpeed)
	game2p.DealerPos = 0
	game2p.determineBlinds()
	if game2p.SmallBlindPos != 0 || game2p.BigBlindPos != 1 {
		t.Errorf("determineBlinds() 2p, Dealer 0: SB=%d BB=%d, want SB=0 BB=1", game2p.SmallBlindPos, game2p.BigBlindPos)
	}
	game2p.DealerPos = 1
	game2p.determineBlinds()
	if game2p.SmallBlindPos != 1 || game2p.BigBlindPos != 0 {
		t.Errorf("determineBlinds() 2p, Dealer 1: SB=%d BB=%d, want SB=1 BB=0", game2p.SmallBlindPos, game2p.BigBlindPos)
	}

	// Test 3 players
	game3p := NewGame([]types.Player{mockP1, mockP2, mockP3}, mockUI, gameSpeed)
	game3p.DealerPos = 0
	game3p.determineBlinds()
	if game3p.SmallBlindPos != 1 || game3p.BigBlindPos != 2 {
		t.Errorf("determineBlinds() 3p, Dealer 0: SB=%d BB=%d, want SB=1 BB=2", game3p.SmallBlindPos, game3p.BigBlindPos)
	}
	game3p.DealerPos = 1
	game3p.determineBlinds()
	if game3p.SmallBlindPos != 2 || game3p.BigBlindPos != 0 {
		t.Errorf("determineBlinds() 3p, Dealer 1: SB=%d BB=%d, want SB=2 BB=0", game3p.SmallBlindPos, game3p.BigBlindPos)
	}
	game3p.DealerPos = 2
	game3p.determineBlinds()
	if game3p.SmallBlindPos != 0 || game3p.BigBlindPos != 1 {
		t.Errorf("determineBlinds() 3p, Dealer 2: SB=%d BB=%d, want SB=0 BB=1", game3p.SmallBlindPos, game3p.BigBlindPos)
	}
}

// TestPostBlinds checks if blinds are posted correctly, including all-in.
func TestPostBlinds(t *testing.T) {
	mockUI := &MockUI{}
	gameSpeed := 0 * time.Millisecond

	// Scenario 1: Both players have enough chips
	mockP1 := NewMockPlayer("P1", 100, true)
	mockP2 := NewMockPlayer("P2", 100, false)
	game := NewGame([]types.Player{mockP1, mockP2}, mockUI, gameSpeed)
	game.DealerPos = 0
	game.determineBlinds() // SB=P1, BB=P2
	game.postBlinds()

	if mockP1.GetChips() != 100-SmallBlind || mockP1.GetCurrentBet() != SmallBlind {
		t.Errorf("PostBlinds() P1 chips/bet incorrect. Got %d/%d, want %d/%d", mockP1.GetChips(), mockP1.GetCurrentBet(), 100-SmallBlind, SmallBlind)
	}
	if mockP2.GetChips() != 100-BigBlind || mockP2.GetCurrentBet() != BigBlind {
		t.Errorf("PostBlinds() P2 chips/bet incorrect. Got %d/%d, want %d/%d", mockP2.GetChips(), mockP2.GetCurrentBet(), 100-BigBlind, BigBlind)
	}
	if game.Pot != SmallBlind+BigBlind {
		t.Errorf("PostBlinds() pot incorrect. Got %d, want %d", game.Pot, SmallBlind+BigBlind)
	}
	if game.Table.CurrentBet != BigBlind {
		t.Errorf("PostBlinds() table current bet incorrect. Got %d, want %d", game.Table.CurrentBet, BigBlind)
	}

	// Scenario 2: Small blind goes all-in
	mockP1 = NewMockPlayer("P1", SmallBlind-1, true)
	mockP2 = NewMockPlayer("P2", 100, false)
	game = NewGame([]types.Player{mockP1, mockP2}, mockUI, gameSpeed)
	game.DealerPos = 0
	game.determineBlinds() // SB=P1, BB=P2
	game.postBlinds()

	if mockP1.GetChips() != 0 || mockP1.GetCurrentBet() != SmallBlind-1 {
		t.Errorf("PostBlinds() All-in SB chips/bet incorrect. Got %d/%d, want %d/%d", mockP1.GetChips(), mockP1.GetCurrentBet(), 0, SmallBlind-1)
	}
	if mockP2.GetChips() != 100-BigBlind || mockP2.GetCurrentBet() != BigBlind {
		t.Errorf("PostBlinds() All-in SB, BB chips/bet incorrect. Got %d/%d, want %d/%d", mockP2.GetChips(), mockP2.GetCurrentBet(), 100-BigBlind, BigBlind)
	}
	if game.Pot != (SmallBlind-1)+BigBlind {
		t.Errorf("PostBlinds() All-in SB pot incorrect. Got %d, want %d", game.Pot, (SmallBlind-1)+BigBlind)
	}
	if game.Table.CurrentBet != BigBlind {
		t.Errorf("PostBlinds() All-in SB, table current bet incorrect. Got %d, want %d", game.Table.CurrentBet, BigBlind)
	}

	// Scenario 3: Big blind goes all-in
	mockP1 = NewMockPlayer("P1", 100, true)
	mockP2 = NewMockPlayer("P2", BigBlind-1, false)
	game = NewGame([]types.Player{mockP1, mockP2}, mockUI, gameSpeed)
	game.DealerPos = 0
	game.determineBlinds() // SB=P1, BB=P2
	game.postBlinds()

	if mockP1.GetChips() != 100-SmallBlind || mockP1.GetCurrentBet() != SmallBlind {
		t.Errorf("PostBlinds() All-in BB, SB chips/bet incorrect. Got %d/%d, want %d/%d", mockP1.GetChips(), mockP1.GetCurrentBet(), 100-SmallBlind, SmallBlind)
	}
	if mockP2.GetChips() != 0 || mockP2.GetCurrentBet() != BigBlind-1 {
		t.Errorf("PostBlinds() All-in BB chips/bet incorrect. Got %d/%d, want %d/%d", mockP2.GetChips(), mockP2.GetCurrentBet(), 0, BigBlind-1)
	}
	if game.Pot != SmallBlind+(BigBlind-1) {
		t.Errorf("PostBlinds() All-in BB pot incorrect. Got %d, want %d", game.Pot, SmallBlind+(BigBlind-1))
	}
	// Current bet should still be the attempted Big Blind value, even if player couldn't meet it
	if game.Table.CurrentBet != BigBlind {
		t.Errorf("PostBlinds() All-in BB, table current bet incorrect. Got %d, want %d", game.Table.CurrentBet, BigBlind)
	}
}

// TestDealHands checks if the correct number of cards are dealt.
func TestDealHands(t *testing.T) {
	mockP1 := NewMockPlayer("P1", 100, true)
	mockP2 := NewMockPlayer("P2", 100, false)
	mockP3 := NewMockPlayer("P3", 0, false) // Player with 0 chips
	mockUI := &MockUI{}
	gameSpeed := 0 * time.Millisecond
	game := NewGame([]types.Player{mockP1, mockP2, mockP3}, mockUI, gameSpeed)
	initialDeckSize := len(game.Deck.cards)
	numCardsToDeal := 2

	game.dealHands(numCardsToDeal)

	if len(mockP1.GetHand().Cards) != numCardsToDeal {
		t.Errorf("dealHands() P1 got %d cards, want %d", len(mockP1.GetHand().Cards), numCardsToDeal)
	}
	if len(mockP2.GetHand().Cards) != numCardsToDeal {
		t.Errorf("dealHands() P2 got %d cards, want %d", len(mockP2.GetHand().Cards), numCardsToDeal)
	}
	if len(mockP3.GetHand().Cards) != 0 {
		t.Errorf("dealHands() P3 (0 chips) got %d cards, want 0", len(mockP3.GetHand().Cards))
	}

	expectedDeckSize := initialDeckSize - (numCardsToDeal * 2) // Only P1 and P2 get cards
	if len(game.Deck.cards) != expectedDeckSize {
		t.Errorf("dealHands() deck size is %d, want %d", len(game.Deck.cards), expectedDeckSize)
	}
}

// TestDealCommunityCards checks dealing community cards (flop, turn, river).
func TestDealCommunityCards(t *testing.T) {
	mockP1 := NewMockPlayer("P1", 100, true)
	mockUI := &MockUI{}
	gameSpeed := 0 * time.Millisecond
	game := NewGame([]types.Player{mockP1}, mockUI, gameSpeed)
	initialDeckSize := len(game.Deck.cards)

	// Flop
	game.dealCommunityCards("Flop", 3)
	if len(game.Table.CommunityCards) != 3 {
		t.Errorf("dealCommunityCards() Flop dealt %d cards, want 3", len(game.Table.CommunityCards))
	}
	if len(game.Deck.cards) != initialDeckSize-(3+1) { // +1 for burn card
		t.Errorf("dealCommunityCards() Flop deck size is %d, want %d", len(game.Deck.cards), initialDeckSize-4)
	}
	if game.Table.Round != "Flop" {
		t.Errorf("dealCommunityCards() Flop did not set table round correctly")
	}

	// Turn
	initialDeckSize = len(game.Deck.cards)
	game.dealCommunityCards("Turn", 1)
	if len(game.Table.CommunityCards) != 3+1 {
		t.Errorf("dealCommunityCards() Turn dealt %d total cards, want 4", len(game.Table.CommunityCards))
	}
	if len(game.Deck.cards) != initialDeckSize-(1+1) { // +1 for burn card
		t.Errorf("dealCommunityCards() Turn deck size is %d, want %d", len(game.Deck.cards), initialDeckSize-2)
	}
	if game.Table.Round != "Turn" {
		t.Errorf("dealCommunityCards() Turn did not set table round correctly")
	}

	// River
	initialDeckSize = len(game.Deck.cards)
	game.dealCommunityCards("River", 1)
	if len(game.Table.CommunityCards) != 3+1+1 {
		t.Errorf("dealCommunityCards() River dealt %d total cards, want 5", len(game.Table.CommunityCards))
	}
	if len(game.Deck.cards) != initialDeckSize-(1+1) { // +1 for burn card
		t.Errorf("dealCommunityCards() River deck size is %d, want %d", len(game.Deck.cards), initialDeckSize-2)
	}
	if game.Table.Round != "River" {
		t.Errorf("dealCommunityCards() River did not set table round correctly")
	}
}

// TODO: Add tests for runBettingRound (complex scenarios)
// TODO: Add tests for showdown (requires hand evaluation or mocking)
// TODO: Add tests for awardPot, awardPotUncontested
// TODO: Add tests for removeBrokePlayers
// TODO: Add tests for checkGameOver
// TODO: Add tests for playHand (integration)
// TODO: Add tests for Start (integration)
