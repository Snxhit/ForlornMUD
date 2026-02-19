package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	nodeList        map[int]*Room
	spawners        []Spawner
	ItemTemplates   map[int]*ItemTemplate
	EntityTemplates map[int]*EntityTemplate
	items           map[int]*Item
	entities        map[int]*Entity
	merchants       map[int]*Merchant
	connections     []*ConnectionData
	tick            int64
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

type ItemEffect struct {
	effect string
	value  int
}

type Room struct {
	id          int
	name        string
	description string
	exits       [4]int
	entityIDs   []int
	itemIDs     []int
}

type Spawner struct {
	id            int
	locationID    int
	templateType  string
	templateID    int
	duration      int
	maxSpawns     int
	nextSpawnTick int
}

type EntityTemplate struct {
	id          int
	name        string
	description string
	stats       Stats
	level       int
	aggro       bool
	maxHp       int
	baseDam     int
	baseDef     int
	baseExp     int
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
	effects     []ItemEffect
}

type Item struct {
	id           int
	templateID   int
	locationType string
	locationID   int
	equipped     bool
}

type ConnectionData struct {
	ConnID          int
	store           net.Conn
	session         *Session
	isClientWeb     bool
	isPrettyEnabled bool
	isColorEnabled  bool
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
	baseStats  Stats
	exp        int
	level      int
	trains     int
	coins      int
}

