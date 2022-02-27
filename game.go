package main

import (
	"fmt"
	"log"
	"time"
)

//
type GameState struct {
	number      int
	frameTicker *time.Ticker
	hub         *Hub
	reset       chan string
	lastResetBy string
}

func newGameState(hub *Hub) *GameState {
	frameTicker := time.NewTicker(1 * time.Second)
	return &GameState{
		number:      0,
		frameTicker: frameTicker,
		hub:         hub,
		reset:       make(chan string),
		lastResetBy: "",
	}
}

func (gs *GameState) run() {
	for {
		select {
		case <-gs.frameTicker.C:
			gs.number += 1
			log.Println(gs.number)
			gs.hub.broadcast <- []byte(fmt.Sprintf("Game State: %d", gs.number))
		case resetter := <-gs.reset:
			gs.number = 0
			gs.lastResetBy = resetter
			log.Println("Resetting")
			log.Println(gs.number)
		}

		gs.hub.broadcast <- []byte(fmt.Sprintf("Game State:\nnumber: %d\nLast reset by: %s", gs.number, gs.lastResetBy))
	}
}
