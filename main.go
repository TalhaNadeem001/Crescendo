// main.go - Entry point of the application.
// In Go, the runtime looks for package "main" and the function "main()" to start the program.

package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
)

// loadEnv reads .env from the current directory and sets KEY=VALUE as environment variables.
func loadEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return // .env optional
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.Index(line, "="); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			if len(val) >= 2 && (val[0] == '"' && val[len(val)-1] == '"' || val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
			if key != "" {
				_ = os.Setenv(key, val)
			}
		}
	}
	if err := sc.Err(); err != nil {
		log.Println("reading .env:", err)
	}
}

func main() {
	loadEnv()
	// Register HTTP handlers: which function handles which URL path.
	// http.HandleFunc takes a pattern and a function. When a request matches the pattern,
	// Go calls your function with (http.ResponseWriter, *http.Request).
	// The leading slash is required; "/" matches the root path.
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/complete", HandleCompleteHabit)
	http.HandleFunc("/week-review", HandleWeekReview)
	http.HandleFunc("/add-habit", HandleAddHabit)
	http.HandleFunc("/edit-habit", HandleEditHabit)
	http.HandleFunc("/delete-habit", HandleDeleteHabit)
	http.HandleFunc("/add-todo", HandleAddTodo)
	http.HandleFunc("/complete-todo", HandleCompleteTodo)
	http.HandleFunc("/simplify-todo", HandleSimplifyTodo)

	// Start the HTTP server. ListenAndServe listens on port 8080 and blocks until the program exits.
	// The second argument is the handler for all requests; nil means use the default multiplexer
	// (which we configured with HandleFunc above).
	// To stop: press Ctrl+C in the terminal.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err) // panic stops the program and prints the error (ok for startup failures)
	}
}
