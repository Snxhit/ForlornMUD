// TODOs in order
// TODO: Combat
// TODO: PK
// TODO: Colors

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// networking
const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "8899"
	CONN_TYPE = "tcp"
)

// stats
var (
	TotalConnections int = 0
	TotalPlayers     int = 0
)

var (
	TargetPlayer = "player"
	TargetEntity = "entity"
)

// state variables

// type structs
// world structs
type World struct {
	characters  []*Character
	nodeList    []Room
	items       map[int]*Item
	entities    map[int]*Entity
	connections []*ConnectionData
	mu          sync.Mutex
}

type Room struct {
	id          int
	name        string
	description string
	exits       [4]int
	entityIDs   []int
	itemIDs     []int
}

type Entity struct {
	id          int
	name        string
	description string
	aggro       bool
	targetID    *int
	inCombat    bool
	hp          int
	locationID  int
}

type Item struct {
	id           int
	name         string
	description  string
	locationType string
	locationID   int
}

type ConnectionData struct {
	ConnID  int
	store   net.Conn
	session *Session
}

type Session struct {
	authorized bool
	id         int
	username   string
	password   string
	loginctx   *LoginContext
	character  *Character
}

type LoginContext struct {
	newPlayer  bool
	id         int
	username   string
	password   string
	hp         int
	locationID int
}

type Character struct {
	worldID    int
	hp         int
	locationID int
	targetType *string
	targetID   *int
	inCombat   bool

	conn *ConnectionData
}

