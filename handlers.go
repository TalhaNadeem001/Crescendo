// Handlers define what happens when the user visits each URL (route).
// In Go, an HTTP handler is a function with signature: func(w http.ResponseWriter, r *http.Request)

package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// We parse templates once at startup and reuse them (more efficient than parsing on every request).
var tmpl *template.Template

func init() {
	// template.Must panics if there's an error - we want to fail fast at startup if templates are broken.
	// ParseFiles can take multiple files - we'll have one base and one page.
	tmpl = template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
}

// TemplateData holds everything we pass to the HTML template.
type TemplateData struct {
	Habits           []Habit
	History          map[string]DayRecord
	Today            string
	TodayRecord      DayRecord
	NeedsWeekReview  bool
	Streaks          map[int]int  // habit ID -> current streak
	CompletedToday   map[int]bool // habit ID -> completed today (for easy template checks)
	CalendarByHabit  map[int][]string // habit ID -> list of dates from creation to today
	CalendarHabit    map[string]bool // "habitID_date" -> completed (for heatmap)
	Message          string
}

// HandleIndex serves the main page: load data, process yesterday's misses, check week review, render HTML.
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	// Only allow GET for the index page (showing the form). POST goes to other handlers.
	if r.URL.Path != "/" && r.URL.Path != "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure CreatedAt is set on first run (so we have a start date for 7-day cycle).
	if data.CreatedAt == "" {
		data.CreatedAt = Today()
		_ = SaveData(data)
	}

	// Apply miss penalty for yesterday if any habit wasn't completed (only once per day).
	ProcessYesterdayMisses(data)
	if err := SaveData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	needsReview, _ := NeedsWeekReview(data)
	todayRec := data.History[Today()]

	streaks := make(map[int]int)
	completedToday := make(map[int]bool)
	for _, id := range todayRec.CompletedHabits {
		completedToday[id] = true
	}
	for _, h := range data.Habits {
		streaks[h.ID] = GetStreakForHabit(data, h.ID)
	}

	// Build per-habit date ranges for the calendar heatmap.
	// One box per day from habit creation through today; completed days get class "done" (green).
	calMap := make(map[string]bool)
	calendarByHabit := make(map[int][]string)
	now := time.Now()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	for _, h := range data.Habits {
		start := h.CreatedAt
		// If CreatedAt is zero (older data), fall back to app CreatedAt or today.
		if start.IsZero() {
			if data.CreatedAt != "" {
				if t, err := time.Parse("2006-01-02", data.CreatedAt); err == nil {
					start = t
				}
			}
			if start.IsZero() {
				start = now
			}
		}
		var dates []string
		for d := start; !d.After(todayEnd); d = d.AddDate(0, 0, 1) {
			ds := d.Format("2006-01-02")
			dates = append(dates, ds)
			rec := data.History[ds]
			for _, id := range rec.CompletedHabits {
				if id == h.ID {
					calMap[calendarKey(h.ID, ds)] = true
					break
				}
			}
		}
		calendarByHabit[h.ID] = dates
	}

	msg := ""
	switch {
	case r.URL.Query().Get("done") == "1":
		msg = "Habit marked complete for today!"
	case r.URL.Query().Get("review") == "1":
		msg = "Week review complete. All habits incremented!"
	case r.URL.Query().Get("added") == "1":
		msg = "Habit added!"
	case r.URL.Query().Get("edited") == "1":
		msg = "Habit name updated!"
	case r.URL.Query().Get("error") == "name":
		msg = "Please enter a habit name."
	}

	td := TemplateData{
		Habits:          data.Habits,
		History:         data.History,
		Today:           Today(),
		TodayRecord:     todayRec,
		NeedsWeekReview: needsReview,
		Streaks:         streaks,
		CompletedToday:  completedToday,
		CalendarByHabit: calendarByHabit,
		CalendarHabit:   calMap,
		Message:         msg,
	}
	// Execute the template named by the first file we parsed: "layout.html"
	if err := tmpl.ExecuteTemplate(w, "layout.html", td); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleCompleteHabit handles POST when user marks a habit as done for today.
