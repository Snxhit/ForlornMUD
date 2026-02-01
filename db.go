package main

import (
	"database/sql"
	"fmt"
)

func dbInit(db *sql.DB) {
	db.Exec(`
		CREATE TABLE IF NOT EXISTS rooms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING,
			description STRING,
			n INT, s INT, w INT, e INT
		)
	`)

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
			baseDam INT,
			baseDef INT,
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
		CREATE TABLE IF NOT EXISTS entity_templates (
			id INTEGER PRIMARY KEY,
			name STRING NOT NULL,
			description STRING NOT NULL,
			
			str INT,
			dex INT,
			agi INT,
			stam INT,
			int INT,

			aggro INT,
			maxHp INT,
			baseDam INT,
			baseDef INT,
			cMin INT,
			cMax INT
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS entity_template_drops (
			entityTemplateID INT,
			itemTemplateID INT,
			chance INT,
			min INT,
			max INT
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			templateID INT,
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

	var exists int
	err := db.QueryRow("SELECT * FROM rooms WHERE id = ?", 1).Scan(&exists)
	if err == sql.ErrNoRows {
		db.Exec(
			"INSERT OR IGNORE INTO rooms (id, name, description, n, s, w, e) VALUES (?, ?, ?, ?, ?, ?, ?)",
			0, "Green Glade", "You look around to see tall standing trees towering over you...", 1, -1, -1, -1,
		)
		db.Exec(
			"INSERT OR IGNORE INTO rooms (name, description, n, s, w, e) VALUES (?, ?, ?, ?, ?, ?)",
			"Stone Pathway", "Looks like a long pathway. Wonder where it goes.", -1, 0, -1, 2,
		)
		db.Exec(
			"INSERT OR IGNORE INTO rooms (name, description, n, s, w, e) VALUES (?, ?, ?, ?, ?, ?)",
			"Dark Corner", "A very gloomy place...", -1, -1, 1, 3,
		)
		db.Exec(
			"INSERT OR IGNORE INTO rooms (name, description, n, s, w, e) VALUES (?, ?, ?, ?, ?, ?)",
			"A room of Thrones", "Filled with the smell of glory.", -1, -1, 2, -1,
		)

		db.Exec(
			"INSERT OR IGNORE INTO item_template_modifiers (sourceID, stat, value) VALUES (?, ?, ?)",
			0, "str", 1,
		)

		db.Exec(
			"INSERT OR IGNORE INTO entity_template_drops (entityTemplateID, itemTemplateID, chance, min, max) VALUES (?, ?, ?, ?, ?)",
			0, 0, 50, 1, 1,
		)

		prom, _ := db.Exec(
			"INSERT OR IGNORE INTO entities (templateID, hp, locationID) VALUES (?, ?, ?)",
			1, 0, 2,
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
	}

	db.Exec(
		"INSERT OR IGNORE INTO items (templateID, locationType, locationID, equipped) VALUES (?, ?, ?, ?)",
		0, "room", 1, false,
	)
	db.Exec(
		"INSERT OR IGNORE INTO items (templateID, locationType, locationID, equipped) VALUES (?, ?, ?, ?)",
		1, "room", 3, false,
	)

	db.Exec(
		"INSERT OR IGNORE INTO item_templates (id, name, description, itype, baseDam, baseDef, baseValue) VALUES (?, ?, ?, ?, ?, ?, ?)",
		0, "Rusted Spoon", "Looks rusted.", "mainhand", 10, 0, 11,
	)
	db.Exec(
		"INSERT OR IGNORE INTO item_templates (id, name, description, itype, baseDam, baseDef, baseValue) VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, "Shadow Helmet", "Black.", "head", 10, 20, 11,
	)

	db.Exec(
		"INSERT OR IGNORE INTO entity_templates (id, name, description, str, dex, agi, stam, int, aggro, maxHp, baseDam, baseDef, cMin, cMax) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		0, "Green Slime", "Looks jiggly.", 1, 1, 1, 1, 1, 0, 50, 3, 3, 70, 100,
	)
	db.Exec(
		"INSERT OR IGNORE INTO entity_templates (id, name, description, str, dex, agi, stam, int, aggro, maxHp, baseDam, baseDef, cMin, cMax) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, "Shayla, the Merchant", "Looks like she has something to sell!", 10, 10, 10, 10, 10, 0, 10000, 300, 300, 18000, 21000,
	)
	db.Exec(
		"INSERT OR IGNORE INTO entity_templates (id, name, description, str, dex, agi, stam, int, aggro, maxHp, baseDam, baseDef, cMin, cMax) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		2, "King's Last Guard", "Looks like it has a sad backstory.", 20, 20, 20, 20, 20, 0, 200, 30, 30, 1800, 2100,
	)

	db.Exec(
		"INSERT OR IGNORE INTO entities (templateID, hp, locationID) VALUES (?, ?, ?)",
		0, 0, 0,
	)
	db.Exec(
		"INSERT OR IGNORE INTO entities (templateID, hp, locationID) VALUES (?, ?, ?)",
		2, 0, 3,
	)
}

func objectsInit(db *sql.DB, world *World) {
	r_rows, err := db.Query("SELECT id, name, description, n, s, w, e FROM rooms")
	if err != nil {
		fmt.Println(err)
	}
	defer r_rows.Close()

	for r_rows.Next() {
		var rID int
		var name string
		var desc string
		var n, s, w, e int
		r_rows.Scan(&rID, &name, &desc, &n, &s, &w, &e)

		room := Room{rID, name, desc, [4]int{n, s, w, e}, []int{}, []int{}}
		world.nodeList = append(world.nodeList, room)
	}

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

	ml_rows, err := db.Query("SELECT merchantID, templateID FROM merchant_list")
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

	//"INSERT OR IGNORE INTO entity_templates (id, name, description, str, dex, agi, stam, int, aggro, maxHp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	et_rows, err := db.Query("SELECT id, name, description, str, dex, agi, stam, int, aggro, maxHp, baseDam, baseDef, cMin, cMax FROM entity_templates")
	if err != nil {
		fmt.Println(err)
	}
	defer et_rows.Close()

	for et_rows.Next() {
		var id int
		var name string
		var desc string
		var stats Stats
		var agI int
		var aggro bool
		var maxHp int
		var baseDam int
		var baseDef int
		var cMin int
		var cMax int
		et_rows.Scan(&id, &name, &desc, &stats.Str, &stats.Dex, &stats.Agi, &stats.Stam, &stats.Int, &agI, &maxHp, &baseDam, &baseDef, &cMin, &cMax)

		if agI == 0 {
			aggro = false
		} else {
			aggro = true
		}

		template := EntityTemplate{id, name, desc, stats, aggro, maxHp, baseDam, baseDef, cMin, cMax, []DropEntry{}}
		world.EntityTemplates[id] = &template
	}

	entd_rows, err := db.Query("SELECT entityTemplateID, itemTemplateID, chance, min, max FROM entity_template_drops")
	if err != nil {
		fmt.Println(err)
	}
	defer entd_rows.Close()

	for entd_rows.Next() {
		var eTID int
		var iTID int
		var chance int
		var minQty int
		var maxQty int
		entd_rows.Scan(&eTID, &iTID, &chance, &minQty, &maxQty)

		drop := DropEntry{eTID, iTID, chance, minQty, maxQty}
		world.EntityTemplates[eTID].dropTable = append(world.EntityTemplates[eTID].dropTable, drop)
	}

	ent_rows, err := db.Query("SELECT id, templateID, hp, locationID FROM entities")
	if err != nil {
		fmt.Println(err)
	}
	defer ent_rows.Close()

	for ent_rows.Next() {
		var eID int
		var tID int
		var eHp int
		var lID int
		ent_rows.Scan(&eID, &tID, &eHp, &lID)
		for _, b := range world.EntityTemplates {
			// fmt.Println(a)
			fmt.Println(b)
		}
		eHp = world.EntityTemplates[tID].maxHp

		ent := Entity{eID, tID, nil, false, eHp, lID}
		world.entities[eID] = &ent
		world.nodeList[lID].entityIDs = append(world.nodeList[lID].entityIDs, ent.id)
	}

	t_rows, err := db.Query("SELECT id, name, description, itype, baseDam, baseDef, baseValue FROM item_templates")
	if err != nil {
		fmt.Println(err)
	}
	defer t_rows.Close()

	for t_rows.Next() {
		var tID int
		var tName string
		var desc string
		var iType string
		var baseDam int
		var baseDef int
		var baseValue int
		t_rows.Scan(&tID, &tName, &desc, &iType, &baseDam, &baseDef, &baseValue)
		world.ItemTemplates[tID] = &ItemTemplate{tID, tName, desc, iType, baseDam, baseDef, baseValue, []ItemModifier{}}
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
}
