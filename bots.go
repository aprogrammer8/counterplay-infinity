// This file contains bots, which are functions spawned in goroutines for
// each battle against them. They connect with channels just like a player.

package main

import (
	"math/rand"
	"time"
)

// getBotByName is a convenience function to convert a bot's name to the function.
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

// AttackBot spams random attacks whenever it can.
func AttackBot(inputChan chan Message, updateChan chan Update) {
	// Don't attack during the countdown.
	time.Sleep(4500 * time.Millisecond)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	update := <-updateChan
	input := "NONE"
	attacks := []string{"LIGHT", "HEAVY"}
	// waitingState is a way to know whether the bot's last input has been acknowledged, by keeping track of the
	// state it was in when it issued the command. We need this to stop it from sending the command more than once,
	// because that could cause it to auto-lose interrupt it initiates. We won't send a command if waitingState is
	// set.
	waitingState := ""
	for update.Self.Life > 0 && update.Enemy.Life > 0 {
		// It doesn't do any attacks unless it has enough stamina for a heavy,
		// because otherwise it would get stuck spamming light attacks at low stamina.
		if INTERRUPTABLE_STATES[update.Self.State] && update.Self.Stamina >= HEAVY_ATK_COST && update.Self.State != waitingState {
			// Don't send another command if we're still waiting for our state to change.
			// Don't do light attacks into a prepared block.
			if update.Enemy.State == "blocking" {
				input = "HEAVY"
			} else {
				input = attacks[random.Intn(2)]
			}
			inputChan <- Message{Username: "AttackBot", Content: input, Command: ""}
			waitingState = update.Self.State
		}
		update = <-updateChan
		// If our state has changed, we can stop waiting and it's safe to send commands again.
		if update.Self.State != waitingState {
			waitingState = ""
		}
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
	// waitingState is a way to know whether the bot's last input has been acknowledged, by keeping track of the
	// state it was in when it issued the command. We need this to stop it from sending the command more than once,
	// because that could cause it to auto-lose interrupt it initiates. We won't send a command if waitingState is
	// set.
	waitingState := ""
	for update.Self.Life > 0 && update.Enemy.Life > 0 {
		// It doesn't do any attacks unless it has enough stamina for a heavy,
		// because otherwise it would get stuck spamming light attacks at low stamina.
		if INTERRUPTABLE_STATES[update.Self.State] && update.Self.Stamina >= HEAVY_ATK_COST && update.Self.State != waitingState {
			// Don't do light attacks into a prepared block.
			if update.Enemy.State == "blocking" {
				input = "HEAVY"
			} else {
				input = attacks[random.Intn(2)]
			}
			// Wait a small randomized delay before acting.
			time.Sleep(time.Duration((200 + random.Intn(133))) * time.Millisecond)
			inputChan <- Message{Username: "AttackBotSlow", Content: input, Command: ""}
		}
		update = <-updateChan
		// If our state has changed, we can stop waiting and it's safe to send commands again.
		if update.Self.State != waitingState {
			waitingState = ""
		}
	}
}
