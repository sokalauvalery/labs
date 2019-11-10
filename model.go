package main

import "github.com/gofrs/uuid"

var defaultUserUUID = uuid.Nil

// GameState for game state representation (win / lost)
type GameState string

const (
	// Win user win a game
	Win GameState = "win"
	// Lost user lost a game
	Lost GameState = "lost"
)

// User contains user related information
type User struct {
	ID      uuid.UUID
	Balance float64
}

// BalanceUpdate contains balance update data
type BalanceUpdate struct {
	ID          uuid.UUID
	Amount      float64
	GameState   GameState
	DeletedAt   *int64
	PrevBalance float64
}
