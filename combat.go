package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
)

func combatEntity(world *World, conn *ConnectionData, db *sql.DB) int {
	// player rawdamage formula: weapon.baseDam + str * strMultConst
	// entity rawdamage formula: entity.baseDam + str * strMultConst
	// player & entity random variance: [0.85, 1.15]
	// player finaldamage formula: (p.rawdamage / (100 + entity.baseDef) * randfactor)
	// entity finaldamage formula: (e.rawdamage / (100 + player.baseDef) * randfactor)
	// player dodgechance formula: (player.agi / sum(player.agi + entity.agi)) * maxDodge
	// entity dodgechance formula: (entity.agi / sum(player.agi + entity.agi)) * maxDodge
	// where maxDodge is (0, 100) a const

	chr := *conn.session.character
	tEnt := world.entities[*conn.session.character.targetID]
	var chrTotalDef int
	for eq, i := range chr.equipment {
		if eq != "" {
			chrTotalDef += world.ItemTemplates[world.items[i].templateID].baseDef
		}
	}

	strMultConst := 3
	pRawDam := chr.equipment["mainhand"] + chr.getEffectiveStat("str")*strMultConst
	pRandFactor := rand.Float64()*0.3 + 0.85
	eRandFactor := rand.Float64()*0.3 + 0.85
	pFinalDam := int((float64(pRawDam) * (float64(100) / float64(100+world.EntityTemplates[tEnt.templateID].baseDef))) * pRandFactor)
	eRawDam := world.EntityTemplates[tEnt.templateID].baseDam + world.EntityTemplates[tEnt.templateID].stats.Str*strMultConst
	eFinalDam := int((float64(eRawDam) * (float64(100) / float64(100+chrTotalDef))) * eRandFactor)
	maxDodge := 80
	pDodgeChance := int(float64(chr.baseStats.Agi) / float64(chr.baseStats.Agi+world.EntityTemplates[tEnt.templateID].stats.Agi) * float64(maxDodge))
	eDodgeChance := int(float64(world.EntityTemplates[tEnt.templateID].stats.Agi) / float64(chr.baseStats.Agi+world.EntityTemplates[tEnt.templateID].stats.Agi) * float64(maxDodge))
	// this method causes the same dodge chance, leaving it here if the new one doesnt work
	/*	if chr.baseStats.Agi > world.EntityTemplates[tEnt.templateID].stats.Agi {
			pDodgeChance = int(float64(chr.baseStats.Agi-world.EntityTemplates[tEnt.templateID].stats.Agi) / float64(chr.baseStats.Agi+world.EntityTemplates[tEnt.templateID].stats.Agi) * 100)
			eDodgeChance = int(float64(world.EntityTemplates[tEnt.templateID].stats.Agi-chr.baseStats.Agi) / float64(chr.baseStats.Agi+world.EntityTemplates[tEnt.templateID].stats.Agi) * -100)
		} else if world.EntityTemplates[tEnt.templateID].stats.Agi > chr.baseStats.Agi {
			eDodgeChance = int(float64(world.EntityTemplates[tEnt.templateID].stats.Agi-chr.baseStats.Agi) / float64(chr.baseStats.Agi+world.EntityTemplates[tEnt.templateID].stats.Agi) * 100)
			pDodgeChance = int(float64(chr.baseStats.Agi-world.EntityTemplates[tEnt.templateID].stats.Agi) / float64(chr.baseStats.Agi+world.EntityTemplates[tEnt.templateID].stats.Agi) * -100)
		} else {
			eDodgeChance = 0
			pDodgeChance = 0
		}
		eDodgeChance = Clamp(eDodgeChance, 0, 90)
		pDodgeChance = Clamp(pDodgeChance, 0, 90)*/
	fmt.Println(eDodgeChance)
	fmt.Println(pDodgeChance)
	if rand.Intn(100) <= eDodgeChance {
		conn.store.Write([]byte("\nThe " + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + " dodges your attack! (" + strconv.Itoa(world.entities[*conn.session.character.targetID].hp) + ")" + "\n"))
	} else {
		world.entities[*conn.session.character.targetID].hp -= pFinalDam
		conn.store.Write([]byte("\nYou damage the " + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + " for " + strconv.Itoa(pFinalDam) + " (" + strconv.Itoa(world.entities[*conn.session.character.targetID].hp) + ")" + "\n"))
	}
	if rand.Intn(100) <= pDodgeChance {
		conn.store.Write([]byte("You dodge the " + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + "'s attack! (" + strconv.Itoa(conn.session.character.hp) + ")" + "\n"))
	} else {
		conn.session.character.hp -= eFinalDam
		conn.store.Write([]byte("The " + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + " damages you for " + strconv.Itoa(eFinalDam) + " (" + strconv.Itoa(conn.session.character.hp) + ")" + "\n"))
	}
	if conn.session.character.hp <= 0 {
		conn.store.Write([]byte("\nYou died!\n\n> "))
		world.entities[*conn.session.character.targetID].inCombat = false
		world.entities[*conn.session.character.targetID].targetID = nil
		conn.session.character.inCombat = false
		conn.session.character.targetID = nil
		conn.session.character.targetType = nil
		return 1
	}
	if world.entities[*conn.session.character.targetID].hp <= 0 {
		conn.store.Write([]byte("\nYou killed a " + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + "!"))
		c := rand.Intn(100)
		conn.session.character.coins += c
		fmt.Println(world.entities)
		conn.store.Write([]byte("\nYou loot the " + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + "'s body and find " + strconv.Itoa(c) + " coins!\n\n> "))
		db.Exec("DELETE FROM entities WHERE id = ?", *conn.session.character.targetID)
		for i, id := range world.nodeList[conn.session.character.locationID].entityIDs {
			if world.entities[id] == world.entities[*conn.session.character.targetID] {
				world.nodeList[conn.session.character.locationID].entityIDs = slices.Delete(world.nodeList[conn.session.character.locationID].entityIDs, i, i+1)
			}
		}
		conn.session.character.inCombat = false
		conn.session.character.targetID = nil
		conn.session.character.targetType = nil
		return 1
	}
	return 0
}

