package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
)

// Manager main business logic object
type Manager interface {
	Update(context.Context, User, BalanceUpdate) error
	Cancel(ctx context.Context, count int) error
}

// NewStateManager update manager constructor
func NewStateManager(db *sql.DB) (Manager, error) {
	repo, err := NewRepo(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository %w", err)
	}
	return &stateManager{
		repo: repo,
	}, nil
}

type stateManager struct {
	repo Repo
	db   *sql.DB
}

// Update updates user balance and store operation history
func (m *stateManager) Update(ctx context.Context, user User, update BalanceUpdate) error {
	// m.db.BeginTx(ctx, nil)
	// txRepo := NewRepo()
	return m.repo.RegisterNewOperation(ctx, user.ID, update, calculateUserBalance)
}

func calculateUserBalance(currentBalance float64, update BalanceUpdate) float64 {
	if update.GameState == Win {
		return currentBalance + update.Amount
	}
	if currentBalance-update.Amount < 0 {
		return 0
	}
	return currentBalance - update.Amount
}

// Cancel lattest opperations
func (m *stateManager) Cancel(ctx context.Context, count int) error {
	return m.repo.CancelOperations(ctx, count, getUpdatesToCancel)
}

func getUpdatesToCancel(updates []BalanceUpdate) ([]string, float64) {
	if len(updates) == 0 {
		return nil, 0
	}

	operationsToDelete := []string{}
	operationsToDeleteMap := map[uuid.UUID]interface{}{}
	for i, update := range updates {
		if i%2 != 0 && update.DeletedAt == nil {
			operationsToDelete = append(operationsToDelete, update.ID.String())
			operationsToDeleteMap[update.ID] = nil
			continue
		}
	}
	if len(operationsToDelete) == 0 {
		return nil, 0
	}
	var balance *float64
	for _, update := range updates {
		if _, ok := operationsToDeleteMap[update.ID]; ok {
			if balance == nil {
				prevBalance := update.PrevBalance
				balance = &prevBalance
			}
			continue
		}
		if balance == nil {
			continue
		}

		newBalance := calculateUserBalance(*balance, update)
		balance = &newBalance

	}

	Log("cancel user updates [%v] set balance %v", operationsToDelete, *balance)
	return operationsToDelete, *balance
}
