// This file contains bots, which are functions that spawned in goroutines for each battle against them. They connect with channels just like a player.

package main

import (
	"math/rand"
	"time"
)

func getBotByName(bot string) func(chan Message, chan Update) {
	switch bot {
	case "AttackBot":
		return AttackBot
	case "AttackBotSlow":
		return AttackBotSlow
	default:
		return nil
	}
}
//var bots = map[string]*func(chan Message, chan Update) {"AttackBot": &AttackBot, "AttackBotSlow": &AttackBotSlow}


// AttackBot spams random attacks whenever it can.
func AttackBot(inputChan chan Message, updateChan chan Update) {
	// Don't attack during the countdown.
	time.Sleep(4500 * time.Millisecond)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	update := <-updateChan
	input := "NONE"
	attacks := []string{"LIGHT", "HEAVY"}
	for update.Self.Life > 0 && update.Enemy.Life > 0 {
		// It doesn't do any attacks unless it has enough stamina for a heavy, because otherwise it would get stuck spamming light attacks at low stamina.
		if INTERRUPTABLE_STATES[update.Self.State] && update.Self.Stamina >= HEAVY_ATK_COST {
			input = attacks[random.Intn(2)]
			inputChan <- Message{Username: "AttackBot", Content: input, Command: ""}
		}
		update = <-updateChan
	}
}

// AttackBotSlow is like AttackBot, but doesn't have perfect reaction time.
func AttackBotSlow(inputChan chan Message, updateChan chan Update) {
	// Don't attack during the countdown.
	time.Sleep(4500 * time.Millisecond)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	update := <-updateChan
	input := "NONE"
	attacks := []string{"LIGHT", "HEAVY"}
	for update.Self.Life > 0 && update.Enemy.Life > 0 {
		// It doesn't do any attacks unless it has enough stamina for a heavy, because otherwise it would get stuck spamming light attacks at low stamina.
		if INTERRUPTABLE_STATES[update.Self.State] && update.Self.Stamina >= HEAVY_ATK_COST {
			input = attacks[random.Intn(2)]
			// Wait a small randomized delay before acting.
			time.Sleep(time.Duration((200 + random.Intn(133))) * time.Millisecond)
			inputChan <- Message{Username: "AttackBot", Content: input, Command: ""}
		}
		update = <-updateChan
	}
}