// Form value: habit_id=1 (and optionally action=uncomplete to uncheck).
func HandleCompleteHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	habitIDStr := r.FormValue("habit_id")
	habitID, err := strconv.Atoi(habitIDStr)
	if err != nil {
		http.Redirect(w, r, "/?error=invalid", http.StatusFound)
		return
	}

	data, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if FindHabitByID(data, habitID) == nil {
		http.Redirect(w, r, "/?error=notfound", http.StatusFound)
		return
	}

	today := Today()
	rec := data.History[today]
	rec.Date = today
	if rec.CompletedHabits == nil {
		rec.CompletedHabits = []int{}
	}

	action := r.FormValue("action")
	if action == "uncomplete" {
		// Remove habit from completed list.
		var newList []int
		for _, id := range rec.CompletedHabits {
			if id != habitID {
				newList = append(newList, id)
			}
		}
		rec.CompletedHabits = newList
	} else {
		// Add to completed if not already there.
		found := false
		for _, id := range rec.CompletedHabits {
			if id == habitID {
				found = true
				break
			}
		}
		if !found {
			rec.CompletedHabits = append(rec.CompletedHabits, habitID)
		}
	}
	data.History[today] = rec
	if err := SaveData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/?done=1", http.StatusFound)
}

// HandleWeekReview handles POST when user completes the mandatory 7-day review (increment all).
func HandleWeekReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	CompleteWeekReview(data)
	if err := SaveData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/?review=1", http.StatusFound)
}

// HandleAddHabit handles POST to add a new habit. Form: name=Pushups&quantity=5&unit=pushups
func HandleAddHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/?error=name", http.StatusFound)
		return
	}
	qtyStr := r.FormValue("quantity")
	qty := 5
	if qtyStr != "" {
		if n, err := strconv.Atoi(qtyStr); err == nil && n > 0 {
			qty = n
		}
	}
	unit := strings.TrimSpace(r.FormValue("unit"))
	if unit == "" {
		unit = "units"
	}

	data, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h := Habit{
		ID:        NextHabitID(data),
		Name:      name,
		Quantity:  qty,
		Unit:      unit,
		CreatedAt: time.Now(),
	}
	data.Habits = append(data.Habits, h)
	if err := SaveData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/?added=1", http.StatusFound)
}

// HandleEditHabit handles POST to edit a habit's name (and optionally quantity/unit).
// Available anytime; especially useful during the 7-day review. Form: habit_id=1&name=New Name
func HandleEditHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	habitID, err := strconv.Atoi(r.FormValue("habit_id"))
	if err != nil {
		http.Redirect(w, r, "/?error=invalid", http.StatusFound)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/?error=name", http.StatusFound)
		return
	}

	data, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	habit := FindHabitByID(data, habitID)
	if habit == nil {
		http.Redirect(w, r, "/?error=notfound", http.StatusFound)
		return
	}
	habit.Name = name
	// Optional: allow editing quantity and unit at week review
	if qtyStr := r.FormValue("quantity"); qtyStr != "" {
		if qty, err := strconv.Atoi(qtyStr); err == nil && qty > 0 {
			habit.Quantity = qty
		}
	}
	if unit := strings.TrimSpace(r.FormValue("unit")); unit != "" {
		habit.Unit = unit
	}
	if err := SaveData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/?edited=1", http.StatusFound)
}

// HandleDeleteHabit handles POST to delete a habit (optional - for cleanup).
func HandleDeleteHabit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	habitID, _ := strconv.Atoi(r.FormValue("habit_id"))
	data, err := LoadData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var newHabits []Habit
	for _, h := range data.Habits {
		if h.ID != habitID {
			newHabits = append(newHabits, h)
		}
	}
	data.Habits = newHabits
	_ = SaveData(data)
	http.Redirect(w, r, "/", http.StatusFound)
}

// calendarKey builds a key for the calendar map: "habitID_date".
func calendarKey(habitID int, date string) string {
	return strconv.Itoa(habitID) + "_" + date
}
