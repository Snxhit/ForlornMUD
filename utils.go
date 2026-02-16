package main

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
)

func validateInt(input string) (bool, int) {
	i, err := strconv.Atoi(input)
	if err != nil {
		return false, 0
	}
	if i < 0 {
		return false, 0
	}
	return true, i
}

func CreateAndInsertItem(connection *ConnectionData, world *World, db *sql.DB, tID int) {
	prom, _ := db.Exec(
		"INSERT INTO items (templateID, locationType, locationID, equipped) VALUES (?, ?, ?, ?)",
		tID, "player", connection.session.id, false,
	)
	x, _ := prom.LastInsertId()
	world.items[int(x)] = &Item{int(x), tID, "player", connection.session.id, false}
}

// THIS IS BUGGED :(
// idek whats wrong :(
func CreateAndInsertItemBatched(connection *ConnectionData, world *World, db *sql.DB, tID int, qty int) {
	vals := ""
	args := make([]any, 0, qty*4)

	for i := 0; i < qty; i++ {
		if i > 0 {
			vals += ","
		}
		vals += "(?, ?, ?, ?)"
		args = append(args, tID, "player", connection.session.id, false)
	}

	query := "INSERT INTO items (templateID, locationType, locationID, equipped) VALUES " + vals

	prom, err := db.Exec(query, args...)
	if err != nil {
		return
	}

	firstID, _ := prom.LastInsertId()

	for i := 0; i < qty; i++ {
		id := int(firstID) + i
		world.items[id] = &Item{id, tID, "player", connection.session.id, false}
	}
}

func CreateAndPlaceItem(world *World, db *sql.DB, tID int, lID int) {
	prom, _ := db.Exec(
		"INSERT INTO items (templateID, locationType, locationID, equipped) VALUES (?, ?, ?, ?)",
		tID, "room", lID, false,
	)
	x, _ := prom.LastInsertId()
	world.items[int(x)] = &Item{int(x), tID, "room", lID, false}
	world.nodeList[lID].itemIDs = append(world.nodeList[lID].itemIDs, int(x))
}

func DeleteItem(connection *ConnectionData, world *World, db *sql.DB, iID int) {
	db.Exec("DELETE FROM items WHERE id = ?", iID)
	delete(world.items, iID)
}

func SpawnAndInsertEntity(world *World, db *sql.DB, lID int, tID int) {
	prom, _ := db.Exec(
		"INSERT OR IGNORE INTO entities (templateID, hp, locationID) VALUES (?, ?, ?)",
		tID, 0, lID,
	)
	x, _ := prom.LastInsertId()
	world.entities[int(x)] = &Entity{int(x), tID, nil, false, world.EntityTemplates[tID].maxHp, lID}
	world.nodeList[lID].entityIDs = append(world.nodeList[lID].entityIDs, int(x))
}

func HandleMovement(connection *ConnectionData, world *World) {
	r := world.nodeList[connection.session.character.locationID]
	//connection.store.Write([]byte("\n\033[32m" + r.name + "\033[0m \n"))
	connection.store.Write([]byte("\n" + color(connection, "green", "tp") + r.name + color(connection, "reset", "reset") + " \n"))
	connection.store.Write([]byte("  " + r.description + "\n\n"))
	HandleLook(world, connection)
}

