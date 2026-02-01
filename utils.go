package main

import (
	"database/sql"
	"fmt"
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

func DeleteItem(connection *ConnectionData, world *World, db *sql.DB, iID int) {
	db.Exec("DELETE FROM items WHERE id = ?", iID)
	delete(world.items, iID)
}

func HandleMovement(connection *ConnectionData, world *World) {
	r := world.nodeList[connection.session.character.locationID]
	fmt.Println(r.itemIDs)
	connection.store.Write([]byte("\n\033[32m" + r.name + "\033[0m \n"))
	connection.store.Write([]byte("  " + r.description + "\n\n"))
	HandleLook(world, connection)
}

func HandleLook(world *World, connection *ConnectionData) {
	stream := connection.store
	stream.Write([]byte("Exits\n"))
	dirs := [4]string{"north", "south", "west", "east"}
	for dir, id := range world.nodeList[connection.session.character.locationID].exits {
		if world.nodeList[connection.session.character.locationID].exits[dir] != -1 {
			stream.Write([]byte("  -> " + dirs[dir] + ": " + world.nodeList[id].name + "\n"))
		} /* else {
			stream.Write([]byte("  - " + dirs[dir] + ": " + "none\n"))
		}*/
	}
	iIDs := map[int]int{}
	for _, itemID := range world.nodeList[connection.session.character.locationID].itemIDs {
		iIDs[world.items[itemID].templateID] += 1
	}
	fmt.Println(iIDs)
	for tID, num := range iIDs {
		s := strconv.Itoa(num)
		stream.Write([]byte("  " + s + strings.Repeat(" ", 3-len(s)) + " | " + world.ItemTemplates[tID].name + "\n"))
	}

	for _, entID := range world.nodeList[connection.session.character.locationID].entityIDs {
		stream.Write([]byte("  " + world.EntityTemplates[world.entities[entID].templateID].name + "\n    " + world.EntityTemplates[world.entities[entID].templateID].description + "\n"))
	}
	for _, conn := range world.connections {
		if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
			stream.Write([]byte("  * " + conn.session.username + " looks at you with lust.\n"))
		}
	}
}

func LeftNotifier(world *World, connection *ConnectionData, dir string) {
	for _, conn := range world.connections {
		if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
			conn.store.Write([]byte("\x1b[2K\r  ! " + connection.session.username + " left towards " + dir + ".\n\n> "))
		}
	}
}

func EnterNotifier(world *World, connection *ConnectionData, dir string) {
	for _, conn := range world.connections {
		if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
			conn.store.Write([]byte("\x1b[2K\r  ! " + connection.session.username + " entered from " + dir + ".\n\n> "))
		}
	}
}

// cuz theres no clamp func in the stdlib???
func Clamp(i int, floor int, ceiling int) int {
	a := min(i, ceiling)
	return max(a, floor)
}
