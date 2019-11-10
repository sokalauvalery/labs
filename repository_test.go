package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

// TODO: Split on multiple small tests, consider using table driven tests here
func TestDBSmoke(t *testing.T) {
	db, err := newTestDB()
	require.Nil(t, err)
	require.NotNil(t, db)

	repo, err := NewRepo(db)
	require.Nil(t, err)

	testUserID := uuid.Nil

	ctx := context.Background()
	id, _ := uuid.NewV4()
	newBalance := 11.66
	update := BalanceUpdate{
		ID:        id,
		GameState: Win,
		Amount:    newBalance,
	}

	err = repo.RegisterNewOperation(ctx, testUserID, update, func(a float64, u BalanceUpdate) float64 {
		return a + u.Amount
	})
	require.Nil(t, err)

	user, err := repo.GetUser(uuid.Nil)
	require.Nil(t, err)
	require.NotNil(t, user)
	require.EqualValues(t, newBalance, user.Balance)

	err = repo.CancelOperations(ctx, 1, func([]BalanceUpdate) ([]string, float64) {
		return []string{id.String()}, 0.1
	})
	require.Nil(t, err)

	user, err = repo.GetUser(uuid.Nil)
	require.Nil(t, err)
	require.NotNil(t, user)
	require.EqualValues(t, 0.1, user.Balance)

}

func newTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("Failed to open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}
