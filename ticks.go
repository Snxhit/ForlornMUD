package main

import (
	"database/sql"
	"fmt"
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
				fmt.Println(conn.session.character)
				if conn.session.character.inCombat {
					fmt.Println(conn.session.username)
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
					case &TargetPlayer:
						r := combatPlayer(world, conn)
						if r == 1 {
							continue
						}
					}
				}
			}
			world.mu.Unlock()
		case <-saveTicker.C:
			// nuh
		}
	}
}
