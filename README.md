# Habit Tracker (Go)

A simple habit tracker web app built in **Go** to help you stay consistent. The code is heavily commented so you can learn Go as you read.

## How it works

1. **Add habits** – e.g. "5 pushups", "Read 30 min". Each habit has a name, quantity, and unit.
2. **Track daily** – Mark habits as done each day. You see a 30-day calendar (green = done) and current streak.
3. **Miss a day** – If you don’t complete a habit on a day, the target is reduced when you next open the app:
   - 5 → 3, 3 → 2, 2 → 1 (minimum 1).
4. **Every 7 days** – You’re prompted to complete a “week review”: all habit targets are incremented by 1. Adding a new habit at that time is optional; you can add habits anytime.

## Run the app

```bash
# From the project root
go run .

# Or build then run
go build -o habit-tracker .
./habit-tracker
```

Open **http://localhost:8080** in your browser.

## Project layout (learning Go)

| File | Purpose |
|------|--------|
| `main.go` | Entry point; registers routes and starts the HTTP server. |
| `models.go` | Data structs: `Habit`, `DayRecord`, `AppData` (with JSON tags). |
| `storage.go` | Load/save `data.json` with a mutex to avoid races. |
| `logic.go` | Business rules: miss penalty, 7-day review, streaks, date helpers. |
| `handlers.go` | HTTP handlers: index, complete habit, week review, add/delete habit. |
| `templates/` | HTML templates: layout + index (with `{{.}}` and `{{range}}`). |

Data is stored in `data.json` in the project directory (create it by running the app and adding a habit).

## Concepts used (for learning)

- **Packages**: `package main` and `import`
- **Structs and JSON**: `struct`, `json:"field"` tags, `encoding/json`
- **Pointers**: `*AppData`, `&data`, modifying in place
- **Errors**: `if err != nil`, returning `(value, error)`
- **HTTP**: `http.HandleFunc`, `http.ResponseWriter`, `*http.Request`
- **Templates**: `html/template`, `{{.}}`, `{{range}}`, `{{if}}`
- **Concurrency**: `sync.Mutex` for safe file access
