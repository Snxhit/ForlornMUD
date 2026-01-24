// TODO: Colors, but after combat and items

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

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

// state variables

// type structs
// world structs
type World struct {
	characters  []Character
	nodeList    []Room
	connections []*ConnectionData
}

type Room struct {
	id          int
	name        string
	description string
	exits       [4]int
	entities    []Entity
	items       []Item
}

type Entity struct {
	id          int
	name        string
	description string
	aggro       bool
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
	inventory  []string
	locationID int
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

	db.Exec(
		"INSERT OR IGNORE INTO players (username, password, hp, locationID) VALUES (?, ?, ?, ?)",
		"testplayer", "hello", 100, 0,
	)

	db.Exec(
		"INSERT OR IGNORE INTO items (name, locationType, locationID) VALUES (?, ?, ?)",
		"Rusted Spoon", "room", 1,
	)

	mobone := Entity{10, "Green Slime", "Looks gooey and lifelike.", false, 10, 0}
	var mlistone []Entity
	mlistone = append(mlistone, mobone)
	roomone := Room{0, "Green Glade", "You look around to see tall standing trees towering over you...", [4]int{1, 0, 0, 0}, mlistone, []Item{}}
	roomtwo := Room{1, "Stone Pathway", "Looks like a long pathway. Wonder where it goes.", [4]int{1, 0, 1, 1}, []Entity{}, []Item{}}
	var rlist []Room
	rlist = append(rlist, roomone)
	rlist = append(rlist, roomtwo)
	world := World{[]Character{}, rlist, []*ConnectionData{}}

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
		if lType == "room" {
			item := Item{iID, iName, "", lType, lID}
			world.nodeList[lID].items = append(world.nodeList[lID].items, item)
		}
	}

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Listening successfuly on: " + CONN_HOST + ":" + CONN_PORT)

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
						world.characters = append(world.characters, Character{len(world.characters), connection.session.loginctx.hp, []string{""}, connection.session.loginctx.locationID})
						fmt.Println(world.characters)
						connection.session.character = &world.characters[len(world.characters)-1]
						stream.Write([]byte("Welcome to the MUD!\n\n"))
						HandleMovement(connection, world)
					} else {
						stream.Write([]byte("Wrong password!"))
					}
				} else {
					stream.Write([]byte("Welcome, " + connection.session.username + " to the game!\n"))
					world.characters = append(world.characters, Character{len(world.characters), connection.session.loginctx.hp, []string{""}, connection.session.loginctx.locationID})
					fmt.Println(world.characters)
					connection.session.character = &world.characters[len(world.characters)-1]
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
						if connection.session.character.locationID != id {
							stream.Write([]byte("  - " + dirs[dir] + ": " + world.nodeList[id].name + "\n"))
						} /* else {
							stream.Write([]byte("  - " + dirs[dir] + ": " + "none\n"))
						}*/
					}
					for _, ent := range world.nodeList[connection.session.character.locationID].entities {
						stream.Write([]byte(ent.name + ": " + ent.description))
					}
					for _, item := range world.nodeList[connection.session.character.locationID].items {
						stream.Write([]byte("1x - " + item.name + "\n"))
					}
				case "north":
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[0]
					HandleMovement(connection, world)
				case "south":
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[1]
					HandleMovement(connection, world)
				case "west":
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[2]
					HandleMovement(connection, world)
				case "east":
					connection.session.character.locationID = world.nodeList[connection.session.character.locationID].exits[3]
					HandleMovement(connection, world)

				case "inventory":
					for _, item := range world.nodeList[connection.session.character.locationID].items {
						if item.locationType == "player" && item.locationID == connection.session.id {
							stream.Write([]byte("  " + item.name + "\n"))
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
						if len(world.nodeList[connection.session.character.locationID].items) > i {
							stream.Write([]byte("You pick up a " + world.nodeList[connection.session.character.locationID].items[i].name))
							world.nodeList[connection.session.character.locationID].items[i].locationType = "player"
							world.nodeList[connection.session.character.locationID].items[i].locationID = connection.session.id
						} else {
							stream.Write([]byte("There is no such item."))
						}
					}
				default:
					stream.Write([]byte("Command not found!"))
				}
			}

			stream.Write([]byte("\n> "))
		}
	}
}

func HandleClientDisconnect(connection *ConnectionData, world *World, db *sql.DB) {
	connection.store.Write([]byte("Exiting game."))
	connection.store.Close()
	if connection.session.authorized {
		TotalPlayers -= 1
		fmt.Println(connection.session.character.hp)
		_, err := db.Exec("UPDATE players SET (hp, locationID) = (?, ?) WHERE id = ?", connection.session.character.hp, connection.session.character.locationID, connection.session.id)
		fmt.Println(err)
		world.characters[connection.session.character.worldID] = Character{connection.session.character.worldID, 100, []string{""}, 0}
		connection.session = &Session{false, 0, "", "", nil, nil}
	}
	TotalConnections -= 1
	fmt.Println(world.characters)
}

func HandleMovement(connection *ConnectionData, world *World) {
	r := world.nodeList[connection.session.character.locationID]
	connection.store.Write([]byte("\033[32m" + r.name + "\033[0m \n"))
	connection.store.Write([]byte(r.description))
}