func main() {
	db, err := sql.Open("sqlite", "game.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec(`
		CREATE TABLE IF NOT EXISTS players (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			hp INT,
			locationID INT
		)
	`)
	db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING NOT NULL,
			locationType STRING,
			locationID INT
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING NOT NULL,
			desc STRING NOT NULL,
			hp INT,
			locationID INT
		)
	`)

	db.Exec(
		"INSERT OR IGNORE INTO players (username, password, hp, locationID) VALUES (?, ?, ?, ?)",
		"testplayer", "hello", 100, 0,
	)

	db.Exec(
		"INSERT OR IGNORE INTO items (name, locationType, locationID) VALUES (?, ?, ?)",
		"Rusted Spoon", "room", 1,
	)

	db.Exec(
		"INSERT OR IGNORE INTO entities (name, desc, hp, locationID) VALUES (?, ?, ?, ?)",
		"Green Slime", "Looks goeey.", 10, 0,
	)

	roomone := Room{0, "Green Glade", "You look around to see tall standing trees towering over you...", [4]int{1, -1, -1, -1}, []int{}, []int{}}
	roomtwo := Room{1, "Stone Pathway", "Looks like a long pathway. Wonder where it goes.", [4]int{-1, 0, -1, -1}, []int{}, []int{}}

	var rlist []Room
	rlist = append(rlist, roomone)
	rlist = append(rlist, roomtwo)
	world := World{[]*Character{}, rlist, map[int]*Item{}, map[int]*Entity{}, []*ConnectionData{}, sync.Mutex{}}

	ent_rows, err := db.Query("SELECT id, name, desc, hp, locationID FROM entities")
	if err != nil {
		fmt.Println(err)
	}
	defer ent_rows.Close()

	for ent_rows.Next() {
		var eID int
		var eName string
		var eDesc string
		var eHp int
		var lID int
		ent_rows.Scan(&eID, &eName, &eDesc, &eHp, &lID)

		ent := Entity{eID, eName, eDesc, false, nil, false, eHp, lID}
		world.entities[eID] = &ent
		world.nodeList[lID].entityIDs = append(world.nodeList[lID].entityIDs, ent.id)
	}

	rows, err := db.Query("SELECT id, name, locationType, locationID FROM items")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var iID int
		var iName string
		var lType string
		var lID int
		rows.Scan(&iID, &iName, &lType, &lID)
		world.items[iID] = &Item{iID, iName, "", lType, lID}
		switch lType {
		case "room":
			itemID := world.items[iID].id
			world.nodeList[lID].itemIDs = append(world.nodeList[lID].itemIDs, itemID)
		case "player":
			itemID := world.items[iID].id
			world.nodeList[0].itemIDs = append(world.nodeList[0].itemIDs, itemID)
		}
	}

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Listening successfully on: " + CONN_HOST + ":" + CONN_PORT)

	go func() {
		for {
			newConn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting connection: ", err.Error())
				os.Exit(1)
			}

			sesh := Session{false, 0, "", "", &LoginContext{false, 0, "", "", 100, 0}, nil}
			conn := ConnectionData{TotalConnections, newConn, &sesh}
			TotalConnections += 1
			world.connections = append(world.connections, &conn)
			fmt.Println(conn)
			go HandleNewClient(&conn, &world, db)
		}
	}()

	worldTicker := time.NewTicker(5 * time.Second)
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
				if conn.session.character.hp < 100 && !conn.session.character.inCombat {
					conn.session.character.hp += 5
				}
				if conn.session.character.hp > 100 {
					conn.session.character.hp = 100
				}
				fmt.Println(conn.session.character)
				if conn.session.character.inCombat {
					fmt.Println(conn.session.username)
					if conn.session.character.targetID == nil {
						conn.session.character.inCombat = false
						continue
					}
					if conn.session.character.targetType == &TargetEntity {
						pDam := rand.Intn(7) + 3
						eDam := rand.Intn(5)
						world.entities[*conn.session.character.targetID].hp -= pDam
						conn.session.character.hp -= eDam
						conn.store.Write([]byte("\nYou damage the " + world.entities[*conn.session.character.targetID].name + " for " + strconv.Itoa(pDam) + " (" + strconv.Itoa(world.entities[*conn.session.character.targetID].hp) + ")" + "\n"))
						conn.store.Write([]byte("The " + world.entities[*conn.session.character.targetID].name + " damages you for " + strconv.Itoa(eDam) + " (" + strconv.Itoa(conn.session.character.hp) + ")" + "\n"))
						if conn.session.character.hp <= 0 {
							conn.store.Write([]byte("\nYou died!\n\n> "))
							world.entities[*conn.session.character.targetID].inCombat = false
							world.entities[*conn.session.character.targetID].targetID = nil
							conn.session.character.inCombat = false
							conn.session.character.targetID = nil
							conn.session.character.targetType = nil
							continue
						}
						if world.entities[*conn.session.character.targetID].hp <= 0 {
							conn.store.Write([]byte("\nYou killed a " + world.entities[*conn.session.character.targetID].name + "!\n\n> "))
							world.entities[*conn.session.character.targetID].inCombat = false
							world.entities[*conn.session.character.targetID].targetID = nil
							conn.session.character.inCombat = false
							conn.session.character.targetID = nil
							conn.session.character.targetType = nil
							continue
						}
					} else if conn.session.character.targetType == &TargetPlayer {
						p1Dam := rand.Intn(40)
						p2Dam := rand.Intn(40)
						p2Chr := world.characters[*conn.session.character.targetID]
						p1Chr := conn.session.character
						if p1Chr.worldID > p2Chr.worldID {
							continue
						}
						p2Chr.hp -= p1Dam
						p1Chr.hp -= p2Dam
						conn.store.Write([]byte("\nYou damage " + p2Chr.conn.session.username + " for " + strconv.Itoa(p1Dam) + " (" + strconv.Itoa(p2Chr.hp) + ")"))
						conn.store.Write([]byte("\n" + conn.session.username + " damages you for " + strconv.Itoa(p2Dam) + " (" + strconv.Itoa(p1Chr.hp) + ")" + "\n"))
						p2Chr.conn.store.Write([]byte("\nYou damage " + p1Chr.conn.session.username + " for " + strconv.Itoa(p2Dam) + " (" + strconv.Itoa(p1Chr.hp) + ")"))
						p2Chr.conn.store.Write([]byte("\n" + conn.session.username + " damages you for " + strconv.Itoa(p1Dam) + " (" + strconv.Itoa(p2Chr.hp) + ")" + "\n"))
						if p2Chr.hp <= 0 {
							p2Chr.inCombat = false
							p2Chr.targetID = nil
							p2Chr.targetType = nil
							p1Chr.inCombat = false
							p1Chr.targetID = nil
							p1Chr.targetType = nil
							p2Chr.conn.store.Write([]byte("\nYou died!\n\n> "))
							conn.store.Write([]byte("\nYou killed " + p2Chr.conn.session.username + "!\n\n> "))
							continue
						} else if p1Chr.hp <= 0 {
							p2Chr.inCombat = false
							p2Chr.targetID = nil
							p2Chr.targetType = nil
							p1Chr.inCombat = false
							p1Chr.targetID = nil
							p1Chr.targetType = nil
							p2Chr.conn.store.Write([]byte("\nYou killed " + conn.session.username + "!\n\n> "))
							conn.store.Write([]byte("\nYou died!\n\n> "))
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

func HandleNewClient(connection *ConnectionData, world *World, db *sql.DB) {
	stream := connection.store
	stream.Write([]byte("What is your name?\n> "))
	for {
		buf := make([]byte, 2048)
		_, err := stream.Read(buf)
		if err != nil {
			fmt.Println("Error reading input: ", err.Error())
		}

		bufCmd := bytes.Trim(buf, string([]byte{0}))
		// if the user interrupts
		if len(bufCmd) == 0 {
			HandleClientDisconnect(connection, world, db)
			return
		}
		strCmd := strings.ToLower(string(bufCmd[0 : len(bufCmd)-1]))
		cmdTokens := strings.Split(strCmd, " ")

		if !connection.session.authorized {
			if len(cmdTokens) != 1 {
				stream.Write([]byte("No inputs can have spaces!"))
			} else if len(cmdTokens) == 1 && len(connection.session.username) == 0 {
				row := db.QueryRow("SELECT id, username, password, hp, locationID FROM players WHERE username = ?", cmdTokens[0])

				err := row.Scan(&connection.session.loginctx.id, &connection.session.loginctx.username, &connection.session.loginctx.password, &connection.session.loginctx.hp, &connection.session.loginctx.locationID)
				if err != nil {
					stream.Write([]byte("User " + cmdTokens[0] + " not found, creating user.\n"))
					stream.Write([]byte("Please enter a new password."))
					connection.session.username = cmdTokens[0]
					connection.session.loginctx.newPlayer = true
				} else {
					connection.session.loginctx.newPlayer = false
					connection.session.username = cmdTokens[0]
					stream.Write([]byte("User " + cmdTokens[0] + " found, what is your password?"))
				}
			} else if len(cmdTokens) == 1 && len(connection.session.username) != 0 {
				fmt.Println(connection.session.loginctx.password)
				fmt.Println(connection.session.loginctx.newPlayer)
				if !connection.session.loginctx.newPlayer {
					if cmdTokens[0] == connection.session.loginctx.password {
						stream.Write([]byte("Correct password! Welcome, " + connection.session.username + "!\n"))
						connection.session.password = connection.session.loginctx.password
						connection.session.id = connection.session.loginctx.id
						connection.session.authorized = true
						TotalPlayers += 1
						world.mu.Lock()
						world.characters = append(world.characters, &Character{len(world.characters), connection.session.loginctx.hp, connection.session.loginctx.locationID, nil, nil, false, connection})
						fmt.Println(world.characters)
						world.mu.Unlock()
						connection.session.character = world.characters[len(world.characters)-1]
						stream.Write([]byte("Welcome to the MUD!\n\n"))
						HandleMovement(connection, world)
					} else {
						stream.Write([]byte("Wrong password!"))
					}
				} else {
					stream.Write([]byte("Welcome, " + connection.session.username + " to the game!\n"))
					world.mu.Lock()
					world.characters = append(world.characters, &Character{len(world.characters), connection.session.loginctx.hp, connection.session.loginctx.locationID, nil, nil, false, connection})
					fmt.Println(world.characters)
					connection.session.character = world.characters[len(world.characters)-1]
					world.mu.Unlock()
					prom, err := db.Exec("INSERT INTO players (username, password, hp, locationID) VALUES (?, ?, ?, ?)", connection.session.username, cmdTokens[0], connection.session.character.hp, connection.session.character.locationID)
					fmt.Println(err)
					id, _ := prom.LastInsertId()
					connection.session.password = cmdTokens[0]
					connection.session.id = int(id)
					connection.session.authorized = true
					TotalPlayers += 1
					stream.Write([]byte("Welcome to the MUD!\n\n"))
				}
			}

			stream.Write([]byte("\n> "))
		} else {
			if !connection.session.character.inCombat {
				switch len(cmdTokens) {
				case 1:
					switch cmdTokens[0] {
					case "exit":
						HandleClientDisconnect(connection, world, db)
						return
					case "selfharm":
						connection.session.character.hp -= 10
						fmt.Println(connection.session.character.hp)
						connection.store.Write([]byte("Ow! You poke yourself and lose 10 hp."))
					case "look":
						dirs := [4]string{"north", "south", "west", "east"}
						for dir, id := range world.nodeList[connection.session.character.locationID].exits {
							if world.nodeList[connection.session.character.locationID].exits[dir] != -1 {
								stream.Write([]byte("  - " + dirs[dir] + ": " + world.nodeList[id].name + "\n"))
							} /* else {
								stream.Write([]byte("  - " + dirs[dir] + ": " + "none\n"))
							}*/
						}
						for _, entID := range world.nodeList[connection.session.character.locationID].entityIDs {
							stream.Write([]byte(world.entities[entID].name + "\n   " + world.entities[entID].description + "\n"))
						}
						for _, itemID := range world.nodeList[connection.session.character.locationID].itemIDs {
							stream.Write([]byte("1x - " + world.items[itemID].name + "\n"))
						}
						for _, conn := range world.connections {
							if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
								stream.Write([]byte("* " + conn.session.username + " looks at you with lust."))
							}
						}

					case "north":
						if world.nodeList[connection.session.character.locationID].exits[0] != -1 {
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " left towards north.\n"))
								}
							}
							connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[0]
							HandleMovement(connection, world)
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " entered from south.\n"))
								}
							}
						}
					case "south":
						if world.nodeList[connection.session.character.locationID].exits[1] != -1 {
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " left towards south.\n"))
								}
							}
							connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[1]
							HandleMovement(connection, world)
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " entered from north.\n"))
								}
							}
						}
					case "west":
						if world.nodeList[connection.session.character.locationID].exits[2] != -1 {
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " left towards west.\n"))
								}
							}
							connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[2]
							HandleMovement(connection, world)
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " entered from east.\n"))
								}
							}
						}
					case "east":
						if world.nodeList[connection.session.character.locationID].exits[3] != -1 {
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " left towards east.\n"))
								}
							}
							connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[3]
							HandleMovement(connection, world)
							for _, conn := range world.connections {
								if conn.session.character.locationID == connection.session.character.locationID && conn.session.id != connection.session.id {
									conn.store.Write([]byte("! " + connection.session.username + " entered from west.\n"))
								}
							}
						}

					case "inventory":
						for _, item := range world.items {
							if world.items[item.id].locationType == "player" && world.items[item.id].locationID == connection.session.id {
								stream.Write([]byte("  " + world.items[item.id].name + "\n"))
							}
						}
					default:
						stream.Write([]byte("Command not found!"))
					}

				case 2:
					switch cmdTokens[0] {
					case "echo":
						stream.Write([]byte(cmdTokens[1]))
					case "pickup":
						i, err := strconv.Atoi(cmdTokens[1])
						if err != nil {
							stream.Write([]byte("Invalid syntax. Please provide an integer!"))
						} else {
							if len(world.nodeList[connection.session.character.locationID].itemIDs) > i {
								stream.Write([]byte("You pick up a " + world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].name))
								world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationType = "player"
								world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].locationID = connection.session.id
								world.nodeList[connection.session.character.locationID].itemIDs = slices.Delete(world.nodeList[connection.session.character.locationID].itemIDs, i, i+1)
								_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "player", connection.session.id, world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].id)
								fmt.Println(err)
							} else {
								stream.Write([]byte("There is no such item."))
							}
						}
					case "drop":
						i, err := strconv.Atoi(cmdTokens[1])
						if err != nil {
							stream.Write([]byte("Invalid syntax. Please provide an integer!"))
						} else {
							var playerItems []int
							for _, item := range world.items {
								if item.locationType == "player" && item.locationID == connection.session.id {
									playerItems = append(playerItems, item.id)
								}
							}
							if len(playerItems) > i {
								stream.Write([]byte("You dropped a " + world.items[world.nodeList[connection.session.character.locationID].itemIDs[i]].name))
								world.items[playerItems[i]].locationType = "room"
								world.items[playerItems[i]].locationID = world.nodeList[connection.session.character.locationID].id
								world.nodeList[connection.session.character.locationID].itemIDs = append(world.nodeList[connection.session.character.locationID].itemIDs, playerItems[i])
								_, err := db.Exec("UPDATE items SET (locationType, locationID) = (?, ?) WHERE id = ?", "room", connection.session.character.locationID, playerItems[i])
								fmt.Println(err)
							} else {
								stream.Write([]byte("There is no such item."))
							}
						}
					case "fight":
						i, err := strconv.Atoi(cmdTokens[1])
						if err != nil {
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
								stream.Write([]byte("Engaging a " + world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].name))
								world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].inCombat = true
								world.entities[world.nodeList[connection.session.character.locationID].entityIDs[i]].targetID = &connection.session.id
								connection.session.character.inCombat = true
								connection.session.character.targetID = &world.nodeList[connection.session.character.locationID].entityIDs[i]
								connection.session.character.targetType = &TargetEntity
							}
						}
					default:
						stream.Write([]byte("Command not found!"))
					}
				}
			} else {
			}

			stream.Write([]byte("\n> "))
		}
	}
}

func HandleClientDisconnect(connection *ConnectionData, world *World, db *sql.DB) {
	world.mu.Lock()
	defer world.mu.Unlock()

	connection.store.Write([]byte("Exiting game."))
	connection.store.Close()
	if connection.session.authorized {
		TotalPlayers -= 1
		fmt.Println(connection.session.character.hp)
		_, err := db.Exec("UPDATE players SET (hp, locationID) = (?, ?) WHERE id = ?", connection.session.character.hp, connection.session.character.locationID, connection.session.id)
		fmt.Println(err)
		world.characters[connection.session.character.worldID] = &Character{connection.session.character.worldID, 100, 0, nil, nil, false, nil}

		for i, c := range world.connections {
			if c == connection {
				world.connections = append(world.connections[:i], world.connections[i+1:]...)
				break
			}
		}

		connection.session = nil
	}
	TotalConnections -= 1
	fmt.Println(world.characters)
}

func HandleMovement(connection *ConnectionData, world *World) {
	r := world.nodeList[connection.session.character.locationID]
	connection.store.Write([]byte("\033[32m" + r.name + "\033[0m \n"))
	connection.store.Write([]byte(r.description))
}