type Character struct {
	worldID    int
	hp         int
	maxHp      int
	baseStats  Stats
	exp        int
	level      int
	trains     int
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

	world := World{[]*Character{}, map[int]*Room{}, []Spawner{}, map[int]*ItemTemplate{}, map[int]*EntityTemplate{}, map[int]*Item{}, map[int]*Entity{}, map[int]*Merchant{}, []*ConnectionData{}, 0, sync.Mutex{}}

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

			sesh := Session{false, 0, "", "", &LoginContext{false, 0, "", "", 100, 0, 0, Stats{1, 1, 1, 1, 1}, 100, 1, 0, 0}, nil}
			conn := ConnectionData{TotalConnections, newConn, &sesh, false, true, true}

			reader := bufio.NewReader(newConn)

			newConn.SetReadDeadline(time.Now().Add(2000 * time.Millisecond))
			firstLine, err := reader.ReadString('\n')
			newConn.SetReadDeadline(time.Time{})

			if err == nil && firstLine == "CLIENT WEB\n" {
				conn.isClientWeb = true
			}

			asciiGreeting(&conn)

			TotalConnections += 1
			world.connections = append(world.connections, &conn)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Client crashed: ", r)
						newConn.Close()
					}
				}()
				HandleNewClient(&conn, &world, db)
			}()
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
		var strCmd string
		if connection.session.authorized {
			strCmd = strings.ToLower(string(bufCmd[0 : len(bufCmd)-1]))
		} else {
			strCmd = string(bufCmd[0 : len(bufCmd)-1])
		}
		cmdTokens := strings.Split(strCmd, " ")
		if len(cmdTokens) == 1 && cmdTokens[0] == "exit" || cmdTokens[0] == "quit" {
			HandleClientDisconnect(connection, world, db)
			return
		}

		if !connection.session.authorized {
			if len(cmdTokens) != 1 {
				stream.Write([]byte("\n  No inputs can have spaces!\n"))
			} else if len(cmdTokens) == 1 && len(connection.session.username) == 0 && len(cmdTokens[0]) > 3 && len(cmdTokens[0]) <= 14 {

				row := db.QueryRow("SELECT id, username, password, hp, str, dex, agi, stam, int, exp, level, trains, maxHp, coins, locationID FROM players WHERE username = ?", cmdTokens[0])

				l := connection.session.loginctx
				var str int
				var dex int
				var agi int
				var stam int
				var int int
				err := row.Scan(&l.id, &l.username, &l.password, &l.hp, &str, &dex, &agi, &stam, &int, &l.exp, &l.level, &l.trains, &l.maxHp, &l.coins, &l.locationID)
				connection.session.loginctx.baseStats = Stats{str, dex, agi, stam, int}
				if err != nil {
					if err == sql.ErrNoRows {
						stream.Write([]byte("\n  User " + cmdTokens[0] + " not found, creating user.\n"))
						stream.Write([]byte("  Please enter a new password.\n"))
						connection.session.username = cmdTokens[0]
						connection.session.loginctx.newPlayer = true
					} else {
						fmt.Println(err)
					}
				} else {
					connection.session.loginctx.newPlayer = false
					connection.session.username = cmdTokens[0]
					stream.Write([]byte("\n  User " + cmdTokens[0] + " found, what is your password?\n"))
				}
			} else if len(cmdTokens[0]) <= 3 {
				stream.Write([]byte("\n  Too short!\n"))
			} else if len(cmdTokens[0]) > 15 {
				stream.Write([]byte("\n  Too long!\n"))
			} else if len(cmdTokens) == 1 && len(connection.session.username) != 0 && len(cmdTokens[0]) >= 5 && len(cmdTokens[0]) <= 50 {
				if !connection.session.loginctx.newPlayer {
					err := bcrypt.CompareHashAndPassword([]byte(connection.session.loginctx.password), []byte(cmdTokens[0]))
					if err == nil {
						stream.Write([]byte("\n  Correct password! Logged in as " + color(connection, "cyan", "tp") + connection.session.username + color(connection, "reset", "reset") + "!\n"))
						connection.session.password = connection.session.loginctx.password
						connection.session.id = connection.session.loginctx.id
						connection.session.authorized = true
						TotalPlayers += 1
						world.mu.Lock()
						l := connection.session.loginctx
						world.characters = append(world.characters, &Character{len(world.characters), l.hp, l.maxHp, l.baseStats, l.exp, l.level, l.trains, l.coins, map[string]int{}, []StatModifier{}, l.locationID, nil, nil, false, connection})
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
						stream.Write([]byte("  Welcome to the MUD!\n\n"))
						connection.store.Write([]byte("\n\x01EXP " + "exp:" + strconv.Itoa(connection.session.character.exp) + " lvl:" + strconv.Itoa(connection.session.character.level) + " trains:" + strconv.Itoa(connection.session.character.trains) + "\n"))
						connection.store.Write([]byte("\n\x01SELF coins:" + strconv.Itoa(connection.session.character.coins) + "\n"))
						HandleMovement(connection, world)
					} else {
						stream.Write([]byte("\n  Wrong password!\n"))
					}
				} else {
					stream.Write([]byte("\n  Welcome, " + color(connection, "cyan", "tp") + connection.session.username + color(connection, "reset", "reset") + " to the game!\n"))
					world.mu.Lock()
					world.characters = append(world.characters, &Character{len(world.characters), connection.session.loginctx.hp, 200, Stats{10, 10, 10, 10, 10}, 100, 1, 0, 0, map[string]int{}, []StatModifier{}, connection.session.loginctx.locationID, nil, nil, false, connection})
					connection.session.character = world.characters[len(world.characters)-1]
					hashedPass, err := bcrypt.GenerateFromPassword([]byte(cmdTokens[0]), bcrypt.DefaultCost)
					world.mu.Unlock()
					prom, err := db.Exec("INSERT INTO players (username, password, hp, str, dex, agi, stam, int, exp, level, trains, maxHp, coins, locationID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", connection.session.username, string(hashedPass), connection.session.character.hp, 10, 10, 10, 10, 10, 100, 1, 0, 200, 0, connection.session.character.locationID)
					if err != nil {
						fmt.Println(err)
					}
					id, err := prom.LastInsertId()
					if err != nil {
						fmt.Println(err)
					}
					connection.session.password = string(hashedPass)
					connection.session.id = int(id)
					connection.session.authorized = true
					TotalPlayers += 1
					stream.Write([]byte("  Welcome to the MUD!\n\n"))
					connection.store.Write([]byte("\n\x01EXP " + "exp:" + strconv.Itoa(connection.session.character.exp) + " lvl:" + strconv.Itoa(connection.session.character.level) + " trains:" + strconv.Itoa(connection.session.character.trains) + "\n"))
					connection.store.Write([]byte("\n\x01SELF coins:" + strconv.Itoa(connection.session.character.coins) + "\n"))
					HandleMovement(connection, world)
					stream.Write([]byte("\x1b[2K\r\n  " + color(connection, "red", "tp") + "!!!" + color(connection, "reset", "reset") + " Please enter the command " + color(connection, "red", "tp") + "help newplayer" + color(connection, "reset", "reset") + " to get started." + color(connection, "red", "tp") + " !!! " + color(connection, "reset", "reset") + "\n "))
				}
			} else if len(cmdTokens[0]) < 5 {
				stream.Write([]byte("\n  Too short!\n"))
			} else if len(cmdTokens[0]) > 50 {
				stream.Write([]byte("\n  Too long!\n"))
			}

			stream.Write([]byte("\n> "))
		} else {
			if !connection.session.character.inCombat {
				x := Commands(cmdTokens, db, world, connection)
				if x == 0 {
					return
				}
			} else {
				CommandsCombat(cmdTokens, db, world, connection)
			}
			stream.Write([]byte("\n> "))
		}
	}
}

