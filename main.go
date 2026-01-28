// TODOs in order
// TODO: Equipment
// TODO: Varied damage due to equipment
// TODO: Colors

// BUGS

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
	characters    []*Character
	nodeList      []Room
	ItemTemplates map[int]*ItemTemplate
	items         map[int]*Item
	entities      map[int]*Entity
	merchants     map[int]*Merchant
	connections   []*ConnectionData
	mu            sync.Mutex
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

	db.Exec(`
		CREATE TABLE IF NOT EXISTS players (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			hp INT,

			str INT,
			dex INT,
			agi INT,
			stam INT,
			int INT,

			maxHp INT,
			coins INT,
			locationID INT
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS item_templates (
			id INTEGER PRIMARY KEY,
			name STRING NOT NULL,
			description STRING NOT NULL,
			itype STRING,
			baseValue INT
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS item_template_modifiers (
			sourceID INT,
			stat STRING,
			value INT
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			templateID INT,
			locationType STRING,
			locationID INT,
			equipped INT
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

	db.Exec(`
		CREATE TABLE IF NOT EXISTS merchants (
			entityID INT PRIMARY KEY,
			sellRate INT,
			buyRate INT
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS merchant_list (
			merchantID INT,
			templateID INT
		)
	`)

	db.Exec(
		"INSERT OR IGNORE INTO players (username, password, hp, str, dex, agi, stam, int, maxHp, coins, locationID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"testplayer", "hello", 100, 1, 1, 1, 1, 1, 100, 0, 0,
	)

	db.Exec(
		"INSERT OR IGNORE INTO items (templateID, locationType, locationID, equipped) VALUES (?, ?, ?, ?)",
		0, "room", 1, false,
	)

	/*	db.Exec(
		"INSERT OR IGNORE INTO item_modifiers (sourceID, stat, value) VALUES (?, ?, ?)",
		1, "str", 1,
	)*/

	db.Exec(
		"INSERT OR IGNORE INTO item_templates (id, name, description, itype, baseValue) VALUES (?, ?, ?, ?, ?)",
		0, "Rusted Spoon", "Looks rusted.", "mainhand", 11,
	)

	db.Exec(
		"INSERT OR IGNORE INTO item_template_modifiers (sourceID, stat, value) VALUES (?, ?, ?)",
		0, "str", 1,
	)

	db.Exec(
		"INSERT OR IGNORE INTO entities (name, desc, hp, locationID) VALUES (?, ?, ?, ?)",
		"Green Slime", "Looks goeey.", 100, 0,
	)

	prom, _ := db.Exec(
		"INSERT OR IGNORE INTO entities (name, desc, hp, locationID) VALUES (?, ?, ?, ?)",
		"Shayla, the Merchant", "Looks like she has stuff to sell.", 9999, 2,
	)
	x, _ := prom.LastInsertId()

	db.Exec(
		"INSERT OR IGNORE INTO merchants (entityID, sellRate, buyRate) VALUES (?, ?, ?)",
		int(x), 0.9, 1.1,
	)

	db.Exec(
		"INSERT OR IGNORE INTO merchant_list (merchantID, templateID) VALUES (?, ?)",
		int(x), 0,
	)

	roomone := Room{0, "Green Glade", "You look around to see tall standing trees towering over you...", [4]int{1, -1, -1, -1}, []int{}, []int{}}
	roomtwo := Room{1, "Stone Pathway", "Looks like a long pathway. Wonder where it goes.", [4]int{-1, 0, -1, 2}, []int{}, []int{}}
	roomthree := Room{2, "Dark Corner", "A very gloomy place...", [4]int{-1, -1, 1, -1}, []int{}, []int{}}

	var rlist []Room
	rlist = append(rlist, roomone)
	rlist = append(rlist, roomtwo)
	rlist = append(rlist, roomthree)
	world := World{[]*Character{}, rlist, map[int]*ItemTemplate{}, map[int]*Item{}, map[int]*Entity{}, map[int]*Merchant{}, []*ConnectionData{}, sync.Mutex{}}

	m_rows, err := db.Query("SELECT entityID, sellRate, buyRate FROM merchants")
	if err != nil {
		fmt.Println(err)
	}
	defer m_rows.Close()

	for m_rows.Next() {
		var eID int
		var sRate float64
		var bRate float64
		m_rows.Scan(&eID, &sRate, &bRate)

		merch := Merchant{eID, []int{}, sRate, bRate}
		world.merchants[eID] = &merch
	}

	ml_rows, err := db.Query("SELECT merchantID, templateID from merchant_list")
	if err != nil {
		fmt.Println(err)
	}
	defer ml_rows.Close()

	for ml_rows.Next() {
		var mID int
		var tID int
		ml_rows.Scan(&mID, &tID)

		world.merchants[mID].list = append(world.merchants[mID].list, tID)
	}

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

	t_rows, err := db.Query("SELECT id, name, description, itype, baseValue FROM item_templates")
	if err != nil {
		fmt.Println(err)
	}
	defer t_rows.Close()

	for t_rows.Next() {
		var tID int
		var tName string
		var desc string
		var iType string
		var baseValue int
		t_rows.Scan(&tID, &tName, &desc, &iType, &baseValue)
		world.ItemTemplates[tID] = &ItemTemplate{tID, tName, desc, iType, baseValue, []ItemModifier{}}
	}

	tm_rows, err := db.Query("SELECT sourceID, stat, value FROM item_template_modifiers")
	if err != nil {
		fmt.Println(err)
	}
	defer tm_rows.Close()

	for tm_rows.Next() {
		var sourceID int
		var stat string
		var value int
		tm_rows.Scan(&sourceID, &stat, &value)
		world.ItemTemplates[sourceID].modifiers = append(world.ItemTemplates[sourceID].modifiers, ItemModifier{stat, value})
	}

	rows, err := db.Query("SELECT id, templateID, locationType, locationID, equipped FROM items")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var iID int
		var tID int
		var lType string
		var lID int
		var eqI int
		var equipped bool
		rows.Scan(&iID, &tID, &lType, &lID, &eqI)
		if eqI == 0 {
			equipped = false
		} else {
			equipped = true
		}
		world.items[iID] = &Item{iID, tID, lType, lID, equipped}
		switch lType {
		case "room":
			itemID := world.items[iID].id
			world.nodeList[lID].itemIDs = append(world.nodeList[lID].itemIDs, itemID)
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

			sesh := Session{false, 0, "", "", &LoginContext{false, 0, "", "", 100, 0, 0}, nil}
			conn := ConnectionData{TotalConnections, newConn, &sesh}
			TotalConnections += 1
			world.connections = append(world.connections, &conn)
			fmt.Println(conn)
			go HandleNewClient(&conn, &world, db)
		}
	}()

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
				}
				fmt.Println(conn.session.character)
				if conn.session.character.inCombat {
					fmt.Println(conn.session.username)
					if conn.session.character.targetID == nil {
						conn.session.character.inCombat = false
						continue
					}
					if conn.session.character.targetType == &TargetEntity {
						pDam := (rand.Intn(7) + 3) * conn.session.character.getEffectiveStat("str")
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
							conn.store.Write([]byte("\nYou killed a " + world.entities[*conn.session.character.targetID].name + "!"))
							c := rand.Intn(100)
							conn.session.character.coins += c
							fmt.Println(world.entities)
							conn.store.Write([]byte("\nYou loot the " + world.entities[*conn.session.character.targetID].name + "'s body and find " + strconv.Itoa(c) + " coins!\n\n> "))
							db.Exec("DELETE FROM entities WHERE id = ?", *conn.session.character.targetID)
							for i, id := range world.nodeList[conn.session.character.locationID].entityIDs {
								if world.entities[id] == world.entities[*conn.session.character.targetID] {
									world.nodeList[conn.session.character.locationID].entityIDs = slices.Delete(world.nodeList[conn.session.character.locationID].entityIDs, i, i+1)
								}
							}
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
							p2Chr.conn.store.Write([]byte("\nYou died!"))
							if p2Chr.coins == 0 {
								conn.store.Write([]byte("\n" + p2Chr.conn.session.username + "didn't have any coins for you to loot!"))
							} else {
								c := rand.Intn(p2Chr.coins)
								conn.store.Write([]byte("\nYou loot " + p2Chr.conn.session.username + "'s body to steal " + strconv.Itoa(c) + " coins!"))
								p2Chr.conn.store.Write([]byte("\n" + conn.session.username + " steals " + strconv.Itoa(c) + " coins from you!"))
								p1Chr.coins += c
								p2Chr.coins -= c
							}
							p2Chr.conn.store.Write([]byte("\nYou are teleported to spawn!\n\n> "))
							conn.store.Write([]byte("\nYou killed " + p2Chr.conn.session.username + "!\n\n> "))
							p2Chr.hp = p2Chr.maxHp
							HandleMovement(p2Chr.conn, &world)
							fmt.Println("\n\n> ")
							continue
						} else if p1Chr.hp <= 0 {
							p2Chr.inCombat = false
							p2Chr.targetID = nil
							p2Chr.targetType = nil
							p1Chr.inCombat = false
							p1Chr.targetID = nil
							p1Chr.targetType = nil
							p2Chr.conn.store.Write([]byte("\nYou killed " + conn.session.username + "!\n\n> "))
							conn.store.Write([]byte("\nYou died!"))
							if p1Chr.coins == 0 {
								p1Chr.conn.store.Write([]byte("\n" + conn.session.username + "didn't have any coins for you to loot!"))
							} else {
								c := rand.Intn(p1Chr.coins)
								p2Chr.conn.store.Write([]byte("\nYou loot " + conn.session.username + "'s body to steal " + strconv.Itoa(c) + " coins!"))
								p1Chr.conn.store.Write([]byte("\n" + p2Chr.conn.session.username + " steals " + strconv.Itoa(c) + " coins from you!"))
								p2Chr.coins += c
								p1Chr.coins -= c
							}
							p1Chr.conn.store.Write([]byte("\nYou are teleported to spawn!\n\n> "))
							conn.store.Write([]byte("\nYou killed " + p1Chr.conn.session.username + "!\n\n> "))
							conn.store.Write([]byte("\nYou are teleported to spawn!"))
							p1Chr.hp = p1Chr.maxHp
							HandleMovement(conn, &world)
							fmt.Println("\n\n> ")
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
				row := db.QueryRow("SELECT id, username, password, hp, maxHp, locationID FROM players WHERE username = ?", cmdTokens[0])

				err := row.Scan(&connection.session.loginctx.id, &connection.session.loginctx.username, &connection.session.loginctx.password, &connection.session.loginctx.hp, &connection.session.loginctx.maxHp, &connection.session.loginctx.locationID)
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
					stream.Write([]byte("Welcome, " + connection.session.username + " to the game!\n"))
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
					case "incmhp":
						connection.session.character.maxHp += 10
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
							stream.Write([]byte("1x - " + world.ItemTemplates[world.items[itemID].templateID].name + "\n"))
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
							if world.items[item.id].locationType == "player" && world.items[item.id].locationID == connection.session.id && !world.items[item.id].equipped {
								stream.Write([]byte("  " + world.ItemTemplates[world.items[item.id].templateID].name + "\n"))
							}
						}
					case "equipped":
						s := []string{"mainhand", "offhand", "head", "body", "legs", "ring"}
						sSpace := []string{"mainhand", "offhand ", "head    ", "body    ", "legs    ", "ring    "}
						stream.Write([]byte("\n"))
						for i, slot := range s {
							if connection.session.character.equipment[slot] != 0 {
								stream.Write([]byte("  - " + sSpace[i] + " - " + world.ItemTemplates[world.items[connection.session.character.equipment[slot]].templateID].name + "\n"))
							} else {
								stream.Write([]byte("  + " + sSpace[i] + " - " + "<empty>" + "\n"))
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
						i, err := strconv.Atoi(cmdTokens[1])
						if err != nil {
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
								stream.Write([]byte("You equip a " + world.ItemTemplates[world.items[playerItems[i]].templateID].name + " on the " + world.ItemTemplates[world.items[playerItems[i]].templateID].itype + "\n\n> "))
								world.items[playerItems[i]].equipped = true
								_, err := db.Exec("UPDATE items SET (equipped) = (?) WHERE id = ?", true, playerItems[i])
								fmt.Println(err)
								fmt.Println(connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype])
								connection.session.character.equipment[world.ItemTemplates[world.items[playerItems[i]].templateID].itype] = playerItems[i]
							} else {
								stream.Write([]byte("There is no such item."))
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
					case "list":
						for en, e := range world.entities {
							keywords := strings.FieldsFunc(strings.ToLower(e.name), func(r rune) bool {
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
						i, err := strconv.Atoi(cmdTokens[1])
						fmt.Println(err)
						if err == nil {
							for en, e := range world.entities {
								keywords := strings.FieldsFunc(strings.ToLower(e.name), func(r rune) bool {
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
										stream.Write([]byte("\nYou buy 1x " + world.ItemTemplates[item].name + " for " + bpS + " coins from " + e.name + "\n"))
										connection.session.character.coins -= int(bp)
									} else {
										stream.Write([]byte("\nYou don't have enough coins to buy this item!\n"))
									}
									break
								}
							}
						}
					case "sell":
						i, err := strconv.Atoi(cmdTokens[1])
						fmt.Println(err)
						if err == nil {
							for en, e := range world.entities {
								keywords := strings.FieldsFunc(strings.ToLower(e.name), func(r rune) bool {
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
											stream.Write([]byte("\nYou sell 1x " + world.ItemTemplates[item].name + " for " + spS + " coins from " + e.name + "\n"))
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
			}
			stream.Write([]byte("\n> "))
		}
	}
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
	connection.store.Write([]byte("\033[32m" + r.name + "\033[0m \n"))
	connection.store.Write([]byte(r.description))
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
