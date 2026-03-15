package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func dbInit(db *sql.DB) {
	err := loadSQL(db, "./template")
	if err != nil {
		fmt.Println(err)
	}
}

func loadSQL(db *sql.DB, dir string) error {
	var name string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name="rooms"`).Scan(&name)
	if err != sql.ErrNoRows {
		fmt.Println("DB already exists!")
		return nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, f.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return err
		}

		fmt.Println("DB file loaded!")
	}

	return nil
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
		world.nodeList[rID] = &room
	}

	s_rows, err := db.Query("SELECT id, locationID, templateType, templateID, duration, maxSpawns FROM spawners")
	if err != nil {
		fmt.Println(err)
	}
	defer s_rows.Close()

	for s_rows.Next() {
		var id int
		var lID int
		var tType string
		var tID int
		var dur int
		var maxSpawns int
		s_rows.Scan(&id, &lID, &tType, &tID, &dur, &maxSpawns)

		spawner := Spawner{id, lID, tType, tID, dur, maxSpawns, int(world.tick) + dur}
		world.spawners = append(world.spawners, spawner)
	}

	c_rows, err := db.Query("SELECT id, name, tag, owner_id, status, created_at FROM clans")
	if err != nil {
		fmt.Println(err)
	}
	defer c_rows.Close()

	for c_rows.Next() {
		var id int
		var name string
		var tag string
		var ownerID int
		var status string
		var createdAt time.Time
		c_rows.Scan(&id, &name, &tag, &ownerID, &status, &createdAt)

		clan := Clan{id, name, tag, ownerID, createdAt, status, []int{}}
		world.clans[id] = &clan
	}

	p_rows, err := db.Query("SELECT id, username, clan_id FROM players")
	if err != nil {
		fmt.Println(err)
	}
	defer p_rows.Close()

	for p_rows.Next() {
		var id int
		var username string
		var clanID sql.NullInt64
		p_rows.Scan(&id, &username, &clanID)

		if clanID.Valid {
			if world.clans[int(clanID.Int64)].ownerID != id {
				world.clans[int(clanID.Int64)].members = append(world.clans[int(clanID.Int64)].members, id)
			}
		}
		world.playerList[id] = &PlayerData{id, username, false}
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

	et_rows, err := db.Query("SELECT id, name, description, str, dex, agi, stam, int, level, aggro, maxHp, baseDam, baseDef, baseExp, cMin, cMax FROM entity_templates")
	if err != nil {
		fmt.Println(err)
	}
	defer et_rows.Close()

	for et_rows.Next() {
		var id int
		var name string
		var desc string
		var stats Stats
		var level int
		var agI int
		var aggro bool
		var maxHp int
		var baseDam int
		var baseDef int
		var baseExp int
		var cMin int
		var cMax int
		et_rows.Scan(&id, &name, &desc, &stats.Str, &stats.Dex, &stats.Agi, &stats.Stam, &stats.Int, &level, &agI, &maxHp, &baseDam, &baseDef, &baseExp, &cMin, &cMax)

		if agI == 0 {
			aggro = false
		} else {
			aggro = true
		}

		template := EntityTemplate{id, name, desc, stats, level, aggro, maxHp, baseDam, baseDef, baseExp, cMin, cMax, []DropEntry{}}
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
		world.ItemTemplates[tID] = &ItemTemplate{tID, tName, desc, iType, baseDam, baseDef, baseValue, []ItemModifier{}, []ItemEffect{}}
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

	te_rows, err := db.Query("SELECT sourceID, effect, value FROM item_template_effects")
	if err != nil {
		fmt.Println(err)
	}
	defer te_rows.Close()

	for te_rows.Next() {
		var sourceID int
		var effect string
		var value int
		te_rows.Scan(&sourceID, &effect, &value)
		world.ItemTemplates[sourceID].effects = append(world.ItemTemplates[sourceID].effects, ItemEffect{effect, value})
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