func HandleClientDisconnect(connection *ConnectionData, world *World, db *sql.DB) {
	world.mu.Lock()
	defer world.mu.Unlock()

	if connection.isClientWeb {
		connection.store.Write([]byte("\n\x01COMBAT type:entity hp:0 maxHp:0 enemyName:None enemyHp:0 enemyMaxHp:0\n"))
		connection.store.Write([]byte("\n\x01EXP exp:100 lvl:1 trains:0\n"))
		connection.store.Write([]byte("\n\x01SELF coins:0\n"))
	}

	connection.store.Write([]byte("Exiting game."))
	if connection.session.authorized {
		if connection.session.character.inCombat {
			if connection.session.character.targetType == &TargetEntity {
				connection.store.Write([]byte("\nYou died!\n\n> "))
				world.entities[*connection.session.character.targetID].inCombat = false
				world.entities[*connection.session.character.targetID].targetID = nil
				connection.session.character.inCombat = false
				connection.session.character.targetID = nil
				connection.session.character.targetType = nil
			} else if connection.session.character.targetType == &TargetPlayer {
				p1Chr := connection.session.character
				p2Chr := world.characters[*connection.session.character.targetID]
				p2Chr.inCombat = false
				p2Chr.targetID = nil
				p2Chr.targetType = nil
				p1Chr.inCombat = false
				p1Chr.targetID = nil
				p1Chr.targetType = nil
				p2Chr.conn.store.Write([]byte("\x1b[2K\r  " + color(p2Chr.conn, "cyan", "tp") + connection.session.username + color(p2Chr.conn, "green", "tp") + " logged" + color(p2Chr.conn, "reset", "reset") + " out, you win!\n\n> "))
				connection.store.Write([]byte("\nYou died!"))
				if p1Chr.coins == 0 {
					p1Chr.conn.store.Write([]byte("\n" + connection.session.username + "didn't have any coins for you to loot!"))
				} else {
					c := rand.Intn(p1Chr.coins)
					p2Chr.conn.store.Write([]byte("\nYou loot " + connection.session.username + "'s body to steal " + strconv.Itoa(c) + " coins!"))
					p1Chr.conn.store.Write([]byte("\n" + p2Chr.conn.session.username + " steals " + strconv.Itoa(c) + " coins from you!"))
					p2Chr.coins += c
					p1Chr.coins -= c
				}
				p1Chr.conn.store.Write([]byte("\nYou are teleported to spawn!\n\n> "))
				connection.store.Write([]byte("\nYou killed " + p1Chr.conn.session.username + "!\n\n> "))
				connection.store.Write([]byte("\nYou are teleported to spawn!"))
				p1Chr.hp = p1Chr.maxHp
				HandleMovement(connection, world)
			}
		}
		connection.store.Close()
		TotalPlayers -= 1
		s := connection.session.character.baseStats
		c := connection.session.character
		_, err := db.Exec("UPDATE players SET (hp, str, dex, agi, stam, int, exp, level, trains, maxHp, coins, locationID) = (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) WHERE id = ?", c.hp, s.Str, s.Dex, s.Agi, s.Stam, s.Int, c.exp, c.level, c.trains, c.maxHp, c.coins, c.locationID, connection.session.id)
		fmt.Println(err)
		world.characters[connection.session.character.worldID] = &Character{connection.session.character.worldID, 100, 0, Stats{1, 1, 1, 1, 1}, 0, 1, 0, 0, map[string]int{}, []StatModifier{}, 0, nil, nil, false, nil}

		for i, c := range world.connections {
			if c == connection {
				world.connections = append(world.connections[:i], world.connections[i+1:]...)
				break
			}
		}

		connection.session = nil
	} else {
		for i, c := range world.connections {
			if c == connection {
				world.connections = append(world.connections[:i], world.connections[i+1:]...)
				break
			}
		}

		connection.session = nil
	}
	TotalConnections -= 1
}
