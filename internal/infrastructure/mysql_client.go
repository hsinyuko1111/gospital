package infrastructure

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/sony/gobreaker"

	"github.com/hsinyuko1111/gospital/internal/middleware"
)

// MySQLClient wraps sql.DB with circuit breaker
type MySQLClient struct {
	db *sql.DB
	cb *gobreaker.CircuitBreaker
}

// CBRow wraps sql.Row with circuit breaker protection
type CBRow struct {
	row *sql.Row
	cb  *gobreaker.CircuitBreaker
}

// Scan wraps sql.Row.Scan with circuit breaker protection
func (r *CBRow) Scan(dest ...interface{}) error {
	_, err := r.cb.Execute(func() (interface{}, error) {
		return nil, r.row.Scan(dest...)
	})
	return err
}

// Err returns any error encountered during the query
func (r *CBRow) Err() error {
	return r.row.Err()
}

// NewMySQLClient creates a new MySQL client with circuit breaker
func NewMySQLClient(db *sql.DB) *MySQLClient {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "MySQL",
		MaxRequests: 3,                // Allow 3 requests in half-open state
		Interval:    10 * time.Second, // Reset failure count every 10 seconds
		Timeout:     30 * time.Second, // Stay open for 30 seconds before half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Open circuit if failure ratio >= 60% with at least 3 requests
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit Breaker '%s' changed from '%s' to '%s'\n", name, from.String(), to.String())
			// Update Prometheus metric
			middleware.UpdateMySQLCircuitBreakerState(to.String())
		},
	})

	// Initialize metric to closed state
	middleware.UpdateMySQLCircuitBreakerState("closed")

	return &MySQLClient{
		db: db,
		cb: cb,
	}
}

// QueryContext executes a query with circuit breaker protection
func (c *MySQLClient) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.db.QueryContext(ctx, query, args...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*sql.Rows), nil
}

// QueryRowContext executes a single-row query with circuit breaker protection
func (c *MySQLClient) QueryRowContext(ctx context.Context, query string, args ...interface{}) *CBRow {
	row := c.db.QueryRowContext(ctx, query, args...)
	return &CBRow{
		row: row,
		cb:  c.cb,
	}
}

// ExecContext executes a command with circuit breaker protection
func (c *MySQLClient) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.db.ExecContext(ctx, query, args...)
	})
	if err != nil {
		return nil, err
	}
	return result.(sql.Result), nil
}

// GetDB returns the underlying sql.DB for advanced operations (transactions, etc.)
func (c *MySQLClient) GetDB() *sql.DB {
	return c.db
}

// PingContext checks the database connection
func (c *MySQLClient) PingContext(ctx context.Context) error {
	_, err := c.cb.Execute(func() (interface{}, error) {
		return nil, c.db.PingContext(ctx)
	})
	return err
}

// Close closes the database connection
func (c *MySQLClient) Close() error {
	return c.db.Close()
}
