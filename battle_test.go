/*
 * Copyright (c) 2018, Ryan Westlund.
 * This code is under the BSD 3-Clause license.
 */

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveState(t *testing.T) {
	p1 := Player{"p1", nil, nil, "NONE", 100, 100.0, "standing", 0, ""}
	p2 := Player{"p2", nil, nil, "NONE", 100, 100.0, "standing", 0, ""}

	// Test light attack against no defense
	p1.Finished = "light attack"
	newp1, newp2 := resolveState(p1, p2)
	assert.Equal(t, newp1, Player{"p1", nil, nil, "NONE", 100, 100.0, "standing", 0, ""})
	assert.Equal(t, newp2, Player{"p2", nil, nil, "NONE", 97, 100.0, "standing", 0, ""})

	// Test light attack against a preemptive block
	p2.State = "blocking"
	p2.StateDuration = -50
	p1.Finished = "light attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, Player{"p1", nil, nil, "NONE", 100, 100.0, "standing", 0, ""})
	assert.Equal(t, newp2, Player{"p2", nil, nil, "NONE", 100, 88.0, "blocking", -50, ""})
	p2 = Player{"p2", nil, nil, "NONE", 100, 100.0, "standing", 0, ""}

	// Test light attack against a reactive block
	p2.State = "blocking"
	p2.StateDuration = -49
	p1.Finished = "light attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, Player{"p1", nil, nil, "NONE", 100, 100.0, "countered", 0, ""})
	assert.Equal(t, newp2, Player{"p2", nil, nil, "NONE", 100, 88.0, "counterattack", 30, ""})
	p2 = Player{"p2", nil, nil, "NONE", 100, 100.0, "standing", 0, ""}

	// Test counterattack hitting
	p1.Finished = "counterattack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, Player{"p1", nil, nil, "NONE", 100, 100.0, "standing", 0, ""})
	assert.Equal(t, newp2, Player{"p2", nil, nil, "NONE", 97, 100.0, "standing", 0, ""})

	// Test heavy attack against no defense
	p1.Finished = "heavy attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, Player{"p1", nil, nil, "NONE", 100, 100.0, "standing", 0, ""})
	assert.Equal(t, newp2, Player{"p2", nil, nil, "NONE", 94, 100.0, "standing", 0, ""})

	// Test blocked heavy attack
	p2.State = "blocking"
	p1.Finished = "heavy attack"
	newp1, newp2 = resolveState(p1, p2)
	assert.Equal(t, newp1, Player{"p1", nil, nil, "NONE", 100, 100.0, "standing", 0, ""})
	assert.Equal(t, newp2, Player{"p2", nil, nil, "NONE", 98, 80.0, "blocking", 0, ""})
	p2 = Player{"p2", nil, nil, "NONE", 100, 100.0, "standing", 0, ""}
}
