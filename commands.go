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
		case "exit", "quit":
			HandleClientDisconnect(connection, world, db)
			return 0
		case "look", "l":
			HandleMovement(connection, world)
		case "north", "n":
			if world.nodeList[connection.session.character.locationID].exits[0] != -1 {
				LeftNotifier(world, connection, "north")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[0]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "north")
			} else {
				stream.Write([]byte("\n  You cannot move in that direction!\n"))
			}
		case "south", "s":
			if world.nodeList[connection.session.character.locationID].exits[1] != -1 {
				LeftNotifier(world, connection, "south")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[1]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "south")
			} else {
				stream.Write([]byte("\n  You cannot move in that direction!\n"))
			}
		case "west", "w":
			if world.nodeList[connection.session.character.locationID].exits[2] != -1 {
				LeftNotifier(world, connection, "west")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[2]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "west")
			} else {
				stream.Write([]byte("\n  You cannot move in that direction!\n"))
			}
		case "east", "e":
			if world.nodeList[connection.session.character.locationID].exits[3] != -1 {
				LeftNotifier(world, connection, "east")
				connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[3]
				HandleMovement(connection, world)
				EnterNotifier(world, connection, "east")
			} else {
				stream.Write([]byte("\n  You cannot move in that direction!\n"))
			}

		case "inventory", "inv", "i":
			stream.Write([]byte("\n"))
			iIDs := map[int]int{}
			for _, item := range world.items {
				if world.items[item.id].locationType == "player" && world.items[item.id].locationID == connection.session.id && !world.items[item.id].equipped {
					iIDs[world.items[item.id].templateID] += 1
				}
			}
			for tID, num := range iIDs {
				s := strconv.Itoa(num)
				stream.Write([]byte("  " + color(connection, "cyan", "tp") + s + color(connection, "reset", "reset") + strings.Repeat(" ", 3-len(s)) + " | " + world.ItemTemplates[tID].name + "\n"))
			}
		case "equipped", "worn", "armor":
			s := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
			stream.Write([]byte("\n"))
			for i, slot := range s {
				if connection.session.character.equipment[slot] != 0 {
					stream.Write([]byte(color(connection, "cyan", "tp") + "  - " + color(connection, "reset", "reset") + s[i] + strings.Repeat(" ", 8-len(s[i])) + " - " + world.ItemTemplates[world.items[connection.session.character.equipment[slot]].templateID].name + "\n"))
				} else {
					stream.Write([]byte(color(connection, "cyan", "tp") + "  + " + color(connection, "reset", "reset") + s[i] + strings.Repeat(" ", 8-len(s[i])) + " - " + "<empty>" + "\n"))
				}
			}
		case "effects", "effs":
			stream.Write([]byte(color(connection, "cyan", "tp") + "\n  Modifiers:\n" + color(connection, "reset", "reset")))
			for _, mod := range connection.session.character.modifiers {
				if mod.value > 0 {
					stream.Write([]byte("    " + color(connection, "green", "tp") + glphys(connection, "sau") + color(connection, "reset", "reset") + " " + world.ItemTemplates[world.items[mod.sourceID].templateID].name + ": " + color(connection, "yellow", "tp") + mod.stat + color(connection, "reset", "reset") + " +" + strconv.Itoa(mod.value) + "\n"))
				} else {
					stream.Write([]byte("    " + color(connection, "red", "tp") + glphys(connection, "sad") + color(connection, "reset", "reset") + world.ItemTemplates[world.items[mod.sourceID].templateID].name + ": " + color(connection, "yellow", "tp") + mod.stat + color(connection, "reset", "reset") + " -" + strconv.Itoa(mod.value) + "\n"))
				}
			}
		case "profile", "pf":
			e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
			eList := [6]string{}
			for i, slot := range e {
				if !connection.isColorEnabled {
					if connection.session.character.equipment[slot] != 0 {
						eList[i] = "  - " + e[i] + strings.Repeat(" ", 8-len(e[i])) + " - " + world.ItemTemplates[world.items[connection.session.character.equipment[slot]].templateID].name
					} else {
						eList[i] = "  + " + e[i] + strings.Repeat(" ", 8-len(e[i])) + " - " + "<empty>"
					}
				} else {
					if connection.session.character.equipment[slot] != 0 {
						eList[i] = "  - " + color(connection, "magenta", "tp") + e[i] + color(connection, "reset", "reset") + strings.Repeat(" ", 8-len(e[i])) + " - "
						eP := color(connection, "cyan", "tp") + world.ItemTemplates[world.items[connection.session.character.equipment[slot]].templateID].name + color(connection, "reset", "reset")
						// framework laptop
						if visibleLen(eP) > 15 {
							eList[i] += eP[0:17]
							eList[i] += "..."
							eList[i] += color(connection, "reset", "reset")
						} else {
							eList[i] += eP
						}
					} else {
						eList[i] = "  + " + color(connection, "magenta", "tp") + e[i] + color(connection, "reset", "reset") + strings.Repeat(" ", 8-len(e[i])) + " - " + "<empty>"
					}
				}
			}

			var nameMedian int
			if len(connection.session.username)%2 == 1 {
				nameMedian = int(math.Floor(float64(len(connection.session.username))/2.0)) + 1
			} else {
				nameMedian = len(connection.session.username)/2 + 1
			}

			cardLength := 60
			c := strconv.Itoa(connection.session.character.coins)
			t := strconv.Itoa(connection.session.character.trains)
			lvl := strconv.Itoa(connection.session.character.level)
			exp := strconv.Itoa(connection.session.character.exp)
			expBars := (connection.session.character.exp % 100) / 5
			hp := strconv.Itoa(connection.session.character.hp)
			maxHp := strconv.Itoa(connection.session.character.maxHp)
			hpBars := int(math.Floor(float64(connection.session.character.hp)/float64(connection.session.character.maxHp)*100)) / 4
			s := connection.session.character.baseStats
			str, dex, agi, stam, int := strconv.Itoa(s.Str), strconv.Itoa(s.Dex), strconv.Itoa(s.Agi), strconv.Itoa(s.Stam), strconv.Itoa(s.Int)
			printProfileCard(connection, nameMedian, c, t, lvl, exp, expBars, str, dex, agi, stam, int, cardLength, eList, hp, maxHp, hpBars)

		// all the case 2 or 3s
		case "help", "h":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <query>\n"))
		case "pickup", "take", "pick", "collect":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "drop", "put", "throw":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "examine":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "equip", "wear", "eq":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "unequip", "remove", "uneq":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <slot>\n"))
		case "use", "u":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "fight", "attack", "kick", "f":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <entity>\n"))
		case "train":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <stat_name>\n"))
		case "list", "browse":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <entity>\n"))
		case "toggle", "switch":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <option>\n"))
		case "move", "go", "m":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <n|s|w|e>\n"))
		case "buy":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <item_id> <entity>\n"))
		case "sell":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <index_item> <entity>\n"))

		default:
			stream.Write([]byte("\n  Command not found!\n"))
		}

	case 2:
		switch cmdTokens[0] {
		case "help", "h":
			printHelpFile(cmdTokens[1], connection)
		case "pickup", "take", "pick", "collect":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				itemFound := false
				for _, id := range world.nodeList[connection.session.character.locationID].itemIDs {
					keywords := strings.FieldsFunc(strings.ToLower(world.ItemTemplates[world.items[id].templateID].name), func(r rune) bool {
						return r == ' ' || r == ','
					})
					matchBool := false
					for _, k := range keywords {
						if k == cmdTokens[1] {
							matchBool = true
						}
					}
					if matchBool {
						stream.Write([]byte("\n  You pick up a " + color(connection, "cyan", "tp") + world.ItemTemplates[world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].templateID].name + color(connection, "reset", "reset") + "\n"))
						world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationType = "player"
						world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationID = connection.session.id
						_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "player", connection.session.id, world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].id)
						world.nodeList[connection.session.character.locationID].itemIDs = slices.Delete(world.nodeList[connection.session.character.locationID].itemIDs, i, i+1)
						fmt.Println(err)
						itemFound = true
						break
					}
				}
				if !itemFound {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			} else {
				if len(world.nodeList[connection.session.character.locationID].itemIDs) > i {
					stream.Write([]byte("\n  You pick up a " + color(connection, "cyan", "tp") + world.ItemTemplates[world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].templateID].name + color(connection, "reset", "reset") + "\n"))
					world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationType = "player"
					world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationID = connection.session.id
					_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "player", connection.session.id, world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].id)
					world.nodeList[connection.session.character.locationID].itemIDs = slices.Delete(world.nodeList[connection.session.character.locationID].itemIDs, i, i+1)
					fmt.Println(err)
				} else {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			}
		case "drop", "put", "throw":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				itemFound := false
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id {
						id := item.id
						keywords := strings.FieldsFunc(strings.ToLower(world.ItemTemplates[world.items[id].templateID].name), func(r rune) bool {
							return r == ' ' || r == ','
						})
						matchBool := false
						for _, k := range keywords {
							if k == cmdTokens[1] {
								matchBool = true
							}
						}
						if matchBool {
							stream.Write([]byte("\n  You dropped a " + color(connection, "cyan", "tp") + world.ItemTemplates[world.items[id].templateID].name + color(connection, "reset", "reset") + "\n"))
							world.items[id].locationType = "room"
							world.items[id].locationID = world.nodeList[connection.session.character.locationID].id
							world.nodeList[connection.session.character.locationID].itemIDs = append(world.nodeList[connection.session.character.locationID].itemIDs, id)
							_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "room", connection.session.character.locationID, id)
							fmt.Println(err)
							itemFound = true
							break
						}
					}
				}
				if !itemFound {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			} else {
				var playerItems []int
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id {
						playerItems = append(playerItems, item.id)
					}
				}
				if len(playerItems) > i {
					stream.Write([]byte("\n  You dropped a " + color(connection, "cyan", "tp") + world.ItemTemplates[world.items[playerItems[i]].templateID].name + color(connection, "reset", "reset") + "\n"))
					world.items[playerItems[i]].locationType = "room"
					world.items[playerItems[i]].locationID = world.nodeList[connection.session.character.locationID].id
					world.nodeList[connection.session.character.locationID].itemIDs = append(world.nodeList[connection.session.character.locationID].itemIDs, playerItems[i])
					_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "room", connection.session.character.locationID, playerItems[i])
					fmt.Println(err)
				} else {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			}
		case "examine":
			var itemT *ItemTemplate
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped {
						keywords := strings.FieldsFunc(strings.ToLower(world.ItemTemplates[item.templateID].name), func(r rune) bool {
							return r == ' ' || r == ','
						})
						matchBool := false
						for _, k := range keywords {
							if k == cmdTokens[1] {
								matchBool = true
							}
						}
						if matchBool {
							itemT = world.ItemTemplates[item.templateID]
						}
					}
				}
				if itemT == nil {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			} else {
				var playerItems []int
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id {
						playerItems = append(playerItems, item.id)
					}
				}
				if len(playerItems) > i {
					itemT = world.ItemTemplates[world.items[playerItems[i]].templateID]
				} else {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			}
			if itemT != nil {
				stream.Write([]byte("\n  " + color(connection, "cyan", "tp") + itemT.name + color(connection, "reset", "reset")))
				stream.Write([]byte("\n    " + itemT.description + "\n"))
				e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
				if slices.Contains(e, itemT.itype) {
					stream.Write([]byte("\n  Type:" + color(connection, "green", "tp") + " Equipment\n" + color(connection, "reset", "reset")))
					stream.Write([]byte("  Slot: " + color(connection, "magenta", "tp") + itemT.itype + color(connection, "reset", "reset")))
					stream.Write([]byte("\n\n  Base Damage: " + color(connection, "red", "tp") + strconv.Itoa(itemT.baseDam) + color(connection, "reset", "reset") + "\n"))
					stream.Write([]byte("  Base Defense: " + color(connection, "red", "tp") + strconv.Itoa(itemT.baseDef) + color(connection, "reset", "reset")))
				} else {
					stream.Write([]byte("\n  Type:" + color(connection, "green", "tp") + " Consumable" + color(connection, "reset", "reset")))
					for _, eff := range itemT.effects {
						if eff.effect == "hp" {
							stream.Write([]byte("\n  HP: +" + color(connection, "red", "tp") + strconv.Itoa(eff.value) + color(connection, "reset", "reset")))
						}
					}
				}
				stream.Write([]byte("\n\n  Modifiers:\n"))
				for _, mod := range itemT.modifiers {
					if mod.value > 0 {
						stream.Write([]byte("    " + mod.stat + color(connection, "green", "tp") + " +" + color(connection, "yellow", "tp") + strconv.Itoa(mod.value) + color(connection, "reset", "reset") + "\n"))
					} else {
						stream.Write([]byte("    " + mod.stat + color(connection, "red", "tp") + " -" + color(connection, "yellow", "tp") + strconv.Itoa(mod.value) + color(connection, "reset", "reset") + "\n"))
					}
				}
			}
		case "equip", "wear", "eq":
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				itemFound := false
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped {
						keywords := strings.FieldsFunc(strings.ToLower(world.ItemTemplates[item.templateID].name), func(r rune) bool {
							return r == ' ' || r == ','
						})
						matchBool := false
						for _, k := range keywords {
							if k == cmdTokens[1] {
								matchBool = true
							}
						}
						e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
						if matchBool && slices.Contains(e, world.ItemTemplates[item.templateID].itype) {
							if connection.session.character.equipment[world.ItemTemplates[item.templateID].itype] != 0 {
								out := connection.session.character.modifiers[:0]
								for _, mod := range connection.session.character.modifiers {
									if mod.sourceID != connection.session.character.equipment[world.ItemTemplates[item.templateID].itype] {
										out = append(out, mod)
									}
								}
								connection.session.character.modifiers = out
								stream.Write([]byte("\n  You unequip " + color(connection, "cyan", "tp") + world.ItemTemplates[item.templateID].name + color(connection, "reset", "reset") + "\n"))
								world.items[connection.session.character.equipment[world.ItemTemplates[item.templateID].itype]].equipped = false
								connection.session.character.equipment[cmdTokens[1]] = 0
								_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", false, item.id)
								fmt.Println(err)
							}
							for _, mod := range world.ItemTemplates[item.templateID].modifiers {
								connection.session.character.modifiers = append(connection.session.character.modifiers, StatModifier{"item", item.id, mod.stat, mod.value})
							}
							stream.Write([]byte("\n  You equip a " + color(connection, "cyan", "tp") + world.ItemTemplates[item.templateID].name + color(connection, "reset", "reset") + " on the " + color(connection, "yellow", "tp") + world.ItemTemplates[item.templateID].itype + color(connection, "reset", "reset") + "\n"))
							item.equipped = true
							_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", true, item.id)
							fmt.Println(err)
							connection.session.character.equipment[world.ItemTemplates[item.templateID].itype] = item.id
							itemFound = true
							break
						}
					}
				}
				if !itemFound {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			} else {
				var playerItems []int
				e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped && slices.Contains(e, world.ItemTemplates[item.templateID].itype) {
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
						stream.Write([]byte("\n  You unequip " + color(connection, "cyan", "tp") + world.ItemTemplates[world.items[playerItems[i]].templateID].name + color(connection, "reset", "reset") + "\n"))
						world.items[connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype]].equipped = false
						connection.session.character.equipment[cmdTokens[1]] = 0
						_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", false, world.items[playerItems[i]].id)
						fmt.Println(err)
					}
					for _, mod := range world.ItemTemplates[world.items[playerItems[i]].templateID].modifiers {
						connection.session.character.modifiers = append(connection.session.character.modifiers, StatModifier{"item", playerItems[i], mod.stat, mod.value})
					}
					stream.Write([]byte("\n  You equip a " + color(connection, "cyan", "tp") + world.ItemTemplates[world.items[playerItems[i]].templateID].name + color(connection, "reset", "reset") + " on the " + color(connection, "yellow", "tp") + world.ItemTemplates[world.items[playerItems[i]].templateID].itype + color(connection, "reset", "reset") + "\n"))
					world.items[playerItems[i]].equipped = true
					_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", true, playerItems[i])
					fmt.Println(err)
					fmt.Println(connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype])
					connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype] = playerItems[i]
				} else {
					stream.Write([]byte("\n  Item not found!\n"))
				}
			}
		case "unequip", "remove", "uneq":
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
					stream.Write([]byte("\n  You unequip " + color(connection, "cyan", "tp") + world.ItemTemplates[world.items[connection.session.character.equipment[cmdTokens[1]]].templateID].name + color(connection, "reset", "reset") + "\n"))
					world.items[connection.session.character.equipment[cmdTokens[1]]].equipped = false
					_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", false, world.items[connection.session.character.equipment[cmdTokens[1]]].id)
					connection.session.character.equipment[cmdTokens[1]] = 0
					fmt.Println(err)
				}
			} else {
				stream.Write([]byte("\n  There is no such slot to unequip from.\n"))
			}
		case "use", "u":
			var itemT *ItemTemplate
			var itemM *Item
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped {
						keywords := strings.FieldsFunc(strings.ToLower(world.ItemTemplates[item.templateID].name), func(r rune) bool {
							return r == ' ' || r == ','
						})
						matchBool := false
						for _, k := range keywords {
							if k == cmdTokens[1] {
								matchBool = true
							}
						}
						e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
						if matchBool && !slices.Contains(e, world.ItemTemplates[item.templateID].itype) {
							itemT = world.ItemTemplates[item.templateID]
							itemM = item
							break
						}
					}
				}
			} else {
				var playerItems []int
				e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped && !slices.Contains(e, world.ItemTemplates[item.templateID].itype) {
						playerItems = append(playerItems, item.id)
					}
				}
				if len(playerItems) > i && len(playerItems) != 0 {
					itemT = world.ItemTemplates[world.items[playerItems[i]].templateID]
					itemM = world.items[playerItems[i]]
				}
			}
			if itemT != nil {
				for _, effect := range itemT.effects {
					switch effect.effect {
					case "hp":
						if connection.session.character.hp != connection.session.character.maxHp {
							stream.Write([]byte("\n  You use a " + color(connection, "cyan", "tp") + itemT.name + color(connection, "reset", "reset") + " to heal " + color(connection, "red", "tp") + strconv.Itoa(effect.value) + color(connection, "reset", "reset") + " hp! (" + strconv.Itoa(connection.session.character.hp) + "/" + strconv.Itoa(connection.session.character.maxHp) + ")\n"))
							if connection.isClientWeb {
								connection.store.Write([]byte("\x01SELF " + "hp:" + strconv.Itoa(connection.session.character.hp) + " coins:" + strconv.Itoa(connection.session.character.coins) + "\n"))
							}
							connection.session.character.hp += effect.value
							_, err := db.Exec("DELETE FROM items WHERE id = ?", itemM.id)
							delete(world.items, itemM.id)
							fmt.Println(err)
						} else {
							stream.Write([]byte("\n  Your health is already full!\n"))
						}
					}
				}
			} else {
				stream.Write([]byte("\n  Item not found!\n"))
			}
		case "fight", "attack", "kick", "f":
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
					if char.conn != nil && char.conn.session.username == cmdTokens[1] && char.locationID == connection.session.character.locationID && !char.inCombat {
						p2Chr = char
						p2Index = i
					}
				}
				if p2Chr != nil && p1Index != -1 {
					p1Idx := new(int)
					*p1Idx = p1Index
					p2Idx := new(int)
					*p2Idx = p2Index

					stream.Write([]byte("\n  Engaging " + color(connection, "cyan", "tp") + p2Chr.conn.session.username + color(connection, "reset", "reset") + "\n"))
					p2Chr.conn.store.Write([]byte("\x1b[2K\r  " + color(p2Chr.conn, "cyan", "tp") + connection.session.username + color(p2Chr.conn, "reset", "reset") + " wants to fight!\n\n> "))
					p2Chr.inCombat = true
					p2Chr.targetID = p1Idx
					p2Chr.targetType = &TargetPlayer
					connection.session.character.inCombat = true
					connection.session.character.targetID = p2Idx
					connection.session.character.targetType = &TargetPlayer
				} else {
					entityFound := false
					for _, id := range world.nodeList[connection.session.character.locationID].entityIDs {
						keywords := strings.FieldsFunc(strings.ToLower(world.EntityTemplates[world.entities[id].templateID].name), func(r rune) bool {
							return r == ' ' || r == ','
						})
						matchBool := false
						for _, k := range keywords {
							if k == cmdTokens[1] {
								matchBool = true
							}
						}
						if matchBool && !world.entities[id].inCombat {
							stream.Write([]byte("\n  Engaging a " + color(connection, "cyan", "tp") + world.EntityTemplates[world.entities[id].templateID].name + color(connection, "reset", "reset") + "\n"))
							world.entities[id].inCombat = true
							world.entities[id].targetID = &connection.session.id
							connection.session.character.inCombat = true
							connection.session.character.targetID = &id
							connection.session.character.targetType = &TargetEntity
							entityFound = true
							break
						}
					}
					if !entityFound && p2Chr == nil {
						stream.Write([]byte("\n  Entity not found!\n"))
					}
				}
				world.mu.Unlock()
			} else {
				if len(world.nodeList[connection.session.character.locationID].entityIDs) > i && !world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].inCombat {
					stream.Write([]byte("\n  Engaging a " + color(connection, "cyan", "tp") + world.EntityTemplates[world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].templateID].name + color(connection, "reset", "reset") + "\n"))
					world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].inCombat = true
					world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].targetID = &connection.session.id
					connection.session.character.inCombat = true
					connection.session.character.targetID = &world.nodeList[connection.session.character.locationID].entityIDs[i]
					connection.session.character.targetType = &TargetEntity
				} else {
					stream.Write([]byte("\n  Entity not found!\n"))
				}
			}
		case "train":
			opts := []string{"str", "dex", "agi", "stam", "int", "hp"}
			if connection.session.character.trains < 1 {
				stream.Write([]byte("\n  You do not have enough trains!\n"))
			}
			if slices.Contains(opts, cmdTokens[1]) {
				switch cmdTokens[1] {
				case "str":
					connection.session.character.baseStats.Str += 1
					stream.Write([]byte("\n  You train once and increase your " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset") + " stat by one!"))
				case "dex":
					connection.session.character.baseStats.Dex += 1
					stream.Write([]byte("\n  You train once and increase your " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset") + " stat by one!"))
				case "agi":
					connection.session.character.baseStats.Agi += 1
					stream.Write([]byte("\n  You train once and increase your " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset") + " stat by one!"))
				case "stam":
					connection.session.character.baseStats.Stam += 1
					stream.Write([]byte("\n  You train once and increase your " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset") + " stat by one!"))
				case "int":
					connection.session.character.baseStats.Int += 1
					stream.Write([]byte("\n  You train once and increase your " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset") + " stat by one!"))
				case "hp":
					connection.session.character.maxHp += 10
					stream.Write([]byte("\n  You train once and increase your maximum " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset") + " by ten!"))
				}
				connection.session.character.trains -= 1
				if connection.isClientWeb {
					connection.store.Write([]byte("\n\x01EXP " + "exp:" + strconv.Itoa(connection.session.character.exp) + " lvl:" + strconv.Itoa(connection.session.character.level) + " trains:" + strconv.Itoa(connection.session.character.trains) + "\n"))
				}
				stream.Write([]byte("\n  You now have one less " + color(connection, "cyan", "tp") + "train" + color(connection, "reset", "reset") + ".\n"))
			}

		case "list", "browse":
			entityFound := false
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
					headingUncolored := "\n   [ ID ] + [QTY] + [ SELL ] + [  BUY  ] + [ NAME ]  \n"
					headingColored := "\n   [\x1b[30;47m ID \x1b[39;49m] + [\x1b[30;47mQTY\x1b[39;49m] + [\x1b[30;47m SELL \x1b[39;49m] + [\x1b[30;47m  BUY  \x1b[39;49m] + [\x1b[30;47m NAME \x1b[39;49m]  \n"
					if !connection.isColorEnabled {
						stream.Write([]byte(headingUncolored))
					} else {
						stream.Write([]byte(headingColored))
					}
					for _, item := range world.merchants[en].list {
						id := strconv.Itoa(world.ItemTemplates[item].id)
						sp := strconv.Itoa(int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].sellRate))
						bp := strconv.Itoa(int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].buyRate))
						stream.Write([]byte("    " + color(connection, "green", "tp") + id + strings.Repeat(" ", 6-len(id)) + "   " + color(connection, "magenta", "tp") + "inf     " + color(connection, "yellow", "tp") + sp + strings.Repeat(" ", 6-len(sp)) + "     " + bp + strings.Repeat(" ", 7-len(bp)) + "     " + color(connection, "cyan", "tp") + world.ItemTemplates[item].name + color(connection, "reset", "reset") + "\n"))
					}
					stream.Write([]byte(color(connection, "green", "tp") + "\n +" + " buy " + color(connection, "reset", "reset") + "<id> " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset")))
					stream.Write([]byte(color(connection, "red", "tp") + "\n -" + " sell " + color(connection, "reset", "reset") + "<id> " + color(connection, "cyan", "tp") + cmdTokens[1] + color(connection, "reset", "reset") + "\n"))
					entityFound = true
					break
				}
			}
			if !entityFound {
				stream.Write([]byte("\n  Entity not found!\n"))
			}

		case "toggle", "switch":
			switch cmdTokens[1] {
			case "pretty":
				connection.isPrettyEnabled = !connection.isPrettyEnabled
				stream.Write([]byte("\n  Toggled pretty, set to " + strconv.FormatBool(connection.isPrettyEnabled) + "\n"))
			case "color":
				connection.isColorEnabled = !connection.isColorEnabled
				stream.Write([]byte("\n  Toggled color, set to " + strconv.FormatBool(connection.isColorEnabled) + "\n"))
			}

		case "move", "go", "m":
			switch cmdTokens[1] {
			case "north", "n":
				if world.nodeList[connection.session.character.locationID].exits[0] != -1 {
					LeftNotifier(world, connection, "north")
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[0]
					HandleMovement(connection, world)
					EnterNotifier(world, connection, "north")
				} else {
					stream.Write([]byte("\n  You cannot move in that direction!\n"))
				}
			case "south", "s":
				if world.nodeList[connection.session.character.locationID].exits[1] != -1 {
					LeftNotifier(world, connection, "south")
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[1]
					HandleMovement(connection, world)
					EnterNotifier(world, connection, "south")
				} else {
					stream.Write([]byte("\n  You cannot move in that direction!\n"))
				}
			case "west", "w":
				if world.nodeList[connection.session.character.locationID].exits[2] != -1 {
					LeftNotifier(world, connection, "west")
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[2]
					HandleMovement(connection, world)
					EnterNotifier(world, connection, "west")
				} else {
					stream.Write([]byte("\n  You cannot move in that direction!\n"))
				}
			case "east", "e":
				if world.nodeList[connection.session.character.locationID].exits[3] != -1 {
					LeftNotifier(world, connection, "east")
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[3]
					HandleMovement(connection, world)
					EnterNotifier(world, connection, "east")
				} else {
					stream.Write([]byte("\n  You cannot move in that direction!\n"))
				}
			}

		// all the case 1s or 3s
		case "exit", "quit":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "look", "l":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "north", "n":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "south", "s":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "west", "w":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "east", "e":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "inventory", "inv", "i":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "equipped", "worn", "armor":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "effects", "effs":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "profile", "pf":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "buy":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <item_id> <entity>\n"))
		case "sell":
			stream.Write([]byte("\n  Not enough arguments! Try " + cmdTokens[0] + " <index_item> <entity>\n"))

		default:
			stream.Write([]byte("\n  Command not found!\n"))
		}

	case 3:
		switch cmdTokens[0] {
		case "buy":
			valid, i := validateInt(cmdTokens[1])
			if valid {
				merchantFound := false
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
					if matchBool && world.merchants[en] != nil && world.entities[en].locationID == connection.session.character.locationID {
						merchantFound = true
						itemInList := false
						for _, merchantItem := range world.merchants[en].list {
							if merchantItem == item {
								itemInList = true
								break
							}
						}
						if !itemInList {
							stream.Write([]byte("\n  Item not found!\n"))
							break
						}
						bp := int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].buyRate)
						bpS := strconv.Itoa(bp)
						if connection.session.character.coins >= int(bp) {
							CreateAndInsertItem(connection, world, db, item)
							stream.Write([]byte("\n  You buy 1x " + color(connection, "cyan", "tp") + world.ItemTemplates[item].name + color(connection, "reset", "reset") + " for " + color(connection, "yellow", "tp") + bpS + color(connection, "reset", "reset") + " coins from " + color(connection, "cyan", "tp") + world.EntityTemplates[e.templateID].name + color(connection, "reset", "reset") + "\n"))
							connection.session.character.coins -= int(bp)
							connection.store.Write([]byte("\n\x01SELF coins:" + strconv.Itoa(connection.session.character.coins) + "\n"))
						} else {
							stream.Write([]byte(color(connection, "magenta", "tp") + "\n  You don't have enough coins to buy this item!\n" + color(connection, "reset", "reset")))
						}
						break
					}
				}
				if !merchantFound {
					stream.Write([]byte("\n  Entity not found!\n"))
				}
			}
		case "sell":
			valid, i := validateInt(cmdTokens[1])
			if valid {
				merchantFound := false
				itemSold := false
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
					if matchBool && world.merchants[en] != nil && world.entities[en].locationID == connection.session.character.locationID {
						merchantFound = true
						itemInList := false
						for _, merchantItem := range world.merchants[en].list {
							if merchantItem == item {
								itemInList = true
								break
							}
						}
						if !itemInList {
							stream.Write([]byte("\n  Item not found!\n"))
							break
						}
						sp := int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].sellRate)
						spS := strconv.Itoa(sp)
						for _, i := range world.items {
							if i.locationType == "player" && i.locationID == connection.session.id && i.templateID == item && !i.equipped {
								DeleteItem(connection, world, db, i.id)
								stream.Write([]byte("\n  You sell 1x " + color(connection, "cyan", "tp") + world.ItemTemplates[item].name + color(connection, "reset", "reset") + " for " + color(connection, "yellow", "tp") + spS + color(connection, "reset", "reset") + " coins to " + color(connection, "cyan", "tp") + world.EntityTemplates[e.templateID].name + color(connection, "reset", "reset") + "\n"))
								connection.session.character.coins += int(sp)
								connection.store.Write([]byte("\n\x01SELF coins:" + strconv.Itoa(connection.session.character.coins) + "\n"))
								itemSold = true
								break
							}
						}
						if !itemSold {
							stream.Write([]byte("\n  Item not found!\n"))
						}
						break
					}
				}
				if !merchantFound {
					stream.Write([]byte("\n  Entity not found!\n"))
				}
			}

		// all the case 1s or 2s
		case "exit", "quit":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "look", "l":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "north", "n":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "south", "s":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "west", "w":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "east", "e":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "inventory", "inv", "i":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "equipped", "worn", "armor":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "effects", "effs":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "profile", "pf":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "help", "h":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <query>\n"))
		case "pickup", "take", "pick", "collect":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "drop", "put", "throw":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "examine":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "equip", "wear", "eq":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "unequip", "remove", "uneq":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <slot>\n"))
		case "use", "u":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "fight", "attack", "kick", "f":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <entity>\n"))
		case "train":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <stat_name>\n"))
		case "list", "browse":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <entity>\n"))
		case "toggle", "switch":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <option>\n"))
		case "move", "go", "m":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <n|s|w|e>\n"))
		}

	default:
		switch cmdTokens[0] {
		case "exit", "quit":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "look", "l":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "north", "n":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "south", "s":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "west", "w":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "east", "e":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "inventory", "inv", "i":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "equipped", "worn", "armor":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "effects", "effs":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "profile", "pf":
			stream.Write([]byte("\n  Too many arguments! Try just " + cmdTokens[0] + "\n"))
		case "help", "h":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <query>\n"))
		case "pickup", "take", "pick", "collect":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "drop", "put", "throw":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "examine":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "equip", "wear", "eq":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "unequip", "remove", "uneq":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <slot>\n"))
		case "use", "u":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item>\n"))
		case "fight", "attack", "kick", "f":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <entity>\n"))
		case "train":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <stat_name>\n"))
		case "list", "browse":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <entity>\n"))
		case "toggle", "switch":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <option>\n"))
		case "move", "go", "m":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <n|s|w|e>\n"))
		case "buy":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <item_id> <entity>\n"))
		case "sell":
			stream.Write([]byte("\n  Too many arguments! Try " + cmdTokens[0] + " <index_item> <entity>\n"))
		}
	}
	return 1
}