func combatPlayer(world *World, conn *ConnectionData) int {
	// player rawdamage formula: weapon.baseDam + str * strMultConst
	// entity rawdamage formula: entity.baseDam + str * strMultConst
	// player & entity random variance: [0.85, 1.15]
	// player finaldamage formula: (p.rawdamage / (100 + entity.baseDef) * randfactor)
	// entity finaldamage formula: (e.rawdamage / (100 + player.baseDef) * randfactor)
	// player dodgechance formula: (player.agi / sum(player.agi + entity.agi)) * maxDodge
	// entity dodgechance formula: (entity.agi / sum(player.agi + entity.agi)) * maxDodge
	// where maxDodge is (0, 100) a const

	p1Chr := conn.session.character
	p2Chr := world.characters[*conn.session.character.targetID]

	if p1Chr.worldID > p2Chr.worldID {
		return 1
	}

	var p1ChrTotalDef int
	var p2ChrTotalDef int

	for _, i := range p1Chr.equipment {
		p1ChrTotalDef += world.ItemTemplates[world.items[i].templateID].baseDef
	}
	for _, i := range p2Chr.equipment {
		p2ChrTotalDef += world.ItemTemplates[world.items[i].templateID].baseDef
	}

	strMultConst := 3
	p1RandFactor := rand.Float64()*0.3 + 0.85
	p2RandFactor := rand.Float64()*0.3 + 0.85

	p1RawDam := p1Chr.equipment["mainhand"] + p1Chr.getEffectiveStat("str")*strMultConst
	p1FinalDam := int((float64(p1RawDam) * (float64(100) / float64(100+p2ChrTotalDef))) * p1RandFactor)

	p2RawDam := p2Chr.equipment["mainhand"] + p2Chr.getEffectiveStat("str")*strMultConst
	p2FinalDam := int((float64(p2RawDam) * (float64(100) / float64(100+p1ChrTotalDef))) * p2RandFactor)

	maxDodge := 80
	p1DodgeChance := int(float64(p1Chr.baseStats.Agi) / float64(p1Chr.baseStats.Agi+p2Chr.baseStats.Agi) * float64(maxDodge))
	p2DodgeChance := int(float64(p2Chr.baseStats.Agi) / float64(p1Chr.baseStats.Agi+p2Chr.baseStats.Agi) * float64(maxDodge))

	if rand.Intn(100) <= p2DodgeChance {
		p2Chr.conn.store.Write([]byte("\nYou dodge " + p1Chr.conn.session.username + "'s attack! (" + strconv.Itoa(p2Chr.hp) + ")"))
		p1Chr.conn.store.Write([]byte("\n" + p2Chr.conn.session.username + " dodges your attack! (" + strconv.Itoa(p2Chr.hp) + ")"))
	} else {
		p2Chr.hp -= p1FinalDam
		p1Chr.conn.store.Write([]byte("\nYou damage " + p2Chr.conn.session.username + " for " + strconv.Itoa(p1FinalDam) + " (" + strconv.Itoa(p2Chr.hp) + ")"))
		p2Chr.conn.store.Write([]byte("\n" + p1Chr.conn.session.username + " damages you for " + strconv.Itoa(p1FinalDam) + " (" + strconv.Itoa(p2Chr.hp) + ")"))
	}

	if rand.Intn(100) <= p1DodgeChance {
		p1Chr.conn.store.Write([]byte("\nYou dodge " + p2Chr.conn.session.username + "'s attack! (" + strconv.Itoa(p1Chr.hp) + ")\n"))
		p2Chr.conn.store.Write([]byte("\n" + p1Chr.conn.session.username + " dodges your attack! (" + strconv.Itoa(p1Chr.hp) + ")\n"))
	} else {
		p1Chr.hp -= p2FinalDam
		p1Chr.conn.store.Write([]byte("\n" + p2Chr.conn.session.username + " damages you for " + strconv.Itoa(p2FinalDam) + " (" + strconv.Itoa(p1Chr.hp) + ")\n"))
		p2Chr.conn.store.Write([]byte("\nYou damage " + p1Chr.conn.session.username + " for " + strconv.Itoa(p2FinalDam) + " (" + strconv.Itoa(p1Chr.hp) + ")\n"))
	}

	if p2Chr.hp <= 0 {
		p2Chr.inCombat = false
		p2Chr.targetID = nil
		p2Chr.targetType = nil
		p1Chr.inCombat = false
		p1Chr.targetID = nil
		p1Chr.targetType = nil
		p2Chr.conn.store.Write([]byte("\nYou died!"))
		if p2Chr.coins == 0 {
			conn.store.Write([]byte("\n" + p2Chr.conn.session.username + "didn't have any coins for you to loot!"))
		} else {
			c := rand.Intn(p2Chr.coins)
			conn.store.Write([]byte("\nYou loot " + p2Chr.conn.session.username + "'s body to steal " + strconv.Itoa(c) + " coins!"))
			p2Chr.conn.store.Write([]byte("\n" + conn.session.username + " steals " + strconv.Itoa(c) + " coins from you!"))
			p1Chr.coins += c
			p2Chr.coins -= c
		}
		p2Chr.conn.store.Write([]byte("\nYou are teleported to spawn!\n\n> "))
		conn.store.Write([]byte("\nYou killed " + p2Chr.conn.session.username + "!\n\n> "))
		p2Chr.hp = p2Chr.maxHp
		HandleMovement(p2Chr.conn, world)
		fmt.Println("\n\n> ")
		return 1
	} else if p1Chr.hp <= 0 {
		p2Chr.inCombat = false
		p2Chr.targetID = nil
		p2Chr.targetType = nil
		p1Chr.inCombat = false
		p1Chr.targetID = nil
		p1Chr.targetType = nil
		p2Chr.conn.store.Write([]byte("\nYou killed " + conn.session.username + "!\n\n> "))
		conn.store.Write([]byte("\nYou died!"))
		if p1Chr.coins == 0 {
			p1Chr.conn.store.Write([]byte("\n" + conn.session.username + "didn't have any coins for you to loot!"))
		} else {
			c := rand.Intn(p1Chr.coins)
			p2Chr.conn.store.Write([]byte("\nYou loot " + conn.session.username + "'s body to steal " + strconv.Itoa(c) + " coins!"))
			p1Chr.conn.store.Write([]byte("\n" + p2Chr.conn.session.username + " steals " + strconv.Itoa(c) + " coins from you!"))
			p2Chr.coins += c
			p1Chr.coins -= c
		}
		p1Chr.conn.store.Write([]byte("\nYou are teleported to spawn!\n\n> "))
		conn.store.Write([]byte("\nYou killed " + p1Chr.conn.session.username + "!\n\n> "))
		conn.store.Write([]byte("\nYou are teleported to spawn!"))
		p1Chr.hp = p1Chr.maxHp
		HandleMovement(conn, world)
		fmt.Println("\n\n> ")
		return 1
	}
	return 0
}
