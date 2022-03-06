package main

import (
	"encoding/json"
	// "fmt"
	"log"
	"math"
	"time"
)

//
type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

//
type Player struct {
	Name     string `json:"name"`
	Location Coord  `json:"location"`
	HexColor string `json:"hexColor"`
	Vector   Vector `json:"-"`
	Speed    int    `json:"-"`
}

type Vector struct {
	Dx int `json:"dx"`
	Dy int `json:"dy"`
}

type MoveCommand struct {
	playerId int
	vector   Vector
}
type PlayerControls struct {
	move chan MoveCommand
}

//
type GameState struct {
	player1     *Player
	player2     *Player
	number      int
	hub         *Hub
	lastResetBy string

	frameTicker *time.Ticker
	reset       chan string
	controls    *PlayerControls
}

type ClientGameState struct {
	Player1 *Player `json:"player1"`
	Player2 *Player `json:"player2"`
	Time    int     `json:"time"`
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func newGameState(hub *Hub) *GameState {
	frameTicker := time.NewTicker(33 * time.Millisecond)
	return &GameState{
		player1:     nil,
		player2:     nil,
		number:      0,
		frameTicker: frameTicker,
		hub:         hub,
		reset:       make(chan string),
		lastResetBy: "",
		controls: &PlayerControls{
			move: make(chan MoveCommand),
		},
	}
}

func resetGame(gs *GameState) {
	gs.player1 = &Player{
		Name:     "player 1",
		Location: Coord{X: 50, Y: 50},
		HexColor: "#88DD88",
		Vector: Vector{
			Dx: 0,
			Dy: 0,
		},
		Speed: 2,
	}
	gs.player2 = &Player{
		Name:     "player 2",
		Location: Coord{X: 300, Y: 300},
		HexColor: "#DD8888",
		Vector: Vector{
			Dx: 0,
			Dy: 0,
		},
		Speed: 2,
	}
	gs.number = 0
}

func unitVector(v Vector) Vector {
	if v.Dx == 0 && v.Dy == 0 {
		return Vector{Dx: 0, Dy: 0}
	}
	hypot := math.Sqrt(math.Pow(float64(v.Dx), 2) + math.Pow(float64(v.Dy), 2))
	return Vector{
		Dx: int(float64(v.Dx) / hypot),
		Dy: int(float64(v.Dy) / hypot),
	}
}

func (gs *GameState) run() {
	for {
		select {
		case <-gs.frameTicker.C:
			gs.number += 1
			// log.Println(
			// 	fmt.Sprintf("%d, %d (%d, %d)",
			// 		gs.player1.Location.X,
			// 		gs.player1.Location.Y,
			// 		gs.player1.Vector.Dx,
			// 		gs.player1.Vector.Dy,
			// 	))
			gs.player1.Location.X += gs.player1.Vector.Dx
			gs.player1.Location.Y += gs.player1.Vector.Dy
		case resetter := <-gs.reset:
			resetGame(gs)
			gs.lastResetBy = resetter
		case moveCmd := <-gs.controls.move:
			log.Println(moveCmd)
			var player *Player
			if moveCmd.playerId == 1 {
				player = gs.player1
			} else {
				player = gs.player2
			}
			unitVec := unitVector(moveCmd.vector)
			player.Vector = Vector{
				Dx: unitVec.Dx * player.Speed,
				Dy: unitVec.Dy * player.Speed,
			}
		}

		clientState := &ClientGameState{
			Player1: gs.player1,
			Player2: gs.player2,
			Time:    gs.number,
		}
		msg := &Message{
			Type: "GS",
			Data: clientState,
		}
		msgJson, err := json.Marshal(msg)
		if err != nil {
			// error
			return
		}
		gs.hub.broadcast <- msgJson
	}
}
