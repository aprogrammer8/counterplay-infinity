/*
 * Copyright (c) 2018, Ryan Westlund.
 * This code is under the BSD 3-Clause license.
 */

package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveState(t *testing.T) {
	p1 := NewPlayer(nil, nil)
	p2 := NewPlayer(nil, nil)

	// Test light attack against no defense
	p1.Finished = "light attack"
	newp1, newp2 := resolveState(p1, p2)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100 - LIGHT_ATK_DMG, 100.0, "standing", 0, ""})

	// Test light attack canceling light attack
	p1.Finished = "light attack"
	p2.SetState("light attack", 5)
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100 - LIGHT_ATK_DMG, 100.0, "standing", 0, ""})
	p2 = NewPlayer(nil, nil)

	// Test light attack against a block too slow to counter
	p2.SetState("blocking", -1)
	p1.Finished = "light attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100, 100.0 - LIGHT_ATK_BLK_COST, "blocking", -1, ""})
	p2 = NewPlayer(nil, nil)

	// Test light attack against a block fast enough to counter
	p2.SetState("blocking", -(LIGHT_ATK_SPD - LIGHT_ATK_CNTR_WINDOW))
	p1.Finished = "light attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, Player{nil, nil, "NONE", 100, 100.0, "countered", 0, ""})
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100, 100.0 - LIGHT_ATK_BLK_COST, "counterattack", 30, ""})
	p2 = NewPlayer(nil, nil)

	// Test counterattack hitting
	p1.Finished = "counterattack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100 - LIGHT_ATK_CNTR_DMG, 100.0, "standing", 0, ""})

	// Test heavy attack against no defense
	p1.Finished = "heavy attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100 - HEAVY_ATK_DMG, 100.0, "standing", 0, ""})

	// Test blocked heavy attack
	p2.State = "blocking"
	p1.Finished = "heavy attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100 - HEAVY_ATK_BLKED_DMG, 100 - DODGE_COST, "blocking", 0, ""})
	p2 = NewPlayer(nil, nil)
}

func testResolveCommand(t *testing.T) {
	p1 := NewPlayer(nil, nil)
	p2 := NewPlayer(nil, nil)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Test light attack
	p1.Command = "LIGHT"
	newp1, newp2 := resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, Player{nil, nil, "NONE", 100, 100.0 - LIGHT_ATK_COST, "light attack", LIGHT_ATK_SPD, ""})

	// Test save
	p1.SetState("countered", 0)
	p2.SetState("countering", LIGHT_ATK_SPD)
	p1.Command = "SAVE"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, NewPlayer(nil, nil))

	// Test that save does nothing when not in a countered state
	p1.SetState("standing", 0)
	p2.SetState("light attack", LIGHT_ATK_SPD)
	p1.Command = "SAVE"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100, 100.0, "light attack", LIGHT_ATK_SPD, ""})

	// Test light attack interrupting a heavy
	p2.SetState("heavy attack", 100)
	p1.Command = "LIGHT"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1.Life, 100)
	assert.Equal(t, newp1.Stamina, 90)
	assert.Contains(t, newp1.State, "interruping heavy")
	assert.Equal(t, newp2.Life, 100-LIGHT_ATK_DMG)
	assert.Equal(t, newp2.Stamina, 100)
	assert.Contains(t, newp2.State, "interrupted heavy")

	// Test that the interrupt resolution keys don't do anything outside of interrupt mode
	p1.SetState("standing", 0)
	p2.SetState("standing", 0)
	p1.Command = "INTERRUPT_UP"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, NewPlayer(nil, nil))

	// Test interrupt resolution: the light attack player hits it first
	p1.SetState("interrupting heavy_up", 0)
	p2.SetState("interrupted heavy_up", 0)
	p1.Command = "INTERRUPT_UP"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, NewPlayer(nil, nil))

	// Test interrupt resolution: the light attack player hits the wrong button
	p1.SetState("interrupting heavy_up", 0)
	p2.SetState("interrupted heavy_up", 0)
	p1.Command = "INTERRUPT_DOWN"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, Player{nil, nil, "NONE", 100 - HEAVY_ATK_DMG, 100.0, "standing", 0, ""})
	assert.Equal(t, newp2, NewPlayer(nil, nil))

	// Test interrupt resolution: the heavy attack player hits it first
	p1.SetState("interrupted heavy_up", 0)
	p2.SetState("interrupting heavy_up", 0)
	p1.Command = "INTERRUPT_UP"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100 - HEAVY_ATK_DMG, 100.0, "standing", 0, ""})

	// Test interrupt resolution: the heavy attack player hits the wrong button
	p1.SetState("interrupted heavy_up", 0)
	p2.SetState("interrupting heavy_up", 0)
	p1.Command = "INTERRUPT_DOWN"
	newp1, newp2 = resolveCommand(p1, p2, random)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, NewPlayer(nil, nil))

	// Test dodging: too slow
	p1.SetState("standing", 0)
	p2.SetState("light attack", DODGE_SPD-1)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, Player{nil, nil, "NONE", 100, 100.0, "light attack", DODGE_SPD - 1, ""})

	// Test dodging: in time
	p1.SetState("standing", 0)
	p2.SetState("light attack", DODGE_SPD)
	assert.Equal(t, newp1, NewPlayer(nil, nil))
	assert.Equal(t, newp2, NewPlayer(nil, nil))
}