func HandleLook(world *World, connection *ConnectionData) {
	stream := connection.store
	var state string
	stream.Write([]byte(color(connection, "green", "tp") + "Exits\n" + color(connection, "reset", "reset")))
	dirs := [4]string{"north", "south", "west", "east"}
	for dir, id := range world.nodeList[connection.session.character.locationID].exits {
		if world.nodeList[connection.session.character.locationID].exits[dir] != -1 {
			stream.Write([]byte("  " + color(connection, "cyan", "tp") + glphys(connection, "a"+string(dirs[dir][0])) + " " + color(connection, "yellow", "tp") + dirs[dir] + color(connection, "reset", "reset") + ": " + world.nodeList[id].name + "\n"))
		} /* else {
			stream.Write([]byte("  - " + dirs[dir] + ": " + "none\n"))
		}*/
		state += dirs[dir] + ":" + strconv.Itoa(id) + " "
	}
	stream.Write([]byte(color(connection, "green", "tp") + "\nSurroundings\n" + color(connection, "reset", "reset")))
	iIDs := map[int]int{}
	for _, itemID := range world.nodeList[connection.session.character.locationID].itemIDs {
		iIDs[world.items[itemID].templateID] += 1
	}
	for tID, num := range iIDs {
		s := strconv.Itoa(num)
		stream.Write([]byte("  " + color(connection, "cyan", "tp") + s + color(connection, "reset", "reset") + strings.Repeat(" ", 3-len(s)) + " | " + world.ItemTemplates[tID].name + "\n"))
	}

	for _, entID := range world.nodeList[connection.session.character.locationID].entityIDs {
		if !world.entities[entID].inCombat && world.merchants[entID] == nil {
			stream.Write([]byte("  " + color(connection, "cyan", "tp") + world.EntityTemplates[world.entities[entID].templateID].name + color(connection, "reset", "reset") + "\n    " + world.EntityTemplates[world.entities[entID].templateID].description + "\n"))
		} else if !world.entities[entID].inCombat && world.merchants[entID] != nil {
			stream.Write([]byte("  " + color(connection, "cyan", "tp") + world.EntityTemplates[world.entities[entID].templateID].name + color(connection, "green", "tp") + " M" + color(connection, "reset", "reset") + "\n    " + world.EntityTemplates[world.entities[entID].templateID].description + "\n"))
		} else if world.entities[entID].inCombat && world.merchants[entID] == nil {
			stream.Write([]byte(color(connection, "red", "tp") + "C " + color(connection, "cyan", "tp") + world.EntityTemplates[world.entities[entID].templateID].name + color(connection, "reset", "reset") + "\n    " + world.EntityTemplates[world.entities[entID].templateID].description + "\n"))
		} else if world.entities[entID].inCombat && world.merchants[entID] != nil {
			stream.Write([]byte(color(connection, "red", "tp") + "C " + color(connection, "cyan", "tp") + world.EntityTemplates[world.entities[entID].templateID].name + color(connection, "green", "tp") + " M" + color(connection, "reset", "reset") + "\n    " + world.EntityTemplates[world.entities[entID].templateID].description + "\n"))
		}
	}
	for _, conn := range world.connections {
		if conn.session.authorized && conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
			if !conn.session.character.inCombat {
				stream.Write([]byte(color(connection, "cyan", "tp") + "  * " + color(connection, "yellow", "tp") + conn.session.username + color(connection, "reset", "reset") + " looks at you.\n"))
			} else {
				stream.Write([]byte(color(connection, "red", "tp") + "C" + color(connection, "cyan", "tp") + " * " + color(connection, "yellow", "tp") + conn.session.username + color(connection, "reset", "reset") + " looks at you.\n"))
			}
		}
	}
	if connection.isClientWeb {
		stream.Write([]byte("\n\x01EXITS " + state + "\n"))
	}
}

func LeftNotifier(world *World, connection *ConnectionData, dir string) {
	for _, conn := range world.connections {
		if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
			conn.store.Write([]byte("\x1b[2K\r  " + color(conn, "cyan", "tp") + " " + color(conn, "yellow", "tp") + connection.session.username + color(conn, "reset", "reset") + " left towards " + dir + ".\n\n> "))
		}
	}
}

func EnterNotifier(world *World, connection *ConnectionData, dir string) {
	for _, conn := range world.connections {
		if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
			conn.store.Write([]byte("\x1b[2K\r  " + color(conn, "cyan", "tp") + " " + color(conn, "yellow", "tp") + connection.session.username + color(conn, "reset", "reset") + " entered from " + dir + ".\n\n> "))
		}
	}
}

// cuz theres no clamp func in the stdlib???
func Clamp(i int, floor int, ceiling int) int {
	a := min(i, ceiling)
	return max(a, floor)
}

func calcExpMultiplier(diff int) float64 {
	switch {
	case diff >= 5:
		return 1.5
	case diff >= 3:
		return 1.25
	case diff >= 1:
		return 1.2
	case diff == 0:
		return 1
	case diff >= -2:
		return 0.7
	case diff >= -4:
		return 0.5
	default:
		return 0.2
	}
}

