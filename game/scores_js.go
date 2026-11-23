//go:build js

package game

import (
	"fmt"
	"log"
	"syscall/js"
)

// storageKey is the localStorage key for the high scores (a JSON string).
const storageKey = "eldfari.scores"

// loadScores reads the high scores from the browser's localStorage. If
// storage is blocked, the scores are kept in memory only.
func loadScores() *scores {
	var data string
	err := withStorage(func(ls js.Value) {
		if v := ls.Call("getItem", storageKey); !v.IsNull() {
			data = v.String()
		}
	})
	if err != nil {
		log.Printf("scores will not be saved: %v", err)
		return newScores()
	}
	s, err := decodeScores([]byte(data))
	if err != nil {
		log.Printf("ignoring scores: %v", err)
	}
	s.save = func(data []byte) error {
		return withStorage(func(ls js.Value) { ls.Call("setItem", storageKey, string(data)) })
	}
	return s
}

// withStorage calls f with window.localStorage. The browser can throw here
// (storage blocked, private mode, quota full), and syscall/js turns a thrown
// JavaScript exception into a Go panic, so recover turns it into an error.
func withStorage(f func(ls js.Value)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("localStorage: %v", r)
		}
	}()
	ls := js.Global().Get("localStorage")
	if !ls.Truthy() {
		return fmt.Errorf("localStorage is not available")
	}
	f(ls)
	return nil
}
