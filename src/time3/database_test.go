package main

import (
	"slices"
	"testing"
	"time"
)

func Test_StoreValue(t *testing.T) {
	db := createDB(t)

	now, _ := time.Parse(time.RFC3339, "2025-05-31T13:14:15Z")
	db.StoreValue(now, 12345*time.Millisecond, 67890*time.Millisecond)

	rows := db.ReadTotals("2025-05-28", "2025-05-31")
	want := []string{"2025-05-31 12.35 67.89"}
	if !slices.Equal(rows, want) {
		t.Errorf("db.ReadTotals(), want: %v, got: %v", want, rows)
	}
}

func Test_StoreValue_IfDifferent(t *testing.T) {
	db := createDB(t)

	// 2025-05-31 is added.
	now, _ := time.Parse(time.RFC3339, "2025-05-31T13:14:15Z")
	db.StoreValue(now, 12345*time.Millisecond, 67890*time.Millisecond)

	// 2025-06-01 is added (different values).
	now = now.Add(24 * time.Hour)
	db.StoreValue(now, 11111*time.Millisecond, 67890*time.Millisecond)

	rows := db.ReadTotals("2025-05-15", "2025-06-01")
	want := []string{"2025-06-01 11.11 67.89", "2025-05-31 12.35 67.89"}
	if !slices.Equal(rows, want) {
		t.Errorf("db.ReadTotals(), want: %v, got: %v", want, rows)
	}

	// 2025-06-02 is not added (exactly the same values).
	now = now.Add(24 * time.Hour)
	db.StoreValue(now, 11111*time.Millisecond, 67890*time.Millisecond)
	rows = db.ReadTotals("2025-05-15", "2025-06-02")
	if !slices.Equal(rows, want) {
		t.Errorf("db.ReadTotals(), want: %v, got: %v", want, rows)
	}
}

func createDB(t *testing.T) *Database {
	db, err := newInMemoryDB()
	if err != nil {
		t.Fatalf("Can't create database: %s", err)
	}
	return db
}