func printProfileCard(connection *ConnectionData, nameMedian int, c string, t string, lvl string, exp string, expBars int, str string, dex string, agi string, stam string, int string, cardLength int, eList [6]string) {
	stream := connection.store

	if !connection.isPrettyEnabled {
		us := "  | " + strings.Repeat(" ", cardLength/2-nameMedian) + color(connection, "green", "tp") + connection.session.username + color(connection, "reset", "reset")
		stream.Write([]byte("\n  +" + strings.Repeat("-", cardLength) + "+\n"))
		stream.Write([]byte(us + strings.Repeat(" ", cardLength-visibleLen(us)+2) + " |\n"))
		stream.Write([]byte("  +" + strings.Repeat("-", 27) + "+" + strings.Repeat("-", cardLength-28) + "+\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", 27) + "|" + strings.Repeat(" ", cardLength-28) + "|\n"))
		stream.Write([]byte("  | " + "stats:" + strings.Repeat(" ", 19) + " | "))
		stream.Write([]byte("equipment:" + strings.Repeat(" ", cardLength-40) + " |\n"))
		stream.Write([]byte("  | " + color(connection, "cyan", "tp") + "  Strength     :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(str)) + str + " ]" + " | "))
		stream.Write([]byte(eList[0] + strings.Repeat(" ", cardLength-30-visibleLen(eList[0])) + " |\n"))
		stream.Write([]byte("  | " + color(connection, "cyan", "tp") + "  Dexterity    :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(dex)) + dex + " ]" + " | "))
		stream.Write([]byte(eList[1] + strings.Repeat(" ", cardLength-30-visibleLen(eList[1])) + " |\n"))
		stream.Write([]byte("  | " + color(connection, "cyan", "tp") + "  Agility      :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(agi)) + agi + " ]" + " | "))
		stream.Write([]byte(eList[2] + strings.Repeat(" ", cardLength-30-visibleLen(eList[2])) + " |\n"))
		stream.Write([]byte("  | " + color(connection, "cyan", "tp") + "  Stamina      :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(stam)) + stam + " ]" + " | "))
		stream.Write([]byte(eList[3] + strings.Repeat(" ", cardLength-30-visibleLen(eList[3])) + " |\n"))
		stream.Write([]byte("  | " + color(connection, "cyan", "tp") + "  Intelligence :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(int)) + int + " ]" + " | "))
		stream.Write([]byte(eList[4] + strings.Repeat(" ", cardLength-30-visibleLen(eList[4])) + " |\n"))
		stream.Write([]byte("  | " + strings.Repeat(" ", 25) + " | "))
		stream.Write([]byte(eList[5] + strings.Repeat(" ", cardLength-30-visibleLen(eList[5])) + " |\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", 27) + "|" + strings.Repeat(" ", cardLength-28) + "|\n"))
		stream.Write([]byte("  +" + strings.Repeat("-", 20) + "+" + strings.Repeat("-", 6) + "+" + strings.Repeat("-", cardLength-28) + "+\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", 20) + "|" + strings.Repeat(" ", cardLength-21) + "|\n"))
		stream.Write([]byte("  |" + color(connection, "yellow", "tp") + " Coins:" + color(connection, "reset", "reset") + " [ " + color(connection, "yellow", "tp") + strings.Repeat("0", 7-len(c)) + c + color(connection, "reset", "reset") + " ] |" + " Level Info :" + strings.Repeat(" ", cardLength-34) + "|\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", 20) + "|" + color(connection, "blue", "tp") + "   Level    :" + color(connection, "reset", "reset") + " [ " + strings.Repeat("0", 4-len(lvl)) + lvl + " ]" + strings.Repeat(" ", cardLength-43) + "|\n"))
		stream.Write([]byte("  +" + strings.Repeat("-", 20) + "+" + color(connection, "blue", "tp") + "   Exp      :" + color(connection, "reset", "reset") + " [ " + strings.Repeat("0", 7-len(exp)) + exp + " ]" + strings.Repeat(" ", cardLength-46) + "|\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", 20) + "|" + color(connection, "blue", "tp") + "   Progress : " + color(connection, "reset", "reset") + strings.Repeat(" ", cardLength-35) + "|\n"))
		stream.Write([]byte("  |" + color(connection, "cyan", "tp") + " Trains:" + color(connection, "reset", "reset") + " [ " + color(connection, "cyan", "tp") + strings.Repeat("0", 6-len(t)) + t + color(connection, "reset", "reset") + " ] |" + "     [" + color(connection, "blue", "tp") + strings.Repeat("#", expBars) + color(connection, "reset", "reset") + strings.Repeat("-", 20-expBars) + "]" + strings.Repeat(" ", cardLength-48) + "|\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", 20) + "|" + strings.Repeat(" ", cardLength-21) + "|\n"))
		stream.Write([]byte("  +" + strings.Repeat("-", 20) + "+" + strings.Repeat("-", cardLength-21) + "+\n"))
	} else {
		us := "  │ " + strings.Repeat(" ", cardLength/2-nameMedian) + color(connection, "green", "tp") + connection.session.username + color(connection, "reset", "reset")
		stream.Write([]byte("\n  ╭" + strings.Repeat("─", cardLength) + "╮\n"))
		stream.Write([]byte(us + strings.Repeat(" ", cardLength-visibleLen(us)+4) + " │\n"))
		stream.Write([]byte("  ├" + strings.Repeat("─", 27) + "┬" + strings.Repeat("─", cardLength-28) + "┤\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", 27) + "│" + strings.Repeat(" ", cardLength-28) + "│\n"))
		stream.Write([]byte("  │ " + "stats:" + strings.Repeat(" ", 19) + " │ "))
		stream.Write([]byte("equipment:" + strings.Repeat(" ", cardLength-40) + " │\n"))
		stream.Write([]byte("  │ " + color(connection, "cyan", "tp") + "  Strength     :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(str)) + str + " ]" + " │ "))
		stream.Write([]byte(eList[0] + strings.Repeat(" ", cardLength-30-visibleLen(eList[0])) + " │\n"))
		stream.Write([]byte("  │ " + color(connection, "cyan", "tp") + "  Dexterity    :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(dex)) + dex + " ]" + " │ "))
		stream.Write([]byte(eList[1] + strings.Repeat(" ", cardLength-30-visibleLen(eList[1])) + " │\n"))
		stream.Write([]byte("  │ " + color(connection, "cyan", "tp") + "  Agility      :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(agi)) + agi + " ]" + " │ "))
		stream.Write([]byte(eList[2] + strings.Repeat(" ", cardLength-30-visibleLen(eList[2])) + " │\n"))
		stream.Write([]byte("  │ " + color(connection, "cyan", "tp") + "  Stamina      :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(stam)) + stam + " ]" + " │ "))
		stream.Write([]byte(eList[3] + strings.Repeat(" ", cardLength-30-visibleLen(eList[3])) + " │\n"))
		stream.Write([]byte("  │ " + color(connection, "cyan", "tp") + "  Intelligence :" + color(connection, "reset", "reset") + "  [ " + strings.Repeat("0", 3-len(int)) + int + " ]" + " │ "))
		stream.Write([]byte(eList[4] + strings.Repeat(" ", cardLength-30-visibleLen(eList[4])) + " │\n"))
		stream.Write([]byte("  │ " + strings.Repeat(" ", 25) + " │ "))
		stream.Write([]byte(eList[5] + strings.Repeat(" ", cardLength-30-visibleLen(eList[5])) + " │\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", 27) + "│" + strings.Repeat(" ", cardLength-28) + "│\n"))
		stream.Write([]byte("  ├" + strings.Repeat("─", 20) + "┬" + strings.Repeat("─", 6) + "┴" + strings.Repeat("─", cardLength-28) + "┤\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", 20) + "│" + strings.Repeat(" ", cardLength-21) + "│\n"))
		stream.Write([]byte("  │" + color(connection, "yellow", "tp") + " Coins:" + color(connection, "reset", "reset") + " [ " + color(connection, "yellow", "tp") + strings.Repeat("0", 7-len(c)) + c + color(connection, "reset", "reset") + " ] │" + " Level Info :" + strings.Repeat(" ", cardLength-34) + "│\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", 20) + "│" + color(connection, "blue", "tp") + "   Level    :" + color(connection, "reset", "reset") + " [ " + strings.Repeat("0", 4-len(lvl)) + lvl + " ]" + strings.Repeat(" ", cardLength-43) + "│\n"))
		stream.Write([]byte("  ├" + strings.Repeat("─", 20) + "┤" + color(connection, "blue", "tp") + "   Exp      :" + color(connection, "reset", "reset") + " [ " + strings.Repeat("0", 7-len(exp)) + exp + " ]" + strings.Repeat(" ", cardLength-46) + "│\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", 20) + "│" + color(connection, "blue", "tp") + "   Progress : " + color(connection, "reset", "reset") + strings.Repeat(" ", cardLength-35) + "│\n"))
		stream.Write([]byte("  │" + color(connection, "cyan", "tp") + " Trains:" + color(connection, "reset", "reset") + " [ " + color(connection, "cyan", "tp") + strings.Repeat("0", 6-len(t)) + t + color(connection, "reset", "reset") + " ] │" + "     [" + color(connection, "blue", "tp") + strings.Repeat("█", expBars) + color(connection, "reset", "reset") + strings.Repeat("░", 20-expBars) + "]" + strings.Repeat(" ", cardLength-48) + "│\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", 20) + "│" + strings.Repeat(" ", cardLength-21) + "│\n"))
		stream.Write([]byte("  ╰" + strings.Repeat("─", 20) + "┴" + strings.Repeat("─", cardLength-21) + "╯\n"))
	}
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func visibleLen(s string) int {
	cleanStr := ansiRegex.ReplaceAllString(s, "")
	return len(cleanStr)
}

