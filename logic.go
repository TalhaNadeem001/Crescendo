// This file contains the business logic: miss penalty, 7-day review, and date helpers.
// We keep this separate from HTTP handlers so the logic is easy to test and understand.

package main

import (
	"sort"
	"time"
)

// Go uses "reference time" for formatting: Mon Jan 2 15:04:05 MST 2006
// So "2006-01-02" means YYYY-MM-DD format.
const dateLayout = "2006-01-02"

// Today returns today's date as a string in YYYY-MM-DD format.
func Today() string {
	return time.Now().Format(dateLayout)
}

// Yesterday returns yesterday's date string.
func Yesterday() string {
	t := time.Now().AddDate(0, 0, -1)
	return t.Format(dateLayout)
}

// ParseDate converts a string like "2025-01-28" into a time.Time.
func ParseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}

// DaysBetween returns the number of days between start and end (end - start in days).
func DaysBetween(start, end string) (int, error) {
	s, err := time.Parse(dateLayout, start)
	if err != nil {
		return 0, err
	}
	e, err := time.Parse(dateLayout, end)
	if err != nil {
		return 0, err
	}
	days := int(e.Sub(s).Hours() / 24)
	if days < 0 {
		return 0, nil
	}
	return days, nil
}

// ApplyMissPenalty reduces a habit's quantity when the user missed a day.
// Rule: 5 -> 3, 3 -> 2, 2 -> 1. Minimum 1.
func ApplyMissPenalty(h *Habit) {
	if h.Quantity <= 1 {
		return
	}
	if h.Quantity >= 3 {
		h.Quantity -= 2
	} else {
		h.Quantity--
	}
}

// containsInt is a helper to check if a slice contains an integer (Go has no built-in for this).
func containsInt(slice []int, id int) bool {
	for _, v := range slice {
		if v == id {
			return true
		}
	}
	return false
}

// ProcessYesterdayMisses runs when you load the app: for yesterday only, for each habit
// that was NOT completed, we apply the miss penalty once and record it (so we don't apply again).
// So: one missed day = one reduction per habit. If you don't open the app for several days,
// we only penalize for "yesterday" (the day before today) each time you open the app.
func ProcessYesterdayMisses(data *AppData) {
	yesterday := Yesterday()
	rec, exists := data.History[yesterday]
	if !exists {
		rec = DayRecord{Date: yesterday}
	}
	// Ensure we have a slice to track penalty-applied (might be nil from old JSON).
	if rec.PenaltyAppliedForHabits == nil {
		rec.PenaltyAppliedForHabits = []int{}
	}

	changed := false
	for i := range data.Habits {
		h := &data.Habits[i]
		completed := containsInt(rec.CompletedHabits, h.ID)
		alreadyApplied := containsInt(rec.PenaltyAppliedForHabits, h.ID)
		if !completed && !alreadyApplied {
			ApplyMissPenalty(h)
			rec.PenaltyAppliedForHabits = append(rec.PenaltyAppliedForHabits, h.ID)
			changed = true
		}
	}
	if changed {
		data.History[yesterday] = rec
	}
}

// GetOrSetLastWeekReview returns the date we use for "last 7-day review".
func GetOrSetLastWeekReview(data *AppData) string {
	if data.LastWeekReview != "" {
		return data.LastWeekReview
	}
	if data.CreatedAt != "" {
		return data.CreatedAt
	}
	t := time.Now().AddDate(0, 0, -7)
	return t.Format(dateLayout)
}

// NeedsWeekReview returns true if 7 or more days have passed since the last week review.
func NeedsWeekReview(data *AppData) (bool, error) {
	last := GetOrSetLastWeekReview(data)
	days, err := DaysBetween(last, Today())
	if err != nil {
		return false, err
	}
	return days >= 7, nil
}

// CompleteWeekReview increments all habits by 1 and sets LastWeekReview to today.
func CompleteWeekReview(data *AppData) {
	for i := range data.Habits {
		data.Habits[i].Quantity++
	}
	data.LastWeekReview = Today()
}

// FindHabitByID returns a pointer to the habit with the given ID, or nil.
func FindHabitByID(data *AppData, id int) *Habit {
	for i := range data.Habits {
		if data.Habits[i].ID == id {
			return &data.Habits[i]
		}
	}
	return nil
}

// NextHabitID returns the next unused habit ID (max existing + 1).
func NextHabitID(data *AppData) int {
	max := 0
	for _, h := range data.Habits {
		if h.ID > max {
			max = h.ID
		}
	}
	return max + 1
}

// DatesInRange returns all date strings from start to end (inclusive), sorted.
func DatesInRange(start, end string) ([]string, error) {
	s, err := time.Parse(dateLayout, start)
	if err != nil {
		return nil, err
	}
	e, err := time.Parse(dateLayout, end)
	if err != nil {
		return nil, err
	}
	var out []string
	for d := s; !d.After(e); d = d.AddDate(0, 0, 1) {
		out = append(out, d.Format(dateLayout))
	}
	sort.Strings(out)
	return out, nil
}

// GetStreakForHabit returns the current streak (consecutive days completed) for a habit.
// We count backwards from yesterday (today doesn't count until the day is over).
func GetStreakForHabit(data *AppData, habitID int) int {
	streak := 0
	t := time.Now().AddDate(0, 0, -1) // yesterday
	for {
		key := t.Format(dateLayout)
		rec, exists := data.History[key]
		completed := false
		if exists {
			completed = containsInt(rec.CompletedHabits, habitID)
		}
		if !completed {
			break
		}
		streak++
		t = t.AddDate(0, 0, -1)
	}
	return streak
}