func CommandsCombat(cmdTokens []string, db *sql.DB, world *World, connection *ConnectionData) {
	stream := connection.store
	switch len(cmdTokens) {
	case 2:
		switch cmdTokens[0] {
		case "use", "u":
			var itemT *ItemTemplate
			var itemM *Item
			valid, i := validateInt(cmdTokens[1])
			if !valid {
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped {
						keywords := strings.FieldsFunc(strings.ToLower(world.ItemTemplates[item.templateID].name), func(r rune) bool {
							return r == ' ' || r == ','
						})
						matchBool := false
						for _, k := range keywords {
							if k == cmdTokens[1] {
								matchBool = true
							}
						}
						e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
						if matchBool && !slices.Contains(e, world.ItemTemplates[item.templateID].itype) {
							itemT = world.ItemTemplates[item.templateID]
							itemM = item
							break
						}
					}
				}
			} else {
				var playerItems []int
				e := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
				for _, item := range world.items {
					if item.locationType == "player" && item.locationID == connection.session.id && !item.equipped && !slices.Contains(e, world.ItemTemplates[item.templateID].itype) {
						playerItems = append(playerItems, item.id)
					}
				}
				if len(playerItems) > i && len(playerItems) != 0 {
					itemT = world.ItemTemplates[world.items[playerItems[i]].templateID]
					itemM = world.items[playerItems[i]]
				}
			}
			if itemT != nil {
				for _, effect := range itemT.effects {
					switch effect.effect {
					case "hp":
						if connection.session.character.hp != connection.session.character.maxHp {
							stream.Write([]byte("\n  You use a " + color(connection, "cyan", "tp") + itemT.name + color(connection, "reset", "reset") + " to heal " + color(connection, "red", "tp") + strconv.Itoa(effect.value) + color(connection, "reset", "reset") + " hp! (" + strconv.Itoa(connection.session.character.hp) + "/" + strconv.Itoa(connection.session.character.maxHp) + ")\n"))
							if connection.isClientWeb {
								connection.store.Write([]byte("\x01SELF " + "hp:" + strconv.Itoa(connection.session.character.hp) + " coins:" + strconv.Itoa(connection.session.character.coins) + "\n"))
							}
							connection.session.character.hp += effect.value
							_, err := db.Exec("DELETE FROM items WHERE id = ?", itemM.id)
							delete(world.items, itemM.id)
							fmt.Println(err)
						} else {
							stream.Write([]byte("\n  Your health is already full!\n"))
						}
					}
				}
			} else {
				stream.Write([]byte("\n  Item not found!\n"))
			}
		}
	}
}