// nvm removes colors
func visibleSlice(s string, start int, end int) string {
	cleanStr := ansiRegex.ReplaceAllString(s, "")
	return cleanStr[start:end]
}

func glphys(conn *ConnectionData, glyph string) string {
	if conn.isPrettyEnabled {
		/*
		* convention:
		* sl - standing line
		* sll - sleeping line
		* tlc - top left corner
		* trc - top right corner
		* blc - bottom left corner
		* brc - bottom right corner
		* rtj - right T junction
		* ltj - left T junction
		* utj - up T junction
		* dtj - down T junction
		* an - arrow towards north
		* aw - arrow towards west
		* as - arrow towards south
		* ae - arrow towards east
		* sau - strong arrow up
		* sad - strong arrow down
		 */
		switch glyph {
		case "sl":
			return "│"
		case "sll":
			return "─"
		case "tlc":
			return "╭"
		case "trc":
			return "╮"
		case "blc":
			return "╰"
		case "brc":
			return "╯"
		case "rtj":
			return "├"
		case "ltj":
			return "┤"
		case "utj":
			return "┴"
		case "dtj":
			return "┬"
		case "an":
			return "↑"
		case "aw":
			return "←"
		case "as":
			return "↓"
		case "ae":
			return "→"
		case "sau":
			return "⬆"
		case "sad":
			return "⬇"
		}
	} else {
		switch glyph {
		case "sl":
			return "|"
		case "sll":
			return "-"
		case "tlc":
			return "+"
		case "trc":
			return "+"
		case "blc":
			return "+"
		case "brc":
			return "+"
		case "rtj":
			return "+"
		case "ltj":
			return "+"
		case "utj":
			return "+"
		case "dtj":
			return "+"
		case "an":
			return "^"
		case "aw":
			return "<"
		case "as":
			return "v"
		case "ae":
			return ">"
		case "sau":
			return "+"
		case "sad":
			return "-"
		}
	}
	return ""
}

