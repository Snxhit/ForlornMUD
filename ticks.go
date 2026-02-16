package main

import (
	"database/sql"
	"strconv"
	"time"
)

func ticks(world *World, db *sql.DB) {
	worldTicker := time.NewTicker(3 * time.Second)
	saveTicker := time.NewTicker(30 * time.Second)
	defer worldTicker.Stop()
	defer saveTicker.Stop()
	for {
		select {
		case <-worldTicker.C:
			world.mu.Lock()
			world.tick++
			for _, conn := range world.connections {
				if conn.session == nil || !conn.session.authorized || conn.session.character == nil {
					continue
				}
				if conn.session.character.hp < conn.session.character.maxHp && !conn.session.character.inCombat {
					conn.session.character.hp += 5
				}
				if conn.session.character.hp > conn.session.character.maxHp {
					conn.session.character.hp = conn.session.character.maxHp
				} else if conn.session.character.hp < 0 {
					conn.session.character.hp = 0
				}
				if conn.session.character.inCombat {
					if conn.session.character.targetID == nil {
						conn.session.character.inCombat = false
						continue
					}
					switch conn.session.character.targetType {
					case &TargetEntity:
						r := combatEntity(world, conn, db)
						if r == 1 {
							continue
						}
						if conn.isClientWeb {
							conn.store.Write([]byte("\n\x01COMBAT " + "type:entity" + " hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"" + world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].name + "\" enemyHp:" + strconv.Itoa(world.entities[*conn.session.character.targetID].hp) + " enemyMaxHp:" + strconv.Itoa(world.EntityTemplates[world.entities[*conn.session.character.targetID].templateID].maxHp) + "\n"))
						}
					case &TargetPlayer:
						r := combatPlayer(world, conn)
						if r == 1 {
							continue
						}
						if conn.isClientWeb {
							conn.store.Write([]byte("\n\x01COMBAT " + "type:player" + " hp:" + strconv.Itoa(conn.session.character.hp) + " maxHp:" + strconv.Itoa(conn.session.character.maxHp) + " enemyName:\"" + world.characters[*conn.session.character.targetID].conn.session.username + "\" enemyHp:" + strconv.Itoa(world.characters[*conn.session.character.targetID].hp) + " enemyMaxHp:" + strconv.Itoa(world.characters[*conn.session.character.targetID].maxHp) + "\n"))
						}
					}
				}
			}
			for i := range world.spawners {
				s := &world.spawners[i]
				if s.templateType == "entity" {
					var currentAlive int = 0
					for _, i := range world.nodeList[s.locationID].entityIDs {
						if world.EntityTemplates[world.entities[i].templateID].id == s.templateID {
							currentAlive += 1
						}
					}
					if currentAlive >= s.maxSpawns {
						s.nextSpawnTick += 1
						continue
					}
					if world.tick >= int64(s.nextSpawnTick) {
						SpawnAndInsertEntity(world, db, s.locationID, s.templateID)
						s.nextSpawnTick = int(world.tick) + s.duration
					}
				} else if s.templateType == "item" {
					var currentSpawned int = 0
					for _, i := range world.nodeList[s.locationID].itemIDs {
						if world.ItemTemplates[world.items[i].templateID].id == s.templateID {
							currentSpawned += 1
						}
					}
					if currentSpawned >= s.maxSpawns {
						continue
					}
					if world.tick >= int64(s.nextSpawnTick) {
						CreateAndPlaceItem(world, db, s.templateID, s.locationID)
						s.nextSpawnTick = int(world.tick) + s.duration
					}
				}
			}
			world.mu.Unlock()
		case <-saveTicker.C:
			// nuh
		}
	}
}
