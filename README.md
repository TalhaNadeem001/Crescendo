# Habit Tracker (Go)

A simple habit tracker web app built in **Go** to help you stay consistent. It includes a **TODO list** for daily tasks and a **habit tracker** with streaks and a 7-day level-up cycle. The code is heavily commented so you can learn Go as you read.

## How it works

### TODO List

1. **Add tasks** – Type a task in the card and click Add. Tasks appear in the same card.
2. **Complete a task** – Click ✓ on a task; it is removed from the list (checklist style).
3. **Simplify a task** – If a task feels too big, click **Simplify**. The app calls the OpenAI API to break it into up to 3 simpler subtasks, which replace the original task. Requires `OPENAI_KEY` in a `.env` file (see below).

### Habit Tracker

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

### Optional: Simplify (OpenAI)

To use the **Simplify** button on todo tasks (break a task into 3 subtasks via OpenAI), create a `.env` file in the project root:

```
OPENAI_KEY=sk-your-openai-api-key
```

The app loads `.env` at startup. If `OPENAI_KEY` is missing, Simplify will show an error when used. The rest of the app works without it.

## Project layout (learning Go)

| File | Purpose |
|------|--------|
| `main.go` | Entry point; loads `.env`, registers routes, starts the HTTP server. |
| `models.go` | Data structs: `Habit`, `Todo`, `DayRecord`, `AppData` (with JSON tags). |
| `storage.go` | Load/save `data.json` with a mutex to avoid races. |
| `logic.go` | Business rules: miss penalty, 7-day review, streaks, date helpers, `NextTodoID`. |
| `handlers.go` | HTTP handlers: index, complete/simplify todo, complete habit, week review, add/edit/delete habit. |
| `openai.go` | OpenAI API: break a task into 3 subtasks (Chat Completions). |
| `templates/` | HTML templates: layout (todo card + habit section) + index (with `{{.}}` and `{{range}}`). |

Data is stored in `data.json` in the project directory (create it by running the app). It includes `habits` and `todos`.

## Concepts used (for learning)

- **Packages**: `package main` and `import`
- **Structs and JSON**: `struct`, `json:"field"` tags, `encoding/json`
- **Pointers**: `*AppData`, `&data`, modifying in place
- **Errors**: `if err != nil`, returning `(value, error)`
- **HTTP**: `http.HandleFunc`, `http.ResponseWriter`, `*http.Request`
- **Templates**: `html/template`, `{{.}}`, `{{range}}`, `{{if}}`
- **Concurrency**: `sync.Mutex` for safe file access
