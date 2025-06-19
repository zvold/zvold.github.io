package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Abstracts actual database operations.
type Database struct {
	db      *sql.DB
	stopped chan struct{} // Receives an object when logger is stopped.
	stop    chan struct{}
}

// Opens an existing database or creates a new one at the specified path.
func InitDB(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return &Database{
		db:      createTables(db),
		stopped: make(chan struct{}),
		stop:    make(chan struct{}),
	}, nil
}

// Create in-memory database for testing.
func newInMemoryDB() (*Database, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	return &Database{
		db:      createTables(db),
		stopped: make(chan struct{}),
		stop:    make(chan struct{}),
	}, nil
}

func createTables(db *sql.DB) *sql.DB {
	var version string
	db.QueryRow(`SELECT sqlite_version()`).Scan(&version)
	slog.Info("Database version:", "version", version)

	// Initialize necessary tables.
	_, err := db.Exec(`create table if not exists days (
		date text primary key,
		work integer not null,
		rest integer not null
	);`)
	if err != nil {
		slog.Error("failed to create tables.", "err", err)
	}
	return db
}

func (db *Database) DaysCount() (count int) {
	db.db.QueryRow(`select count(*) from days;`).Scan(&count)
	return
}

// Starts a logger goroutine that stored the day's data into the db at midnight.
func (db *Database) StartLogger(state *State) {
	go func() {
		for {
			now := time.Now()
			dayEnd := time.Date(
				now.Year(), now.Month(), now.Day(),
				23, 59, 59, 1_000_000_000-1,
				now.Location(),
			)
			slog.Info("next daily logger tick at:", "date", dayEnd, "day", formatDate(now))
			var t *time.Ticker
			if now.After(dayEnd) {
				t = time.NewTicker(0 * time.Second) // tick immediately
			} else {
				t = time.NewTicker(dayEnd.Sub(now)) // tick at ~ 23:59:59
			}

			select {
			case <-t.C: // Tick: log the totals for the day.
				if err := db.StoreDailyTotals(state, dayEnd); err != nil {
					slog.Error("failed to update the daily total.", "err", err)
				}
				// Sleep for a while so the next day definitely starts.
				time.Sleep(2 * time.Second)
			case <-db.stop: // Request to stop the logger (server shutdown).
				t.Stop()
				defer func() {
					db.stopped <- struct{}{}
				}()
				return
			}
		}
	}()
}

// Updates the daily totals for the current day.
func (db *Database) StoreDailyTotals(state *State, now time.Time) error {
	work, rest := state.getTotalDurations(now)
	if work == 0 && rest == 0 {
		return nil
	}
	return db.StoreValue(now, work, rest)
}

func (db *Database) StoreValue(t time.Time, work, rest time.Duration) error {
	date := formatDate(t)
	slog.Info("updating the daily total.", "date", date, "work", work, "rest", rest)

	stmt, err := db.db.Prepare(`
		insert or replace into days(date, work, rest) select ?, ?, ?
		where not exists (
			select 1 from (select work, rest from days where date < ? order by date desc limit 1) as a
			where a.work = ? and a.rest = ?
		)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(
		date, work.Seconds(), rest.Seconds(),
		date, work.Seconds(), rest.Seconds())
	if err2 != nil {
		slog.Error("StoreValue() failed.", "err", err2)
	}
	return err2
}

func (db *Database) ReadTotals(t1, t2 string) []string {
	query := `select date, work, rest from days where date >= ? and date <= ? order by date desc`
	rows, err := db.db.Query(query, t1, t2)

	if err != nil {
		slog.Info("error reading data.", "err", err)
		return []string{}
	}

	result := make([]string, 0)
	for rows.Next() {
		var date string
		var work float64
		var rest float64
		if err := rows.Scan(&date, &work, &rest); err != nil {
			return nil
		}
		result = append(result, fmt.Sprintf("%s %.2f %.2f", date, work, rest))
	}
	return result
}

// Stops logger goroutine and blocks until it exits.
func (db *Database) StopLogger() {
	db.stop <- struct{}{}
	<-db.stopped
	slog.Info("daily logger has stopped.")
}
