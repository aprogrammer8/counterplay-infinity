package main

import (
	"log"
	"math/rand"
	"time"
)

// AttackBot is spawned in a goroutine for each battle with AttackBot. It connects with channels just like a player.
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
	log.Println("exiting")
}
