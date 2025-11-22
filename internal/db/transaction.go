package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TransactionManager manages database transactions
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

type transactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store transaction in context so repositories can use it
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

// GetDBFromContext retrieves the database connection from context
// If a transaction is active, returns the transaction, otherwise returns the main DB
func GetDBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return defaultDB
}

type contextKey string

const txKey contextKey = "tx"

// TransactionError wraps transaction-related errors
type TransactionError struct {
	Err error
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("transaction error: %v", e.Err)
}

func (e *TransactionError) Unwrap() error {
	return e.Err
}

