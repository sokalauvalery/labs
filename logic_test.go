package main

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	updateToBalance map[uuid.UUID]float64
}

func (r *mockRepo) RegisterNewOperation(ctx context.Context, userID uuid.UUID, update BalanceUpdate, balanceCalc func(float64, BalanceUpdate) float64) error {
	r.updateToBalance[update.ID] = balanceCalc(0, update)
	return nil
}

func (r *mockRepo) GetUser(userID uuid.UUID) (*User, error) {
	return nil, nil
}

func (r *mockRepo) CancelOperations(ctx context.Context, limit int, cfn cancelFunc) error {
	return nil
}

// TODO: Add more tests using table driven
func TestLogicSmoke(t *testing.T) {
	repo := &mockRepo{
		updateToBalance: make(map[uuid.UUID]float64),
	}
	mgr := stateManager{repo: repo}

	ctx := context.Background()
	err := mgr.repo.RegisterNewOperation(ctx, uuid.Nil, BalanceUpdate{
		Amount:    10,
		GameState: Win,
	}, calculateUserBalance)

	require.Nil(t, err)
	require.EqualValues(t, 10, repo.updateToBalance[uuid.Nil])
}

func TestGetUpdatesToCancel(t *testing.T) {
	updates := []BalanceUpdate{
		{
			ID:          uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001"),
			Amount:      20.15,
			PrevBalance: 0.0,
			GameState:   "win",
		},
		{
			ID:          uuid.FromStringOrNil("00000000-0000-0000-0000-000000000002"),
			Amount:      30.15,
			PrevBalance: 20.15,
			GameState:   "win",
		},
		{
			ID:          uuid.FromStringOrNil("00000000-0000-0000-0000-000000000003"),
			Amount:      10.03,
			PrevBalance: 50.3,
			GameState:   "lost",
		},
		{
			ID:          uuid.FromStringOrNil("00000000-0000-0000-0000-000000000004"),
			Amount:      6.03,
			PrevBalance: 40.27,
			GameState:   "win",
		},
	}
	toDeleted, balance := getUpdatesToCancel(updates)
	require.EqualValues(t, []string{"00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000004"}, toDeleted)
	require.EqualValues(t, 10.12, balance)
}
