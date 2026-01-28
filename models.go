// Package main - In Go, every executable program must be in package "main"
// and have a func main() entry point. This file defines our data structures.

package main

import "time"

// Habit represents a single habit the user wants to track.
// In Go, we use structs to group related data together.
// The `json:"id"` tags tell the JSON encoder/decoder what field name to use.
type Habit struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Unit      string    `json:"unit"`
	CreatedAt time.Time `json:"created_at"`
}

// DayRecord stores what happened on a specific day.
type DayRecord struct {
	Date                   string `json:"date"`
	CompletedHabits        []int  `json:"completed_habits"`
	WeekReviewDone         bool   `json:"week_review_done"`
	PenaltyAppliedForHabits []int  `json:"penalty_applied_habits,omitempty"`
}

// AppData is the root structure we persist to JSON.
type AppData struct {
	Habits         []Habit              `json:"habits"`
	History        map[string]DayRecord  `json:"history"`
	LastWeekReview string               `json:"last_week_review"`
	CreatedAt      string               `json:"created_at"`
}
