package main

import (
	"math/rand"
	"time"
	"log"
)

func AttackBot(inputChan chan Message, updateChan chan Update) {
	time.Sleep(4500 * time.Millisecond)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	update := <- updateChan
	input := "NONE"
	attacks := []string{"LIGHT","HEAVY"}
	for update.Self.Life>0 && update.Enemy.Life>0 {

		if INTERRUPTABLE_STATES[update.Self.State] && update.Self.Stamina>=HEAVY_ATK_COST {
			input=attacks[random.Intn(2)]
			inputChan <- Message{Username:"AttackBot",Content:input,Command:""}
		}
		update = <- updateChan
	}
	log.Println("exiting")
}
