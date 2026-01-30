// This file handles reading and writing our app data to a JSON file.
// In Go, we use the encoding/json package from the standard library.

package main

import (
	"encoding/json"
	"os"
	"sync"
)

// dataFile is the path to our JSON file. In Go, we can declare variables at package level.
const dataFile = "data.json"

// mu is a mutex (mutual exclusion lock). We use it so that when one HTTP request
// is reading/writing the file, another request doesn't do it at the same time (race condition).
// sync.Mutex has Lock() and Unlock() methods.
var mu sync.Mutex

// LoadData reads the JSON file from disk and decodes it into an AppData struct.
// It returns a pointer to AppData - in Go, we often use pointers (*AppData) to avoid
// copying large structs. The caller can modify the data and then call SaveData.
func LoadData() (*AppData, error) {
	mu.Lock()         // Acquire the lock - only one goroutine can hold it at a time
	defer mu.Unlock() // defer runs when the function returns - we always unlock, even on error

	// os.ReadFile reads the entire file into a byte slice ([]byte).
	// In Go, error is a built-in interface type - functions often return (value, error).
	bytes, err := os.ReadFile(dataFile)
	if err != nil {
		// os.IsNotExist checks if the error is "file not found" - first run
		if os.IsNotExist(err) {
			return &AppData{
				Habits:  []Habit{},
				Todos:   []Todo{},
				History: make(map[string]DayRecord), // maps must be initialized with make() before use
			}, nil
		}
		return nil, err // Pass through other errors (permission, etc.)
	}

	var data AppData
	// json.Unmarshal decodes JSON bytes into a struct. We pass a pointer &data so Unmarshal can fill it.
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	// If History was null in JSON, it decodes as nil. We need a non-nil map to add entries.
	if data.History == nil {
		data.History = make(map[string]DayRecord)
	}
	if data.Todos == nil {
		data.Todos = []Todo{}
	}
	return &data, nil
}

// SaveData encodes the AppData struct to JSON and writes it to the file.
// We use a pointer receiver (d *AppData) so we don't copy the whole struct.
func SaveData(d *AppData) error {
	mu.Lock()
	defer mu.Unlock()

	// json.MarshalIndent produces pretty-printed JSON (with indentation) - easier to read/debug.
	// The second argument is the prefix for each line (empty), third is indent string.
	bytes, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	// os.WriteFile writes bytes to a file. 0644 means: owner read+write, others read only (Unix permissions).
	return os.WriteFile(dataFile, bytes, 0644)
}
