This is a 1v1 fighting game with no graphics and no movement. The battle screen consists only of a HUD, which includes for both players a green life bar, a yellow stamina bar, a black state duration bar (which shows how long until the player exits their current state and returns to the default standing state), and
some icons below that indicate the player's current state.

The Rules
=========
There are currently five controls in the game: a light attack (mapped to q), a heavy attack (mapped to w), a block (mapped to space), a dodge (mapped to shift), and a 'save' mapped to control.
- The light attack is quick to land, costs a small amount of stamina and does a small amount of damage. If the enemy is doing their own attack when this lands, theirs is canceled. If the enemy blocks barely in time, they will lose a small amount of stamina but not take damage. If they block easily in time, they will **counter** your attack, avoiding damage and initiating their own, faster attack. To avoid being hit by the counterattack, you must save before it lands.
- The heavy attack is slow and costs more stamina but does much more damage. If it hits an attacking enemy, their attack will be canceled. If it hits a blocking opponent, they will still receive a small amount of damage and lose a lot of stamina. It can be dodged to avoid all damage, but dodging costs a lot of stamina and takes time, whereas blocking is instant. If, after you've already started a heavy attack, the enemy initiates a light attack that will land before your heavy attack, then instad of being canceled you will enter **interrupt mode**. You take damage from the light attack, and an arrow key will be displayed on screen. If you hit it first, your heavy attack hits too. If they hit it first, the heavy attack misses. Hitting the wrong button counts as hitting it second.
- The block is instant and costs no stamina by itself, but it can only be used if you are in an interruptable state (not doing an attack).
- The dodge takes time to happen, costs the same amount of stamina regardless of what you dodge, and still requires you to be in a interruptable state.

The Icons
=========
- A shield symbol under your state duration bar means you are blocking.
- A spear symbol means you are doing a light attack.
- A sword symbol means a heavy attack.
- A shield on the left and a spear to the right of it means you're countering a light attack.
- A spear on the left and a shield to the right of it means your light attack is being countered.
- A sword symbol or spear symbol with an arrow next to it means you either had your heavy attack interrupted or are interrupting the enemy's heavy attack, depending on which symbol is on whose side. The arrow is the one you must press to resolve the interrupt in your favor.

The Stats
=========
- Both players start with 100 life and 100 stamina.
- Stamina regenerates by 0.1 points per mainloop cycle (which is 1 centisecond).
- Light attack: deals 3 damage, costs 10 stamina, takes 50 cycles to land, and costs 12 stamina to block.
- Counterattack: deals 3 damage, cost 12 stamina, takes 30 cycles to land, and costs nothing to save against.
- Heavy attack: deals 6 damage, costs 15 stamina, takes 100 cycles to land, costs 20 stamina to block, and deals 2 damage if blocked.
- Dodge: costs 20 stamina, takes 30 cycles.

License
=======
This code is under the BSD 3-Clause license. See the LICENSE file for the full text.