func color(conn *ConnectionData, fg string, bg string) string {
	var color string
	if conn.isColorEnabled {

		switch fg {
		case "black":
			color += "\x1b[30m"
		case "red":
			color += "\x1b[31m"
		case "green":
			color += "\x1b[32m"
		case "yellow":
			color += "\x1b[33m"
		case "blue":
			color += "\x1b[34m"
		case "magenta":
			color += "\x1b[35m"
		case "cyan":
			color += "\x1b[36m"
		case "white":
			color += "\x1b[37m"
		case "reset":
			color += "\x1b[39m"
		}

		switch bg {
		case "tp": // transparent
			color += ""
		case "black":
			color += "\x1b[40m"
		case "red":
			color += "\x1b[41m"
		case "green":
			color += "\x1b[42m"
		case "yellow":
			color += "\x1b[43m"
		case "blue":
			color += "\x1b[44m"
		case "magenta":
			color += "\x1b[45m"
		case "cyan":
			color += "\x1b[46m"
		case "white":
			color += "\x1b[47m"
		case "reset":
			color += "\x1b[49m"
		}
	} else {
		color = ""
	}
	return color
}

func asciiGreeting(conn *ConnectionData) {
	conn.store.Write([]byte(`
+-------------------------------------------------------------+
|  ______         _                    __  __ _    _ _____    |
| |  ____|       | |                  |  \/  | |  | |  __ \   |
| | |__ ___  _ __| | ___  _ __ _ __   | \  / | |  | | |  | |  |
| |  __/ _ \| '__| |/ _ \| '__| '_ \  | |\/| | |  | | |  | |  |
| | | | (_) | |  | | (_) | |  | | | | | |  | | |__| | |__| |  |
| |_|  \___/|_|  |_|\___/|_|  |_| |_| |_|  |_|\____/|_____/   |
|                                                             |
|                   ╻ ╻┏━╸╻  ┏━╸┏━┓┏┳┓┏━╸                     |
+-----------------  ┃╻┃┣╸ ┃  ┃  ┃ ┃┃┃┃┣╸  --------------------+
                    ┗┻┛┗━╸┗━╸┗━╸┗━┛╹ ╹┗━╸`))
}
