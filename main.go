// main.go - Entry point of the application.
// In Go, the runtime looks for package "main" and the function "main()" to start the program.

package main

import (
	"net/http"
)

func main() {
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

	// Start the HTTP server. ListenAndServe listens on port 8080 and blocks until the program exits.
	// The second argument is the handler for all requests; nil means use the default multiplexer
	// (which we configured with HandleFunc above).
	// To stop: press Ctrl+C in the terminal.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err) // panic stops the program and prints the error (ok for startup failures)
	}
}
