// Raketti: vector remake of a gravity game in the style of Thrust/Turboraketti.
// Game logic lives in sim/, window, input and drawing in game/.
package main

import (
	"log"

	"raketti/game"
)

func main() {
	if err := game.Run(); err != nil {
		log.Fatal(err)
	}
}
