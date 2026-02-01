// TODOs in order
// TODO: Colors

// BUGS

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

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
	characters      []*Character
	nodeList        []Room
	ItemTemplates   map[int]*ItemTemplate
	EntityTemplates map[int]*EntityTemplate
	items           map[int]*Item
	entities        map[int]*Entity
	merchants       map[int]*Merchant
	connections     []*ConnectionData
	mu              sync.Mutex
}

type Stats struct {
	Str  int // damage
	Dex  int // chance to hit ++ chance to hit more than once per tick
	Agi  int // chance to dodge
	Stam int // health
	Int  int // mana required for spells
}

type StatModifier struct {
	sourceType string // for now only item
	sourceID   int
	stat       string
	value      int
}

type ItemModifier struct {
	stat  string
	value int
}

type Room struct {
	id          int
	name        string
	description string
	exits       [4]int
	entityIDs   []int
	itemIDs     []int
}

type EntityTemplate struct {
	id          int
	name        string
	description string
	stats       Stats
	aggro       bool
	maxHp       int
	baseDam     int
	baseDef     int
	cMin        int
	cMax        int
	dropTable   []DropEntry
}

type DropEntry struct {
	entityTemplateID int
	itemTemplateID   int
	chance           int
	min              int
	max              int // both inclusive
}

type Entity struct {
	id         int
	templateID int
	targetID   *int
	inCombat   bool
	hp         int
	locationID int
}

type Merchant struct {
	entityID int
	list     []int
	sellRate float64
	buyRate  float64
}

type ItemTemplate struct {
	id          int
	name        string
	description string
	itype       string
	baseDam     int
	baseDef     int
	baseValue   int
	modifiers   []ItemModifier
}

type Item struct {
	id           int
	templateID   int
	locationType string
	locationID   int
	equipped     bool
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
	maxHp      int
	locationID int
}

type Character struct {
	worldID    int
	hp         int
	maxHp      int
	baseStats  Stats
	coins      int
	equipment  map[string]int
	modifiers  []StatModifier
	locationID int
	targetType *string
	targetID   *int
	inCombat   bool

	conn *ConnectionData
}

func (char Character) getEffectiveStat(stat string) int {
	switch stat {
	case "str":
		mStat := 0
		for _, mod := range char.modifiers {
			if mod.stat == "str" {
				mStat += mod.value
			}
		}
		return char.baseStats.Str + mStat
	default:
		return 0
	}
}

func main() {
	db, err := sql.Open("sqlite", "game.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dbInit(db)

	world := World{[]*Character{}, []Room{}, map[int]*ItemTemplate{}, map[int]*EntityTemplate{}, map[int]*Item{}, map[int]*Entity{}, map[int]*Merchant{}, []*ConnectionData{}, sync.Mutex{}}

	objectsInit(db, &world)

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

			sesh := Session{false, 0, "", "", &LoginContext{false, 0, "", "", 100, 0, 0}, nil}
			conn := ConnectionData{TotalConnections, newConn, &sesh}
			TotalConnections += 1
			world.connections = append(world.connections, &conn)
			fmt.Println(conn)
			go HandleNewClient(&conn, &world, db)
		}
	}()

	ticks(&world, db)
}

func HandleNewClient(connection *ConnectionData, world *World, db *sql.DB) {
	stream := connection.store
	stream.Write([]byte("\nWhat is your name?\n> "))
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
				row := db.QueryRow("SELECT id, username, password, hp, maxHp, locationID FROM players WHERE username = ?", cmdTokens[0])

				err := row.Scan(&connection.session.loginctx.id, &connection.session.loginctx.username, &connection.session.loginctx.password, &connection.session.loginctx.hp, &connection.session.loginctx.maxHp, &connection.session.loginctx.locationID)
				if err != nil {
					stream.Write([]byte("\nUser " + cmdTokens[0] + " not found, creating user.\n"))
					stream.Write([]byte("Please enter a new password."))
					connection.session.username = cmdTokens[0]
					connection.session.loginctx.newPlayer = true
				} else {
					connection.session.loginctx.newPlayer = false
					connection.session.username = cmdTokens[0]
					stream.Write([]byte("\nUser " + cmdTokens[0] + " found, what is your password?"))
				}
			} else if len(cmdTokens) == 1 && len(connection.session.username) != 0 {
				fmt.Println(connection.session.loginctx.password)
				fmt.Println(connection.session.loginctx.newPlayer)
				if !connection.session.loginctx.newPlayer {
					if cmdTokens[0] == connection.session.loginctx.password {
						stream.Write([]byte("\nCorrect password! Welcome, " + connection.session.username + "!\n"))
						connection.session.password = connection.session.loginctx.password
						connection.session.id = connection.session.loginctx.id
						connection.session.authorized = true
						TotalPlayers += 1
						world.mu.Lock()
						world.characters = append(world.characters, &Character{len(world.characters), connection.session.loginctx.hp, connection.session.loginctx.maxHp, Stats{1, 1, 1, 1, 1}, 0, map[string]int{}, []StatModifier{}, connection.session.loginctx.locationID, nil, nil, false, connection})
						fmt.Println(world.characters)
						world.mu.Unlock()
						connection.session.character = world.characters[len(world.characters)-1]
						for _, i := range world.items {
							if i.locationType == TargetPlayer && i.locationID == connection.session.id && i.equipped {
								connection.session.character.equipment[world.ItemTemplates[i.templateID].itype] = i.id
								for _, mod := range world.ItemTemplates[i.templateID].modifiers {
									connection.session.character.modifiers = append(connection.session.character.modifiers, StatModifier{"item", i.id, mod.stat, mod.value})
								}
							}
						}
						stream.Write([]byte("Welcome to the MUD!\n\n"))
						HandleMovement(connection, world)
					} else {
						stream.Write([]byte("Wrong password!"))
					}
				} else {
					stream.Write([]byte("\nWelcome, " + connection.session.username + " to the game!\n"))
					world.mu.Lock()
					world.characters = append(world.characters, &Character{len(world.characters), connection.session.loginctx.hp, 200, Stats{1, 1, 1, 1, 1}, 0, map[string]int{}, []StatModifier{}, connection.session.loginctx.locationID, nil, nil, false, connection})
					fmt.Println(world.characters)
					connection.session.character = world.characters[len(world.characters)-1]
					world.mu.Unlock()
					prom, err := db.Exec("INSERT INTO players (username, password, hp, maxHp, locationID) VALUES (?, ?, ?, ?, ?)", connection.session.username, cmdTokens[0], connection.session.character.hp, connection.session.character.maxHp, connection.session.character.locationID)
					fmt.Println(err)
					id, _ := prom.LastInsertId()
					connection.session.password = cmdTokens[0]
					connection.session.id = int(id)
					connection.session.authorized = true
					TotalPlayers += 1
					stream.Write([]byte("Welcome to the MUD!\n\n"))
					HandleMovement(connection, world)
				}
			}

			stream.Write([]byte("\n> "))
		} else {
			if !connection.session.character.inCombat {
				x := Commands(cmdTokens, db, world, connection)
				if x == 0 {
					return
				}
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
		world.characters[connection.session.character.worldID] = &Character{connection.session.character.worldID, 100, 0, Stats{1, 1, 1, 1, 1}, 0, map[string]int{}, []StatModifier{}, 0, nil, nil, false, nil}

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
