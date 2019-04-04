/*
 * Copyright (c) 2018, Ryan Westlund.
 * This code is under the BSD 3-Clause license.
 */

package main

import (
	"math/rand"
	"strings"
	"time"
)

type Player struct {
	// These two channels are the same as the two from the corresponding User struct in server.go.
	InputChan  chan Message
	UpdateChan chan Update
	// The Command field is the current input from the player (kept up to date by the concurrently running input func for the player).
	Command string
	Life    int
	Stamina float32
	// The State field keeps track of what the player is doing. It has values like "standing", "blocking", "light attack", etc.
	State string
	// The StateDuration field shows how much longer the player will remain in their current state.
	StateDuration int
	// The Finished field shows what state the player just exited. It's used to know when an attack is supposed to land.
	Finished string
}

// NewPlayer returns a Player with all the starting values.
func NewPlayer(inputChan chan Message, updateChan chan Update) Player {
	return Player{InputChan: inputChan, UpdateChan: updateChan, Command: "NONE", Life: 100, Stamina: 100, State: "standing", StateDuration: 0, Finished: ""}
}

// Status returns a PlayerStatus from the Player, to be sent in an Update over the network.
func (p *Player) Status() PlayerStatus {
	return PlayerStatus{Life: p.Life, Stamina: p.Stamina, State: p.State, StateDuration: p.StateDuration}

}

// This is called every mainloop cycle, and does two things: regenerate stamina, and make progress toward exiting the current state.
func (p *Player) PassTime(amount int) {
	p.Stamina += 0.1
	if p.Stamina > 100 {
		p.Stamina = 100
	}
	p.StateDuration -= amount
	// If it starts with "interrupt", it's one of the heavy attack interrupt states. There are eight of them, so I didn't think it was practical to just list them all.
	if p.StateDuration <= 0 && !TERMINAL_STATES[p.State] && !strings.HasPrefix(p.State, "interrupt") {
		p.Finished = p.State
		p.State = "standing"
	}
}

func (p *Player) SetState(state string, duration int) {
	p.State = state
	p.StateDuration = duration
}

// This struct is passed instead of Player to the client in Updates so that unneeded fields like the channels aren't passed.
type PlayerStatus struct {
	Life          int     `json:"life"`
	Stamina       float32 `json:"stamina"`
	State         string  `json:"state"`
	StateDuration int     `json:"stateDur"`
}

// One of these is sent back to each player every mainloop cycle. Note that the players don't know which player they are internally - it doesn't matter.
type Update struct {
	Self  PlayerStatus `json:"self"`
	Enemy PlayerStatus `json:"enemy"`
}

// Balance parameters.
const (
	LIGHT_ATK_DMG int = 3
	// 'Speed' here actually means how long it takes, so it's misleading.
	LIGHT_ATK_SPD      int     = 50
	LIGHT_ATK_COST     float32 = 10.0
	LIGHT_ATK_BLK_COST float32 = 12.0
	// The counter window is how long you can counter for after the attacks starts - so a bigger value here means it's easier to counter.
	LIGHT_ATK_CNTR_WINDOW int     = 25
	LIGHT_ATK_CNTR_SPD    int     = 30
	LIGHT_ATK_CNTR_DMG    int     = 3
	HEAVY_ATK_DMG         int     = 6
	HEAVY_ATK_SPD         int     = 100
	HEAVY_ATK_COST        float32 = 15.0
	HEAVY_ATK_BLK_COST    float32 = 20.0
	HEAVY_ATK_BLKED_DMG   int     = 2
	DODGE_COST            float32 = 20.0
	DODGE_SPD             int     = 30
)

var INTERRUPTABLE_STATES = map[string]bool{"standing": true, "blocking": true}
var TERMINAL_STATES = map[string]bool{"standing": true, "blocking": true, "countered": true}
var ATTACK_STATES = map[string]bool{"light attack": true, "heavy attack": true}

// These are suffixes that can be attached to 'interrupted heavy' or 'interrupting heavy' to form the a state value that includes which arrow needs to be pressed.
var INTERRUPT_RESOLVE_KEYS = []string{"_up", "_down", "_left", "_right"}

func battle(player1inputChan, player2inputChan chan Message, player1updateChan, player2updateChan chan Update) {
	// Seed the random number generator and initialize the clock and players.
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	players := []Player{NewPlayer(player1inputChan, player1updateChan), NewPlayer(player2inputChan, player2updateChan)}
	for players[0].Life > 0 && players[1].Life > 0 {
		select {
		// Each mainloop cycle:
		case <-ticker.C:
			// Send updates to the clients.
			select {
			case players[0].UpdateChan <- Update{Self: players[0].Status(), Enemy: players[1].Status()}:
			default:
			}
			select {
			case players[1].UpdateChan <- Update{Self: players[1].Status(), Enemy: players[0].Status()}:
			default:
			}
			players[0].PassTime(1)
			players[1].PassTime(1)
			if players[0].Finished != "" {
				players[0], players[1] = resolveState(players[0], players[1])
			}
			if players[1].Finished != "" {
				players[1], players[0] = resolveState(players[1], players[0])
			}
			players[0], players[1] = resolveCommand(players[0], players[1], random)
			players[1], players[0] = resolveCommand(players[1], players[0], random)
		case input := <-players[0].InputChan:
			players[0].Command = input.Content
		case input := <-players[1].InputChan:
			players[1].Command = input.Content
		}
	}
	// Send one last update to the players so they know how the battle ended.
	players[0].UpdateChan <- Update{Self: players[0].Status(), Enemy: players[1].Status()}
	players[1].UpdateChan <- Update{Self: players[1].Status(), Enemy: players[0].Status()}

	// Make some goroutines to catch the last couple inputs from the players. This is necessary to stop
	// server.go from getting stuck trying to send their input through after the battle is over.
	stop1 := make(chan bool)
	stop2 := make(chan bool)
	go catchInput(players[0].InputChan, stop1)
	go catchInput(players[1].InputChan, stop2)
	time.Sleep(5 * time.Second)
	stop1 <- true
	stop2 <- true
}

