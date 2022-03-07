package main

import (
	"encoding/json"
	"fmt"
	"log"

	// "fmt"
	// "log"
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
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location Coord  `json:"location"`
	HexColor string `json:"hexColor"`
	Vector   Vector `json:"-"`
	Speed    int    `json:"-"`
}

type Vector struct {
	Dx float64 `json:"dx"`
	Dy float64 `json:"dy"`
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
	players     map[int]*Player
	number      int
	hub         *Hub
	lastResetBy string

	frameTicker   *time.Ticker
	reset         chan string
	controls      *PlayerControls
	joinGame      chan *Player
	claimPlayerID chan int
}

type ClientGameState struct {
	Players map[int]*Player `json:"players"`
	Time    int             `json:"time"`
}

func newGameState(hub *Hub) *GameState {
	frameTicker := time.NewTicker(33 * time.Millisecond)
	return &GameState{
		players:     make(map[int]*Player),
		number:      0,
		frameTicker: frameTicker,
		hub:         hub,
		reset:       make(chan string),
		lastResetBy: "",
		controls: &PlayerControls{
			move: make(chan MoveCommand),
		},
		joinGame:      make(chan *Player),
		claimPlayerID: make(chan int),
	}
}

var HexColors = []string{
	"#ff8888",
	"#88ff88",
	"#8888ff",
	"#ffff66",
	"#66ffff",
	"#ff66ff",
}

func resetGame(gs *GameState) {
	gs.players = make(map[int]*Player)
	gs.number = 0
	// gs.player1 = &Player{
	// 	Name:     "player 1",
	// 	Location: Coord{X: 50, Y: 50},
	// 	HexColor: "#88DD88",
	// 	Vector: Vector{
	// 		Dx: 0,
	// 		Dy: 0,
	// 	},
	// 	Speed: 3,
	// }
	// gs.player2 = &Player{
	// 	Name:     "player 2",
	// 	Location: Coord{X: 300, Y: 300},
	// 	HexColor: "#DD8888",
	// 	Vector: Vector{
	// 		Dx: 0,
	// 		Dy: 0,
	// 	},
	// 	Speed: 3,
	// }
	// gs.number = 0
}

func unitVector(v Vector) Vector {
	if v.Dx == 0 && v.Dy == 0 {
		return Vector{Dx: 0, Dy: 0}
	}
	hypot := math.Sqrt(math.Pow(float64(v.Dx), 2) + math.Pow(float64(v.Dy), 2))
	return Vector{
		Dx: float64(v.Dx) / hypot,
		Dy: float64(v.Dy) / hypot,
	}
}

func (gs *GameState) generatePlayerIDs() {
	id := 0
	for {
		id += 1
		gs.claimPlayerID <- id
	}
}

func (gs *GameState) NewPlayer(name string) *Player {
	id := <-gs.claimPlayerID
	log.Println(fmt.Sprintf("New player: %d", id))
	player := &Player{
		ID:       id,
		Name:     name,
		Location: Coord{X: 100, Y: 100},
		HexColor: HexColors[id%len(HexColors)],
		Vector:   Vector{Dx: 0, Dy: 0},
		Speed:    3,
	}
	gs.joinGame <- player
	log.Println(fmt.Sprintf("Player %d joined!", id))
	return player
}

func (gs *GameState) run() {
	go gs.generatePlayerIDs()
	for {
		select {
		case <-gs.frameTicker.C:
			gs.number += 1
			for _, p := range gs.players {
				p.Location.X += int(p.Vector.Dx)
				p.Location.Y += int(p.Vector.Dy)
			}
		case resetter := <-gs.reset:
			resetGame(gs)
			gs.lastResetBy = resetter
		case moveCmd := <-gs.controls.move:
			player := gs.players[moveCmd.playerId]
			if player == nil {
				continue
			}
			unitVec := unitVector(moveCmd.vector)
			player.Vector = Vector{
				Dx: unitVec.Dx * float64(player.Speed),
				Dy: unitVec.Dy * float64(player.Speed),
			}
		case player := <-gs.joinGame:
			gs.players[player.ID] = player
		}

		clientState := &ClientGameState{
			Players: gs.players,
			Time:    gs.number,
		}
		msg := &GameMessage{
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
