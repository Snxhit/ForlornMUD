package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"os"
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
type World struct {
	characters []Character
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
	newPlayer bool
	id        int
	username  string
	password  string
	hp        int
}

type Character struct {
	worldID   int
	hp        int
	inventory []string
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
			hp INT
		)
	`)

	db.Exec(
		"INSERT OR IGNORE INTO players(username, password, hp) VALUES (?, ?, ?)",
		"testplayer", "hello", 100,
	)

	world := World{[]Character{}}

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Listening successfuly on: " + CONN_HOST + ":" + CONN_PORT)

	for {
		newConn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		sesh := Session{false, 0, "", "", &LoginContext{false, 0, "", "", 100}, nil}
		conn := ConnectionData{TotalConnections, newConn, &sesh}
		TotalConnections += 1
		fmt.Println(conn)
		go HandleNewClient(&conn, &world, db)
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
				row := db.QueryRow("SELECT id, username, password, hp FROM players WHERE username = ?", cmdTokens[0])

				err := row.Scan(&connection.session.loginctx.id, &connection.session.loginctx.username, &connection.session.loginctx.password, &connection.session.loginctx.hp)
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
						world.characters = append(world.characters, Character{len(world.characters), connection.session.loginctx.hp, []string{""}})
						fmt.Println(world.characters)
						connection.session.character = &world.characters[len(world.characters)-1]
						stream.Write([]byte("Welcome to the MUD!"))
					} else {
						stream.Write([]byte("Wrong password!"))
					}
				} else {
					stream.Write([]byte("Welcome, " + connection.session.username + " to the game!\n"))
					world.characters = append(world.characters, Character{len(world.characters), connection.session.loginctx.hp, []string{""}})
					fmt.Println(world.characters)
					connection.session.character = &world.characters[len(world.characters)-1]
					prom, err := db.Exec("INSERT INTO players (username, password, hp) VALUES (?, ?, ?)", connection.session.username, cmdTokens[0], connection.session.character.hp)
					fmt.Println(err)
					id, _ := prom.LastInsertId()
					connection.session.password = cmdTokens[0]
					connection.session.id = int(id)
					connection.session.authorized = true
					TotalPlayers += 1
					stream.Write([]byte("Welcome to the MUD!"))
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
				default:
					stream.Write([]byte("Command not found!"))
				}

			case 2:
				switch cmdTokens[0] {
				case "echo":
					stream.Write([]byte(cmdTokens[1]))
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
		_, err := db.Exec("UPDATE players SET hp = ? WHERE id = ?", connection.session.character.hp, connection.session.id)
		fmt.Println(err)
		world.characters[connection.session.character.worldID] = Character{connection.session.character.worldID, 100, []string{""}}
		connection.session = &Session{false, 0, "", "", nil, nil}
	}
	TotalConnections -= 1
	fmt.Println(world.characters)
}