func resolveState(player, enemy Player) (Player, Player) {
	switch player.Finished {
	case "light attack":
		if enemy.State == "blocking" {
			if enemy.Stamina >= LIGHT_ATK_BLK_COST {
				enemy.Stamina -= LIGHT_ATK_BLK_COST
				// If the enemy blocked inside the counterattack window...
				if -enemy.StateDuration >= LIGHT_ATK_SPD-LIGHT_ATK_CNTR_WINDOW {
					// The player is counterattacked. They are placed in a stunned state that they
					// must press a button to escape before the counterattack lands.
					player.SetState("countered", 0)
					enemy.SetState("counterattack", LIGHT_ATK_CNTR_SPD)
				}
			} else {
				// If you try to block an attack but you don't have enough stamina,
				// you still lose your stamina and you also take damage.
				enemy.Stamina = 0.0
				enemy.Life -= LIGHT_ATK_DMG
			}
		} else {
			// If the enemy wasn't blocking, they just take damage and have their light attack canceled.
			enemy.Life -= LIGHT_ATK_DMG
			if enemy.State == "light attack" {
				enemy.SetState("standing", 0)
			}
		}
	case "counterattack":
		// No conditions here because if you save against the counter attack it puts the enemy
		// out of the counterattacking state (so if you're here then they must not have saved).
		enemy.Life -= LIGHT_ATK_CNTR_DMG
		enemy.SetState("standing", 0)
	case "heavy attack":
		if enemy.State == "blocking" {
			if enemy.Stamina >= HEAVY_ATK_BLK_COST {
				enemy.Stamina -= HEAVY_ATK_BLK_COST
				enemy.Life -= HEAVY_ATK_BLKED_DMG
			} else {
				enemy.Stamina = 0.0
				enemy.Life -= HEAVY_ATK_DMG
			}
		} else {
			enemy.Life -= HEAVY_ATK_DMG
			enemy.SetState("standing", 0)
		}
	}
	player.Finished = ""
	return player, enemy
}

func resolveCommand(player, enemy Player, random *rand.Rand) (Player, Player) {
	// Interrupt resolution has to be handled first, otherwise non-arrow keys can't be punished.
	if strings.HasPrefix(player.State, "interrupt") && player.Command != "NONE" {
		// If we hit the right button (position 10 is just after the '_'):
		if strings.HasPrefix(player.Command, "INTERRUPT_") && strings.ToLower(player.Command[10:]) == player.State[strings.Index(player.State, "_")+1:] {
			// If we're not the interrupting player, we're the heavy
			// attack player, so the heavy attack hits.
			if !strings.HasPrefix(player.State, "interrupting") {
				enemy.Life -= HEAVY_ATK_DMG
			}
		} else {
			// Same as above only this time we hit the wrong button, so the condition
			// is reversed - we take damage if we're the interrupting player.
			if strings.HasPrefix(player.State, "interrupting") {
				player.Life -= HEAVY_ATK_DMG
			}
		}
		player.SetState("standing", 0)
		enemy.SetState("standing", 0)
		// Reset the command so it doesn't register again; except for blocking,
		// because that would un-block the player.
		if player.Command != "BLOCK" {
			player.Command = "NONE"
		}
		return player, enemy
	}
	// And now the normal commands.
	switch player.Command {
	case "NONE":
		if player.State == "blocking" {
			player.SetState("standing", 0)
		}
	case "BLOCK":
		if INTERRUPTABLE_STATES[player.State] && player.State != "blocking" {
			player.SetState("blocking", 0)
		}
	case "DODGE":
		// Dodges take time, unlike blocks which can be started at the last second.
		if INTERRUPTABLE_STATES[player.State] && player.Stamina >= DODGE_COST && enemy.StateDuration > DODGE_SPD {
			player.Stamina -= DODGE_COST
			if ATTACK_STATES[enemy.State] {
				enemy.SetState("standing", 0)
			}
		}
	case "SAVE":
		if player.State == "countered" {
			player.SetState("standing", 0)
			enemy.SetState("standing", 0)
		}
	case "LIGHT":
		if INTERRUPTABLE_STATES[player.State] && player.Stamina >= LIGHT_ATK_COST {
			player.Stamina -= LIGHT_ATK_COST
			// If the attack is going to interrupt a heavy attack, enter the interrupt mode.
			if enemy.State == "heavy attack" && enemy.StateDuration > LIGHT_ATK_SPD {
				key := INTERRUPT_RESOLVE_KEYS[random.Intn(4)]
				player.SetState("interrupting heavy"+key, 0)
				enemy.SetState("interrupted heavy"+key, 0)
				enemy.Life -= LIGHT_ATK_DMG
			} else {
				player.SetState("light attack", LIGHT_ATK_SPD)
			}
		}
	case "HEAVY":
		if INTERRUPTABLE_STATES[player.State] && player.Stamina >= HEAVY_ATK_COST {
			player.SetState("heavy attack", HEAVY_ATK_SPD)
			player.Stamina -= HEAVY_ATK_COST
		}
	}
	// Reset the command so it doesn't register again; except for blocking, because that would un-block the player.
	if player.Command != "BLOCK" {
		player.Command = "NONE"
	}
	return player, enemy
}

func catchInput(channel chan Message, stopChan chan bool) {
	for true {
		select {
		case <-channel:
		case <-stopChan:
			return
		}
	}
}
