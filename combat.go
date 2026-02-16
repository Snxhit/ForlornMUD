package main

import (
	"database/sql"
	"math"
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

	if conn.session.character.targetID == nil {
		return 1
	}
	chr := *conn.session.character
	tEnt := world.entities[*conn.session.character.targetID]
	cMin := world.EntityTemplates[tEnt.templateID].cMin
	cMax := world.EntityTemplates[tEnt.templateID].cMax
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

	pPower := chr.getEffectiveStat("str")
	ePower := world.EntityTemplates[tEnt.templateID].stats.Str + world.EntityTemplates[tEnt.templateID].baseDef
	powerRatio := float64(ePower) / float64(pPower+ePower)
	exp := float64(world.EntityTemplates[tEnt.templateID].baseExp)
	exp *= powerRatio
	exp *= calcExpMultiplier(world.EntityTemplates[tEnt.templateID].level - chr.level)
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
	if rand.Intn(100) <= eDodgeChance {
		conn.store.Write([]byte("\x1b[2K\r  The " + color(conn, "cyan", "tp") + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + color(conn, "red", "tp") + " dodges" + color(conn, "reset", "reset") + " your attack! (" + strconv.Itoa(world.entities[*conn.session.character.targetID].hp) + ")" + "\n"))
	} else {
		world.entities[*conn.session.character.targetID].hp -= pFinalDam
		conn.store.Write([]byte("\x1b[2K\r  You damage the " + color(conn, "cyan", "tp") + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + color(conn, "reset", "reset") + " for " + color(conn, "green", "tp") + strconv.Itoa(pFinalDam) + color(conn, "reset", "reset") + " (" + strconv.Itoa(world.entities[*conn.session.character.targetID].hp) + ")" + "\n"))
	}
	if rand.Intn(100) <= pDodgeChance {
		conn.store.Write([]byte("\x1b[2K\r  You " + color(conn, "green", "tp") + "dodge" + color(conn, "reset", "reset") + " the " + color(conn, "cyan", "tp") + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + color(conn, "reset", "reset") + "'s attack! (" + strconv.Itoa(conn.session.character.hp) + ")" + "\n\n> "))
	} else {
		conn.session.character.hp -= eFinalDam
		conn.store.Write([]byte("\x1b[2K\r  The " + color(conn, "cyan", "tp") + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + color(conn, "reset", "reset") + " damages you for " + color(conn, "red", "tp") + strconv.Itoa(eFinalDam) + color(conn, "reset", "reset") + " (" + strconv.Itoa(conn.session.character.hp) + ")" + "\n\n> "))
	}
	if conn.session.character.hp <= 0 {
		conn.store.Write([]byte("\n\x01COMBAT type:entity hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"None\" enemyHp:0 enemyMaxHp:0\n"))
		conn.store.Write([]byte(color(conn, "red", "tp") + "\x1b[2K\r\n  You died!" + color(conn, "reset", "reset") + ""))
		conn.store.Write([]byte(color(conn, "green", "tp") + "\n  You are teleported to spawn!" + color(conn, "reset", "reset") + "\n\n> "))
		conn.session.character.hp = conn.session.character.maxHp
		conn.session.character.locationID = 0
		world.entities[*conn.session.character.targetID].inCombat = false
		world.entities[*conn.session.character.targetID].targetID = nil
		conn.session.character.inCombat = false
		conn.session.character.targetID = nil
		conn.session.character.targetType = nil
		HandleMovement(conn, world)
		conn.store.Write([]byte("\n> "))
		return 1
	}
	if world.entities[*conn.session.character.targetID].hp <= 0 {
		conn.store.Write([]byte("\n\x01COMBAT type:entity hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"None\" enemyHp:0 enemyMaxHp:0\n"))
		conn.store.Write([]byte("\x1b[2K\r\n  You " + color(conn, "red", "tp") + "killed " + color(conn, "reset", "reset") + "a " + color(conn, "cyan", "tp") + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + color(conn, "reset", "reset") + "!"))
		c := rand.Intn(cMax-cMin) + cMin
		conn.session.character.coins += c
		conn.store.Write([]byte("\n  You " + color(conn, "yellow", "tp") + "loot" + color(conn, "reset", "reset") + " the " + color(conn, "cyan", "tp") + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + color(conn, "reset", "reset") + "'s body and find " + color(conn, "yellow", "tp") + strconv.Itoa(c) + color(conn, "reset", "reset") + " coins!"))
		for _, d := range world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].dropTable {
			if rand.Intn(100) <= d.chance {
				var qty int
				if d.max-d.min == 0 {
					qty = d.min
				} else {
					qty = rand.Intn(d.max-d.min) + d.min
				}
				conn.store.Write([]byte("\n  You find " + color(conn, "yellow", "tp") + strconv.Itoa(qty) + "x " + color(conn, "cyan", "tp") + world.ItemTemplates[d.itemTemplateID].name + color(conn, "reset", "reset") + " on the corpse!"))
				/*for range qty {
					CreateAndInsertItem(conn, world, db, d.itemTemplateID)
				}*/
				CreateAndInsertItemBatched(conn, world, db, d.itemTemplateID, qty)
			}
		}
		conn.store.Write([]byte("\n"))
		pLvl := conn.session.character.level
		pExp := conn.session.character.exp
		if math.Floor((float64(pExp)+exp)/100.0) > float64(pLvl) {
			lvls := int(math.Floor(exp / 100.0))
			trains := 5 * int(lvls)
			conn.store.Write([]byte("\n  You gain " + color(conn, "blue", "tp") + strconv.Itoa(int(exp)) + color(conn, "reset", "reset") + " exp from this fight!\n"))
			conn.store.Write([]byte("\n  Congrats!"))
			conn.store.Write([]byte("\n  You gain " + color(conn, "yellow", "tp") + strconv.Itoa(lvls) + " level(s)" + color(conn, "reset", "reset") + " from this fight!"))
			conn.store.Write([]byte("\n  You now have " + color(conn, "yellow", "tp") + strconv.Itoa(trains) + color(conn, "cyan", "tp") + " more trains" + color(conn, "reset", "reset") + "!\n\n> "))
			conn.session.character.level += lvls
			conn.session.character.trains += trains
		} else {
			conn.store.Write([]byte("\n  You gain " + color(conn, "blue", "tp") + strconv.Itoa(int(exp)) + color(conn, "reset", "reset") + " exp from this fight!\n\n> "))
		}
		conn.session.character.exp += int(exp)
		if world.merchants[*conn.session.character.targetID] != nil {
			db.Exec("DELETE FROM merchants WHERE id = ?", *conn.session.character.targetID)
			db.Exec("DELETE FROM merchant_list WHERE id = ?", *conn.session.character.targetID)
			delete(world.merchants, *conn.session.character.targetID)
		}
		db.Exec("DELETE FROM entities WHERE id = ?", *conn.session.character.targetID)
		for i, id := range world.nodeList[conn.session.character.locationID].entityIDs {
			if world.entities[id] == world.entities[*conn.session.character.targetID] {
				world.nodeList[conn.session.character.locationID].entityIDs = slices.Delete(world.nodeList[conn.session.character.locationID].entityIDs, i, i+1)
				break
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

	if conn.session.character.targetID == nil {
		return 1
	}

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
		p2Chr.conn.store.Write([]byte("\x1b[2K\r  You dodge " + color(p2Chr.conn, "cyan", "tp") + p1Chr.conn.session.username + color(p2Chr.conn, "reset", "reset") + "'s attack! (" + strconv.Itoa(p2Chr.hp) + ")\n"))
		p1Chr.conn.store.Write([]byte("\x1b[2K\r  " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "red", "tp") + " dodges" + color(p1Chr.conn, "reset", "reset") + " your attack! (" + strconv.Itoa(p2Chr.hp) + ")\n"))
	} else {
		p2Chr.hp -= p1FinalDam
		p1Chr.conn.store.Write([]byte("\x1b[2K\r  You damage " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "reset", "reset") + " for " + color(p1Chr.conn, "green", "tp") + strconv.Itoa(p1FinalDam) + color(p1Chr.conn, "reset", "reset") + " (" + strconv.Itoa(p2Chr.hp) + ")\n"))
		p2Chr.conn.store.Write([]byte("\x1b[2K\r  " + color(p2Chr.conn, "cyan", "tp") + p1Chr.conn.session.username + color(p2Chr.conn, "reset", "reset") + " damages you for " + color(p2Chr.conn, "red", "tp") + strconv.Itoa(p1FinalDam) + color(p2Chr.conn, "reset", "reset") + " (" + strconv.Itoa(p2Chr.hp) + ")\n"))
	}

	if rand.Intn(100) <= p1DodgeChance {
		p1Chr.conn.store.Write([]byte("\x1b[2K\r  You dodge " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "reset", "reset") + "'s attack! (" + strconv.Itoa(p1Chr.hp) + ")\n\n> "))
		p2Chr.conn.store.Write([]byte("\x1b[2K\r  " + color(p2Chr.conn, "cyan", "tp") + p1Chr.conn.session.username + color(p2Chr.conn, "red", "tp") + " dodges" + color(p2Chr.conn, "reset", "reset") + " your attack! (" + strconv.Itoa(p1Chr.hp) + ")\n\n> "))
	} else {
		p1Chr.hp -= p2FinalDam
		p1Chr.conn.store.Write([]byte("\x1b[2K\r  " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "reset", "reset") + " damages you for " + color(p1Chr.conn, "red", "tp") + strconv.Itoa(p2FinalDam) + color(p1Chr.conn, "reset", "reset") + " (" + strconv.Itoa(p1Chr.hp) + ")\n\n> "))
		p2Chr.conn.store.Write([]byte("\x1b[2K\r  You damage " + color(p2Chr.conn, "cyan", "tp") + p1Chr.conn.session.username + color(p2Chr.conn, "reset", "reset") + " for " + color(p2Chr.conn, "green", "tp") + strconv.Itoa(p2FinalDam) + color(p2Chr.conn, "reset", "reset") + " (" + strconv.Itoa(p1Chr.hp) + ")\n\n> "))
	}

	if p2Chr.hp <= 0 {
		p1Chr.conn.store.Write([]byte("\n\x01COMBAT type:player hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"None\" enemyHp:0 enemyMaxHp:0\n"))
		p2Chr.conn.store.Write([]byte("\n\x01COMBAT type:player hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"None\" enemyHp:0 enemyMaxHp:0\n"))
		p2Chr.inCombat = false
		p2Chr.targetID = nil
		p2Chr.targetType = nil
		p1Chr.inCombat = false
		p1Chr.targetID = nil
		p1Chr.targetType = nil
		p2Chr.conn.store.Write([]byte(color(p2Chr.conn, "red", "tp") + "\x1b[2K\r\n  You died!" + color(p2Chr.conn, "reset", "reset")))
		if p2Chr.coins == 0 {
			conn.store.Write([]byte("\x1b[2K\r\n  " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "reset", "reset") + " didn't have any coins for you to loot!"))
		} else {
			c := rand.Intn(p2Chr.coins)
			conn.store.Write([]byte("\n  You" + color(p1Chr.conn, "yellow", "tp") + " loot " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "reset", "reset") + "'s body to steal " + color(p1Chr.conn, "yellow", "tp") + strconv.Itoa(c) + color(p1Chr.conn, "reset", "reset") + " coins!"))
			p2Chr.conn.store.Write([]byte("\n  " + color(p2Chr.conn, "cyan", "tp") + conn.session.username + color(p2Chr.conn, "reset", "reset") + " steals " + color(p2Chr.conn, "yelow", "tp") + strconv.Itoa(c) + color(p2Chr.conn, "reset", "reset") + " coins from you!"))
			p1Chr.coins += c
			p2Chr.coins -= c
		}
		p2Chr.conn.store.Write([]byte(color(p2Chr.conn, "green", "tp") + "\n  You are teleported to spawn!" + color(p2Chr.conn, "reset", "reset") + "\n\n> "))
		conn.store.Write([]byte("\n  You " + color(p1Chr.conn, "red", "tp") + "killed " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "reset", "reset") + "!"))
		p2Chr.hp = p2Chr.maxHp
		HandleMovement(p2Chr.conn, world)
		p1Chr.conn.store.Write([]byte("\n\n> "))
		p2Chr.conn.store.Write([]byte("\n\n> "))
		return 1
	} else if p1Chr.hp <= 0 {
		p1Chr.conn.store.Write([]byte("\n\x01COMBAT type:player hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"None\" enemyHp:0 enemyMaxHp:0\n"))
		p2Chr.conn.store.Write([]byte("\n\x01COMBAT type:player hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"None\" enemyHp:0 enemyMaxHp:0\n"))
		p2Chr.inCombat = false
		p2Chr.targetID = nil
		p2Chr.targetType = nil
		p1Chr.inCombat = false
		p1Chr.targetID = nil
		p1Chr.targetType = nil
		p1Chr.conn.store.Write([]byte(color(p1Chr.conn, "red", "tp") + "\x1b[2K\r\n  You died!" + color(p1Chr.conn, "reset", "reset")))
		if p1Chr.coins == 0 {
			p2Chr.conn.store.Write([]byte("\x1b[2K\r\n  " + color(p2Chr.conn, "cyan", "tp") + p1Chr.conn.session.username + color(p2Chr.conn, "reset", "reset") + " didn't have any coins for you to loot!"))
		} else {
			c := rand.Intn(p1Chr.coins)
			p2Chr.conn.store.Write([]byte("\n  You" + color(p2Chr.conn, "yellow", "tp") + " loot " + color(p2Chr.conn, "cyan", "tp") + p1Chr.conn.session.username + color(p2Chr.conn, "reset", "reset") + "'s body to steal " + color(p2Chr.conn, "yellow", "tp") + strconv.Itoa(c) + color(p2Chr.conn, "reset", "reset") + " coins!"))
			p1Chr.conn.store.Write([]byte("\n  " + color(p1Chr.conn, "cyan", "tp") + p2Chr.conn.session.username + color(p1Chr.conn, "reset", "reset") + " steals " + color(p1Chr.conn, "yelow", "tp") + strconv.Itoa(c) + color(p1Chr.conn, "reset", "reset") + " coins from you!"))
			p2Chr.coins += c
			p1Chr.coins -= c
		}
		p1Chr.conn.store.Write([]byte(color(p1Chr.conn, "green", "tp") + "\n  You are teleported to spawn!" + color(p1Chr.conn, "reset", "reset") + "\n\n> "))
		p2Chr.conn.store.Write([]byte("\n  You " + color(p2Chr.conn, "red", "tp") + "killed " + color(p2Chr.conn, "cyan", "tp") + p1Chr.conn.session.username + color(p2Chr.conn, "reset", "reset") + "!"))
		p1Chr.hp = p1Chr.maxHp
		HandleMovement(p1Chr.conn, world)
		p2Chr.conn.store.Write([]byte("\n\n> "))
		p1Chr.conn.store.Write([]byte("\n\n> "))
		return 1
	}
	return 0
}
