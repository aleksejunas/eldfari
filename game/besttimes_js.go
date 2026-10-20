//go:build js

package game

import (
	"encoding/json"
	"fmt"
	"log"
	"syscall/js"
)

// storageKey is the localStorage key for the best times (as a JSON string)
const storageKey = "eldfari.best"

// loadBestTimes reads the best times from the browser's localStorage. If
// storage is blocked, the times are kept in memory only.
func loadBestTimes() *bestTimes {
	b := &bestTimes{times: map[string]float64{}, save: saveLocalStorage}
	var data string
	err := withStorage(func(ls js.Value) {
		if v := ls.Call("getItem", storageKey); !v.IsNull() {
			data = v.String()
		}
	})
	if err != nil {
		log.Printf("best times will not be saved: %v", err)
		b.save = nil
		return b
	}
	if data != "" {
		if err := json.Unmarshal([]byte(data), &b.times); err != nil {
			log.Printf("ignoring best times: %v", err)
			b.times = map[string]float64{}
		}
	}
	return b
}

func saveLocalStorage(times map[string]float64) error {
	data, err := json.Marshal(times)
	if err != nil {
		return err
	}
	return withStorage(func(ls js.Value) {
		ls.Call("setItem", storageKey, string(data))
	})
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
