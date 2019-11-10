package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type storage struct {
	db *sql.DB
}

type cancelFunc func([]BalanceUpdate) ([]string, float64)

// Repo contains methods to interact with database
// context is added to be able to cancel database query on rest request cancelation
type Repo interface {
	GetUser(userID uuid.UUID) (*User, error)
	RegisterNewOperation(ctx context.Context, userID uuid.UUID, update BalanceUpdate, balanceCalc func(float64, BalanceUpdate) float64) error
	CancelOperations(ctx context.Context, limit int, cfn cancelFunc) error
}

// NewRepo is constructor for db layer accessor
func NewRepo(db *sql.DB) (Repo, error) {
	storage := storage{db: db}
	if err := migrate(storage.db); err != nil {
		return nil, fmt.Errorf("Database migration failed: %w", err)
	}
	return &storage, nil
}

func (st *storage) CreateUser(ctx context.Context, user User) error {
	// TODO: consider what to do on same user registration request
	insertUserQuery := "INSERT INTO users (id) VALUES (?)"
	_, err := st.db.ExecContext(ctx, insertUserQuery, user.ID)
	return err
}

func (st *storage) GetUser(userID uuid.UUID) (*User, error) {
	var id uuid.UUID
	var balance float64
	row := st.db.QueryRow("SELECT id, balance FROM users WHERE id = ?", userID)
	err := row.Scan(&id, &balance)
	return &User{
		ID:      id,
		Balance: balance,
	}, err
}

// TODO: to many logic in database layer - move it to business layer
func (st *storage) RegisterNewOperation(ctx context.Context, userID uuid.UUID, update BalanceUpdate, balanceCalc func(float64, BalanceUpdate) float64) error {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var id uuid.UUID
	var balance float64
	Log("Update user %v balance %v", userID, update)
	row := tx.QueryRow("SELECT id, balance FROM users WHERE id = ?", userID)
	if err = row.Scan(&id, &balance); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to scan user data %w", err)
	}
	user := User{
		ID:      id,
		Balance: balance,
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO operations (id, amount, game_state, created_at, prev_balance) VALUES (?,?,?,?,?)", update.ID, update.Amount, update.GameState, time.Now().Unix(), balance); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to register operation %w", err)
	}

	newBalance := balanceCalc(user.Balance, update)
	Log("Set user balance %v", newBalance)

	if _, err := tx.ExecContext(ctx, "UPDATE users SET balance = ? WHERE id = ?", newBalance, user.ID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user balance %w", err)
	}
	return tx.Commit()
}

func (st *storage) CancelOperations(ctx context.Context, limit int, cfn cancelFunc) error {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, "SELECT id, amount, prev_balance, game_state, deleted_at FROM operations ORDER BY created_at LIMIT ?", limit)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to select operation to cancel %v", err)
	}
	var operations []BalanceUpdate
	for rows.Next() {
		update := BalanceUpdate{}
		if err = rows.Scan(&update.ID, &update.Amount, &update.PrevBalance, &update.GameState, &update.DeletedAt); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to scan operation rows %w", err)
		}
		operations = append(operations, update)
	}

	updatesToCancel, balance := cfn(operations)
	if len(updatesToCancel) == 0 {
		println("no cancel needed")
		tx.Rollback()
		return nil
	}

	// this is madness but i have no time (i had to use sqlx or gorm here)
	inQuery := []string{}
	inParameters := []interface{}{}
	for i := 0; i < len(updatesToCancel); i++ {
		inQuery = append(inQuery, "?")
		inParameters = append(inParameters, updatesToCancel[i])
	}

	updateQuery := fmt.Sprintf("UPDATE operations SET deleted_at = %v where id in (%v)", time.Now().Unix(), strings.Join(inQuery, ","))

	_, err = tx.ExecContext(ctx, updateQuery, inParameters...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to cancel operations %v", err)
	}

	if _, err := tx.ExecContext(ctx, "UPDATE users SET balance = ? WHERE id = ?", balance, defaultUserUUID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user balance %w", err)
	}

	return tx.Commit()
}
