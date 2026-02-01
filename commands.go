package main

import (
	"database/sql"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

func Commands(cmdTokens []string, db *sql.DB, world *World, connection *ConnectionData) int {
	stream := connection.store
	switch len(cmdTokens) {
	case 1:
		switch cmdTokens[0] {
		case "exit":
			HandleClientDisconnect(connection, world, db)
			return 0
		case "selfharm":
			connection.session.character.hp -= 10
			fmt.Println(connection.session.character.hp)
			connection.store.Write([]byte("Ow! You poke yourself and lose 10 hp."))
		case "incmhp":
			connection.session.character.maxHp += 10
		case "look":
			HandleMovement(connection, world)
		case "north":
			if world.nodeList[connection.session.character.locationID].exits[0] != -1 {
				LeftNotifier(world, connection, "north")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[0]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "north")
			}
		case "south":
			if world.nodeList[connection.session.character.locationID].exits[1] != -1 {
				LeftNotifier(world, connection, "south")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[1]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "south")
			}
		case "west":
			if world.nodeList[connection.session.character.locationID].exits[2] != -1 {
				LeftNotifier(world, connection, "west")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[2]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "west")
			}
		case "east":
			if world.nodeList[connection.session.character.locationID].exits[3] != -1 {
				LeftNotifier(world, connection, "east")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[3]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "east")
			}

		case "inventory":
			iIDs := map[int]int{}
			for _, item := range world.items {
				if world.items[item.id].locationType == "player" && world.items[item.id].locationID == connection.session.id && !world.items[item.id].equipped {
					iIDs[world.items[item.id].templateID] += 1
				}
			}
			for tID, num := range iIDs {
				s := strconv.Itoa(num)
				stream.Write([]byte("  " + s + strings.Repeat(" ", 3-len(s)) + " | " + world.ItemTemplates[tID].name + "\n"))
			}
		case "equipped":
			s := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
			stream.Write([]byte("\n"))
			for i, slot := range s {
				if connection.session.character.equipment[slot] != 0 {
					stream.Write([]byte("  - " + s[i] + strings.Repeat(" ", 8-len(s[i])) + " - " + world.ItemTemplates[world.items[connection.session.character.equipment[slot]].templateID].name + "\n"))
				} else {
					stream.Write([]byte("  + " + s[i] + strings.Repeat(" ", 8-len(s[i])) + " - " + "<empty>" + "\n"))
				}
			}
		case "effects":
			stream.Write([]byte("Modifiers:\n"))
			for _, mod := range connection.session.character.modifiers {
				if mod.value > 0 {
					stream.Write([]byte("  " + world.ItemTemplates[world.items[mod.sourceID].templateID].name + ": " + mod.stat + " +" + strconv.Itoa(mod.value) + "\n"))
				} else {
					stream.Write([]byte("  " + world.ItemTemplates[world.items[mod.sourceID].templateID].name + ": " + mod.stat + " -" + strconv.Itoa(mod.value) + "\n"))
				}
			}
		case "stronk":
			connection.session.character.baseStats = Stats{10, 10, 10, 10, 10}
		case "profile":
			e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
			eList := [6]string{}
			for i, slot := range e {
				if connection.session.character.equipment[slot] != 0 {
					eList[i] = "  - " + e[i] + strings.Repeat(" ", 8-len(e[i])) + " - " + world.ItemTemplates[world.items[connection.session.character.equipment[slot]].templateID].name
				} else {
					eList[i] = "  + " + e[i] + strings.Repeat(" ", 8-len(e[i])) + " - " + "<empty>"
				}
			}

			var nameMedian int
			if len(connection.session.username)%2 == 1 {
				nameMedian = int(math.Floor(float64(len(connection.session.username))/2.0)) + 1
			} else {
				nameMedian = len(connection.session.username)/2 + 1
			}

			cardLength := 60
			s := connection.session.character.baseStats
			str, dex, agi, stam, int := strconv.Itoa(s.Str), strconv.Itoa(s.Dex), strconv.Itoa(s.Agi), strconv.Itoa(s.Stam), strconv.Itoa(s.Int)
			stream.Write([]byte("\n  +" + strings.Repeat("-", cardLength) + "+\n"))
			us := "  | " + strings.Repeat(" ", cardLength/2-nameMedian) + connection.session.username
			stream.Write([]byte(us + strings.Repeat(" ", cardLength-len(us)+2) + " |\n"))
			stream.Write([]byte("  +" + strings.Repeat("-", 27) + "+" + strings.Repeat("-", cardLength-28) + "+\n"))
			stream.Write([]byte("  |" + strings.Repeat(" ", 27) + "|" + strings.Repeat(" ", cardLength-28) + "|\n"))
			stream.Write([]byte("  | " + "stats:" + strings.Repeat(" ", 19) + " | "))
			stream.Write([]byte("equipment:" + strings.Repeat(" ", cardLength-40) + " |\n"))
			// stream.Write([]byte("  | " + "  Strength     :  [ " + strings.Repeat("0", 3-len(str)) + str + " ]" + strings.Repeat(" ", cardLength-27) + " |\n")) <-- for reference
			stream.Write([]byte("  | " + "  Strength     :  [ " + strings.Repeat("0", 3-len(str)) + str + " ]" + " | "))
			stream.Write([]byte(eList[0] + strings.Repeat(" ", cardLength-30-len(eList[0])) + " |\n"))
			stream.Write([]byte("  | " + "  Dexterity    :  [ " + strings.Repeat("0", 3-len(dex)) + dex + " ]" + " | "))
			stream.Write([]byte(eList[1] + strings.Repeat(" ", cardLength-30-len(eList[1])) + " |\n"))
			stream.Write([]byte("  | " + "  Agility      :  [ " + strings.Repeat("0", 3-len(agi)) + agi + " ]" + " | "))
			stream.Write([]byte(eList[2] + strings.Repeat(" ", cardLength-30-len(eList[2])) + " |\n"))
			stream.Write([]byte("  | " + "  Stamina      :  [ " + strings.Repeat("0", 3-len(stam)) + stam + " ]" + " | "))
			stream.Write([]byte(eList[3] + strings.Repeat(" ", cardLength-30-len(eList[3])) + " |\n"))
			stream.Write([]byte("  | " + "  Intelligence :  [ " + strings.Repeat("0", 3-len(int)) + int + " ]" + " | "))
			stream.Write([]byte(eList[4] + strings.Repeat(" ", cardLength-30-len(eList[4])) + " |\n"))
			stream.Write([]byte("  | " + strings.Repeat(" ", 25) + " | "))
			stream.Write([]byte(eList[5] + strings.Repeat(" ", cardLength-30-len(eList[5])) + " |\n"))
			stream.Write([]byte("  |" + strings.Repeat(" ", 27) + "|" + strings.Repeat(" ", cardLength-28) + "|\n"))
			stream.Write([]byte("  +" + strings.Repeat("-", 27) + "+" + strings.Repeat("-", cardLength-28) + "+\n"))

		default:
			stream.Write([]byte("\n  Command not found!\n"))
		}

	case 2:
		switch cmdTokens[0] {
		case "pickup":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				stream.Write([]byte("Invalid syntax. Please provide an integer!"))
			} else {
				if len(world.nodeList[connection.session.character.locationID].itemIDs) > i {
					stream.Write([]byte("You pick up a " + world.ItemTemplates[world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].templateID].name))
					world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationType = "player"
					world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationID = connection.session.id
					fmt.Println(world.nodeList[connection.session.character.locationID].itemIDs)
					_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "player", connection.session.id, world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].id)
					world.nodeList[connection.session.character.locationID].itemIDs = slices.Delete(world.nodeList[connection.session.character.locationID].itemIDs, i, i+1)
					fmt.Println(err)
				} else {
					stream.Write([]byte("There is no such item."))
				}
			}
		case "drop":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				stream.Write([]byte("Invalid syntax. Please provide an integer!"))
			} else {
				var playerItems []int
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id {
						playerItems = append(playerItems, item.id)
					}
				}
				if len(playerItems) > i {
					stream.Write([]byte("You dropped a " + world.ItemTemplates[world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].templateID].name))
					world.items[playerItems[i]].locationType = "room"
					world.items[playerItems[i]].locationID = world.nodeList[connection.session.character.locationID].id
					world.nodeList[connection.session.character.locationID].itemIDs = append(world.nodeList[connection.session.character.locationID].itemIDs, playerItems[i])
					_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "room", connection.session.character.locationID, playerItems[i])
					fmt.Println(err)
				} else {
					stream.Write([]byte("There is no such item."))
				}
			}
		case "examine":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				stream.Write([]byte("Invalid syntax. Please provide an integer!"))
			} else {
				var playerItems []int
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id {
						playerItems = append(playerItems, item.id)
					}
				}
				if len(playerItems) > i {
					stream.Write([]byte("Modifiers:\n"))
					for _, mod := range world.ItemTemplates[world.items[playerItems[i]].templateID].modifiers {
						if mod.value > 0 {
							stream.Write([]byte("  " + mod.stat + " +" + strconv.Itoa(mod.value) + "\n"))
						} else {
							stream.Write([]byte("  " + mod.stat + " -" + strconv.Itoa(mod.value) + "\n"))
						}
					}
				} else {
					stream.Write([]byte("There is no such item."))
				}
			}
		case "equip":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				stream.Write([]byte("Invalid syntax. Please provide an integer!"))
			} else {
				var playerItems []int
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped {
						playerItems = append(playerItems, item.id)
					}
				}
				if len(playerItems) > i && len(playerItems) != 0 {
					if connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype] != 0 {
						out := connection.session.character.modifiers[:0]
						for _, mod := range connection.session.character.modifiers {
							if mod.sourceID != connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype] {
								out = append(out, mod)
							}
						}
						connection.session.character.modifiers = out
						stream.Write([]byte("You unequip " + world.ItemTemplates[world.items[playerItems[i]].templateID].name + "\n"))
						world.items[connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype]].equipped = false
						connection.session.character.equipment[cmdTokens[1]] = 0
						_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", false, world.items[playerItems[i]].id)
						fmt.Println(err)
					}
					for _, mod := range world.ItemTemplates[world.items[playerItems[i]].templateID].modifiers {
						connection.session.character.modifiers = append(connection.session.character.modifiers, StatModifier{"item", playerItems[i], mod.stat, mod.value})
					}
					stream.Write([]byte("\nYou equip a " + world.ItemTemplates[world.items[playerItems[i]].templateID].name + " on the " + world.ItemTemplates[world.items[playerItems[i]].templateID].itype + "\n"))
					world.items[playerItems[i]].equipped = true
					_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", true, playerItems[i])
					fmt.Println(err)
					fmt.Println(connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype])
					connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype] = playerItems[i]
				} else {
					stream.Write([]byte("\nThere is no such item.\n"))
				}
			}
		case "unequip":
			c := false
			s := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
			for _, i := range s {
				if i == cmdTokens[1] {
					c = true
				}
			}
			if c {
				if connection.session.character.equipment[cmdTokens[1]] != 0 {
					out := connection.session.character.modifiers[:0]
					for _, mod := range connection.session.character.modifiers {
						if mod.sourceID != connection.session.character.equipment[cmdTokens[1]] {
							out = append(out, mod)
						}
					}
					connection.session.character.modifiers = out
					stream.Write([]byte("You unequip " + world.ItemTemplates[world.items[connection.session.character.equipment[cmdTokens[1]]].templateID].name + "\n"))
					world.items[connection.session.character.equipment[cmdTokens[1]]].equipped = false
					_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", false, world.items[connection.session.character.equipment[cmdTokens[1]]].id)
					connection.session.character.equipment[cmdTokens[1]] = 0
					fmt.Println(err)
				}
			} else {
				stream.Write([]byte("There is no such slot to unequip from."))
			}
		case "fight":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				world.mu.Lock()
				var p2Chr *Character = nil
				// FIXES THE BUG OF ONE WAY COMBAT
				// remove p1, p2 indexes, and sometimes combat only happens one way cuz of pointers n shi
				p2Index := -1
				p1Index := -1
				for i := range world.characters {
					char := world.characters[i]
					if char.conn == connection {
						p1Index = i
					}
					if char.conn != nil && char.conn.session.username == cmdTokens[1] && char.locationID == connection.session.character.locationID {
						p2Chr = char
						p2Index = i
					}
				}
				if p2Chr != nil && p1Index != -1 {
					p1Idx := new(int)
					*p1Idx = p1Index
					p2Idx := new(int)
					*p2Idx = p2Index

					stream.Write([]byte("Engaging " + p2Chr.conn.session.username + "\n"))
					p2Chr.conn.store.Write([]byte("Engaging " + connection.session.username + "\n"))
					p2Chr.inCombat = true
					p2Chr.targetID = p1Idx
					p2Chr.targetType = &TargetPlayer
					connection.session.character.inCombat = true
					connection.session.character.targetID = p2Idx
					connection.session.character.targetType = &TargetPlayer
				} else {
					stream.Write([]byte("Player not found in this room.\n"))
				}
				world.mu.Unlock()
			} else {
				if len(world.nodeList[connection.session.character.locationID].entityIDs) > i {
					stream.Write([]byte("Engaging a " + world.EntityTemplates[world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].templateID].name))
					world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].inCombat = true
					world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].targetID = &connection.session.id
					connection.session.character.inCombat = true
					connection.session.character.targetID = &world.nodeList[connection.session.character.locationID].entityIDs[i]
					connection.session.character.targetType = &TargetEntity
				}
			}
		case "list":
			for en, e := range world.entities {
				keywords := strings.FieldsFunc(strings.ToLower(world.EntityTemplates[e.templateID].name), func(r rune) bool {
					return r == ' ' || r == ','
				})
				matchBool := false
				for _, k := range keywords {
					if k == cmdTokens[1] {
						matchBool = true
					}
				}
				if matchBool && world.merchants[en] != nil {
					stream.Write([]byte("\n   [ ID ] + [QTY] + [ SELL ] + [  BUY  ] + [ NAME ]  \n"))
					for _, item := range world.merchants[en].list {
						id := strconv.Itoa(world.ItemTemplates[item].id)
						sp := strconv.Itoa(int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].sellRate))
						bp := strconv.Itoa(int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].buyRate))
						stream.Write([]byte("    " + id + strings.Repeat(" ", 6-len(id)) + "   " + "inf     " + sp + strings.Repeat(" ", 6-len(sp)) + "     " + bp + strings.Repeat(" ", 7-len(bp)) + "     " + world.ItemTemplates[item].name + "\n"))
					}
					stream.Write([]byte("\n + buy <id> " + cmdTokens[1]))
					stream.Write([]byte("\n - sell <id> " + cmdTokens[1] + "\n"))
					break
				}
			}

		default:
			stream.Write([]byte("Command not found!"))
		}

	case 3:
		switch cmdTokens[0] {
		case "buy":
			valid, i := validateInt(cmdTokens[1])
			if valid {
				for en, e := range world.entities {
					keywords := strings.FieldsFunc(strings.ToLower(world.EntityTemplates[e.templateID].name), func(r rune) bool {
						return r == ' ' || r == ','
					})
					item := i
					matchBool := false
					for _, k := range keywords {
						if k == cmdTokens[2] {
							matchBool = true
						}
					}
					if matchBool && world.merchants[en] != nil {
						bp := int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].buyRate)
						bpS := strconv.Itoa(bp)
						if connection.session.character.coins >= int(bp) {
							CreateAndInsertItem(connection, world, db, item)
							stream.Write([]byte("\nYou buy 1x " + world.ItemTemplates[item].name + " for " + bpS + " coins from " + world.EntityTemplates[e.templateID].name + "\n"))
							connection.session.character.coins -= int(bp)
						} else {
							stream.Write([]byte("\nYou don't have enough coins to buy this item!\n"))
						}
						break
					}
				}
			}
		case "sell":
			valid, i := validateInt(cmdTokens[1])
			if valid {
				for en, e := range world.entities {
					keywords := strings.FieldsFunc(strings.ToLower(world.EntityTemplates[e.templateID].name), func(r rune) bool {
						return r == ' ' || r == ','
					})
					item := i
					matchBool := false
					for _, k := range keywords {
						if k == cmdTokens[2] {
							matchBool = true
						}
					}
					if matchBool && world.merchants[en] != nil {
						sp := int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].sellRate)
						spS := strconv.Itoa(sp)
						for _, i := range world.items {
							if i.locationType == "player" && i.locationID == connection.session.id && i.templateID == item && !i.equipped {
								DeleteItem(connection, world, db, i.id)
								stream.Write([]byte("\nYou sell 1x " + world.ItemTemplates[item].name + " for " + spS + " coins from " + world.EntityTemplates[e.templateID].name + "\n"))
								connection.session.character.coins += int(sp)
								break
							}
						}
						break
					}
				}
			}
		}

	default:
		stream.Write([]byte("Too many arguments!"))
	}
	return 1
}
