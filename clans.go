package main

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"
)

func CreateClan(connection *ConnectionData, world *World, db *sql.DB, name string) {
	if connection.session.character.clan != nil {
		connection.store.Write([]byte("\n  You are in already in a clan!\n  Leave your clan to create one!\n"))
		return
	}
	for _, c := range world.clans {
		if c.name == name {
			connection.store.Write([]byte(color(connection, "red", "tp") + "\n  A clan by this name already exists!\n" + color(connection, "reset", "reset")))
			return
		}
	}
	if connection.session.character.coins < 50000 {
		connection.store.Write([]byte("  You do not have enough coins to create a clan!\n"))
		connection.store.Write([]byte("  You need " + color(connection, "yellow", "tp") + strconv.Itoa(50000-connection.session.character.coins) + color(connection, "reset", "reset") + " more gold! (50K Total)"))
		return
	}
	AskConfirm(connection, color(connection, "red", "tp")+"\n  Are you sure you want to create a clan?"+color(connection, "reset", "reset")+"\n  Creation will cost "+color(connection, "yellow", "tp")+"50,000"+color(connection, "reset", "reset")+" gold! ("+color(connection, "green", "tp")+"yes"+color(connection, "reset", "reset")+"/"+color(connection, "red", "tp")+"no"+color(connection, "reset", "reset")+")\n\n  => ", func(val bool, db *sql.DB, world *World, conn *ConnectionData) {
		if val {
			prom, err := db.Exec("INSERT INTO clans (name, tag, owner_id, status) VALUES (?, ?, ?, ?)", name, "", connection.session.id, "open")
			if err != nil {
				fmt.Println(err)
				connection.store.Write([]byte("\n  Failed to create clan, please try again.\n"))
				return
			}
			id, _ := prom.LastInsertId()
			var createdAt time.Time
			err1 := db.QueryRow("SELECT created_at FROM clans WHERE id = ?", id).Scan(&createdAt)
			if err1 != nil {
				fmt.Println(err1)
			}
			clan := &Clan{int(id), name, "", connection.session.id, createdAt, "open", []int{}}
			world.clans[int(id)] = clan
			connection.session.character.clan = world.clans[int(id)]
			_, err2 := db.Exec("UPDATE players SET (clan_id) = (?) WHERE id = ?", int(id), connection.session.id)
			if err2 != nil {
				fmt.Println(err2)
			}
			connection.session.character.coins -= 50000
			connection.store.Write([]byte("\n  You have created the clan " + color(connection, "cyan", "tp") + clan.name + color(connection, "reset", "reset") + "!\n"))
			connection.store.Write([]byte("\n  You spent " + color(connection, "yellow", "tp") + "50,000 gold" + color(connection, "reset", "reset") + " to establish the clan!\n\n> "))
		} else {
			connection.store.Write([]byte("\n  You denied the creation of clan.\n\n> "))
		}
	})
}

func JoinClan(connection *ConnectionData, world *World, db *sql.DB, name string) {
	if connection.session.character.clan != nil {
		connection.store.Write([]byte("\n  You are in already in a clan!\n  Leave your clan to join another one!\n"))
		return
	}
	clanFound := false
	for _, c := range world.clans {
		if c.name == name {
			if c.status == "open" {
				world.clans[c.id].members = append(world.clans[c.id].members, connection.session.id)
				_, err := db.Exec("UPDATE players SET (clan_id) = (?) WHERE id = ?", c.id, connection.session.id)
				connection.session.character.clan = world.clans[c.id]
				connection.store.Write([]byte("\n  You have joined " + color(connection, "green", "tp") + c.name + color(connection, "reset", "reset") + "!\n"))
				if err != nil {
					fmt.Println(err)
				}
			} else {
				connection.store.Write([]byte("\n  This clan is not accepting new members!\n"))
			}
			clanFound = true
			break
		}
	}
	if !clanFound {
		connection.store.Write([]byte("\n  Clan not found!\n"))
	}
}

func LeaveClan(connection *ConnectionData, world *World, db *sql.DB) {
	if connection.session.character.clan == nil {
		connection.store.Write([]byte("\n  You are not in a clan!\n"))
		return
	}
	c := world.clans[connection.session.character.clan.id]
	if connection.session.id == c.ownerID {
		if len(c.members) == 0 {
			AskConfirm(connection, "\n  Are you sure?\n  This will "+color(connection, "red", "tp")+"DISBAND"+color(connection, "reset", "reset")+" the clan forever! ("+color(connection, "green", "tp")+"yes"+color(connection, "reset", "reset")+"/"+color(connection, "red", "tp")+"no"+color(connection, "reset", "reset")+")\n\n  => ", func(val bool, db *sql.DB, world *World, conn *ConnectionData) {
				if val {
					connection.store.Write([]byte("\n  The clan " + color(connection, "green", "tp") + c.name + color(connection, "reset", "reset") + " has been disbanded!\n\n> "))
					db.Exec("DELETE FROM clans WHERE id = ?", c.id)
					db.Exec("UPDATE players SET (clan_id) = (?) WHERE id = ?", nil, connection.session.id)
					connection.session.character.clan = nil
					world.clans[c.id] = nil
					delete(world.clans, c.id)
				} else {
					connection.store.Write([]byte("\n  You rejected disbandment of the clan.\n\n> "))
				}
			})
		} else {
			var newOwnerName string
			newOwnerID := world.clans[c.id].members[rand.Intn(len(world.clans[c.id].members))]
			db.QueryRow("SELECT username FROM players WHERE id = ?", newOwnerID).Scan(&newOwnerName)
			AskConfirm(connection, "\n  Are you sure you want to leave?\n  The leadership will be transferred to "+color(connection, "cyan", "tp")+newOwnerName+color(connection, "reset", "reset")+"! ("+color(connection, "green", "tp")+"yes"+color(connection, "reset", "reset")+"/"+color(connection, "red", "tp")+"no"+color(connection, "reset", "reset")+")\n\n  => ", func(val bool, db *sql.DB, world *World, conn *ConnectionData) {
				if val {
					connection.store.Write([]byte("\n  You have left your clan!"))
					connection.store.Write([]byte("\n  The new clan leader is " + color(connection, "cyan", "tp") + newOwnerName + color(connection, "reset", "reset") + "!\n\n> "))
					db.Exec("UPDATE players SET (clan_id) = (?) WHERE id = ?", nil, connection.session.id)
					db.Exec("UPDATE players SET (clan_id) = (?) WHERE id = ?", c.id, newOwnerID)
					db.Exec("UPDATE clans SET (owner_id) = (?) WHERE id = ?", newOwnerID, c.id)
					connection.session.character.clan = nil
					world.clans[c.id].ownerID = newOwnerID
					world.clans[c.id].members = slices.DeleteFunc(world.clans[c.id].members, func(e int) bool {
						return e == newOwnerID
					})
				} else {
					connection.store.Write([]byte("\n  You rejected leaving the clan.\n\n> "))
				}
			})
		}
	} else {
		connection.store.Write([]byte("\n  You have left your clan!\n"))
		world.clans[c.id].members = slices.DeleteFunc(world.clans[c.id].members, func(id int) bool {
			return id == connection.session.id
		})
		db.Exec("UPDATE players SET (clan_id) = (?) WHERE id = ?", nil, connection.session.id)
		connection.session.character.clan = nil
	}
}

func KickFromClan(connection *ConnectionData, world *World, db *sql.DB, name string) {
	if connection.session.character.clan == nil {
		connection.store.Write([]byte("\n  You are not in a clan!\n"))
		return
	}
	if connection.session.id != connection.session.character.clan.ownerID {
		connection.store.Write([]byte("\n  You are not the clan leader!\n"))
		return
	}
	if name == connection.session.username {
		connection.store.Write([]byte("\n  You cannot kick yourself!\n"))
		return
	}
	c := world.clans[connection.session.character.clan.id]
	var pcon *ConnectionData
	var pID int
	var pFound bool
	pExists := false
	for _, p := range world.playerList {
		if p.username == name {
			pExists = true
			for i, cn := range world.connections {
				if cn.session != nil && cn.session.id == p.id && cn.session.character != nil && cn.session.character.clan != nil && cn.session.character.clan.id == c.id {
					pcon = world.connections[i]
					pID = p.id
					pFound = true
					break
				}
			}
			if pFound == false {
				var cID *int
				err := db.QueryRow("SELECT id, clan_id FROM players WHERE username = ?", name).Scan(&pID, &cID)
				if err != nil {
					fmt.Println(err)
					return
				}
				if cID == nil || *cID != c.id {
					connection.store.Write([]byte("\n  Player not found in clan!\n"))
					return
				}
			}
			break
		}
	}
	if pExists {
		AskConfirm(connection, "\n  Are you sure?\n  This will kick "+color(connection, "cyan", "tp")+name+color(connection, "reset", "reset")+" from the clan! ("+color(connection, "green", "tp")+"yes"+color(connection, "reset", "reset")+"/"+color(connection, "red", "tp")+"no"+color(connection, "reset", "reset")+")\n\n  => ", func(val bool, db *sql.DB, world *World, conn *ConnectionData) {
			if val {
				if pcon != nil && pcon.session != nil && pcon.session.character != nil {
					pcon.store.Write([]byte("\x1b[2K\r  You have been " + color(pcon, "red", "tp") + "kicked" + color(pcon, "reset", "reset") + " from the clan " + color(pcon, "green", "tp") + pcon.session.character.clan.name + color(pcon, "reset", "reset") + "!\n\n> "))
					pcon.session.character.clan = nil
					world.clans[c.id].members = slices.DeleteFunc(world.clans[c.id].members, func(e int) bool {
						return e == pID
					})
				} else {
					db.Exec("UPDATE players SET clan_id = ? WHERE id = ?", nil, pID)
					world.clans[c.id].members = slices.DeleteFunc(world.clans[c.id].members, func(e int) bool {
						return e == pID
					})
				}
				conn.store.Write([]byte("\n  You have kicked " + color(conn, "cyan", "tp") + name + color(conn, "reset", "reset") + " from the clan!\n\n> "))
			} else {
				connection.store.Write([]byte("\n  You decided against kicking them.\n\n> "))
			}
		})
	} else {
		connection.store.Write([]byte("\n  Player not found in clan!\n"))
	}
}

func UpdateClanTag(connection *ConnectionData, world *World, db *sql.DB, tag string) {
	if connection.session.character.clan == nil {
		connection.store.Write([]byte("\n  You are not in a clan!\n"))
		return
	}
	if connection.session.id != connection.session.character.clan.ownerID {
		connection.store.Write([]byte("\n  You are not the clan leader!\n"))
		return
	}
	connection.store.Write([]byte("\n  Updated clan tag from [" + color(connection, "red", "white") + connection.session.character.clan.tag + color(connection, "reset", "reset") + "] to [" + color(connection, "red", "white") + tag + color(connection, "reset", "reset") + "]!\n"))
	world.clans[connection.session.character.clan.id].tag = tag
	db.Exec("UPDATE clans SET (tag) = (?) WHERE id = ?", tag, connection.session.character.clan.id)
}

func UpdateClanStatus(connection *ConnectionData, world *World, db *sql.DB, status string) {
	if connection.session.character.clan == nil {
		connection.store.Write([]byte("\n  You are not in a clan!\n"))
		return
	}
	if connection.session.id != connection.session.character.clan.ownerID {
		connection.store.Write([]byte("\n  You are not the clan leader!\n"))
		return
	}
	connection.store.Write([]byte("\n  Updated clan status from " + connection.session.character.clan.status + " to " + status + "!\n"))
	world.clans[connection.session.character.clan.id].status = status
	db.Exec("UPDATE clans SET (status) = (?) WHERE id = ?", status, connection.session.character.clan.id)
}

func PrintClanInfo(conn *ConnectionData, world *World) {
	stream := conn.store

	clan := conn.session.character.clan
	cardLength := 50
	var nameMedian int
	if len(conn.session.character.clan.name)%2 == 1 {
		nameMedian = int(math.Floor(float64(len(conn.session.character.clan.name))/2.0)) + 1
	} else {
		nameMedian = len(conn.session.character.clan.name)/2 + 1
	}

	if conn.isPrettyEnabled {
		us := "  │ " + strings.Repeat(" ", cardLength/2-nameMedian) + color(conn, "green", "tp") + clan.name + color(conn, "reset", "reset")
		estDate := clan.createdAt.Format("Monday, 02 Jan 2006")
		stream.Write([]byte("\n  ╭" + strings.Repeat("─", cardLength) + "╮\n"))
		stream.Write([]byte(us + strings.Repeat(" ", cardLength-visibleLen(us)+4) + " │\n"))
		stream.Write([]byte("  ├" + strings.Repeat("─", cardLength) + "┤\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", cardLength) + "│\n"))
		stream.Write([]byte("  │" + "  Owner          : " + color(conn, "cyan", "tp") + world.playerList[clan.ownerID].username + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-19-len(world.playerList[clan.ownerID].username)) + "│\n"))
		stream.Write([]byte("  │" + "  Status         : " + clan.status + strings.Repeat(" ", cardLength-19-len(clan.status)) + "│\n"))
		stream.Write([]byte("  │" + "  Tag            : [" + color(conn, "red", "white") + clan.tag + color(conn, "reset", "reset") + "]" + strings.Repeat(" ", cardLength-21-len(clan.tag)) + "│\n"))
		stream.Write([]byte("  │" + "  Established At : " + color(conn, "magenta", "tp") + estDate + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-19-len(estDate)) + "│\n"))
		stream.Write([]byte("  │" + strings.Repeat(" ", cardLength) + "│\n"))
		stream.Write([]byte("  │" + "  Members" + strings.Repeat(" ", cardLength-9) + "│\n"))
		if len(clan.members) == 0 {
			stream.Write([]byte("  │" + color(conn, "yellow", "tp") + "    Empty :(" + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-12) + "│\n"))
		} else {
			for _, mi := range clan.members {
				m := world.playerList[mi]
				ob := m.online
				o := " "
				if ob == true {
					o = color(conn, "green", "tp") + "!" + color(conn, "reset", "reset")
				} else {
					o = color(conn, "cyan", "tp") + "*" + color(conn, "reset", "reset")
				}
				stream.Write([]byte("  │" + color(conn, "cyan", "tp") + "    " + o + " " + color(conn, "yellow", "tp") + m.username + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-6-len(m.username)) + "│\n"))
			}
		}
		stream.Write([]byte("  │" + strings.Repeat(" ", cardLength) + "│\n"))
		stream.Write([]byte("  ╰" + strings.Repeat("─", cardLength) + "╯\n"))
	} else {
		us := "  | " + strings.Repeat(" ", cardLength/2-nameMedian) + color(conn, "green", "tp") + clan.name + color(conn, "reset", "reset")
		estDate := clan.createdAt.Format("Monday, 02 Jan 2006")
		stream.Write([]byte("\n  +" + strings.Repeat("-", cardLength) + "+\n"))
		stream.Write([]byte(us + strings.Repeat(" ", cardLength-visibleLen(us)+2) + " |\n"))
		stream.Write([]byte("  +" + strings.Repeat("-", cardLength) + "+\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", cardLength) + "|\n"))
		stream.Write([]byte("  |" + "  Owner          : " + color(conn, "cyan", "tp") + world.playerList[clan.ownerID].username + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-19-len(world.playerList[clan.ownerID].username)) + "|\n"))
		stream.Write([]byte("  |" + "  Status         : " + clan.status + strings.Repeat(" ", cardLength-19-len(clan.status)) + "|\n"))
		stream.Write([]byte("  |" + "  Tag            : [" + color(conn, "red", "white") + clan.tag + color(conn, "reset", "reset") + "]" + strings.Repeat(" ", cardLength-21-len(clan.tag)) + "|\n"))
		stream.Write([]byte("  |" + "  Established At : " + color(conn, "magenta", "tp") + estDate + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-19-len(estDate)) + "|\n"))
		stream.Write([]byte("  |" + strings.Repeat(" ", cardLength) + "|\n"))
		stream.Write([]byte("  |" + "  Members" + strings.Repeat(" ", cardLength-9) + "|\n"))
		if len(clan.members) == 0 {
			stream.Write([]byte("  |" + color(conn, "yellow", "tp") + "    Empty :(" + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-12) + "|\n"))
		} else {
			for _, mi := range clan.members {
				m := world.playerList[mi]
				ob := m.online
				o := " "
				if ob == true {
					o = color(conn, "green", "tp") + "!" + color(conn, "reset", "reset")
				} else {
					o = color(conn, "cyan", "tp") + "*" + color(conn, "reset", "reset")
				}
				stream.Write([]byte("  |" + color(conn, "cyan", "tp") + "    " + o + " " + color(conn, "yellow", "tp") + m.username + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-6-len(m.username)) + "|\n"))
			}
		}
		stream.Write([]byte("  |" + strings.Repeat(" ", cardLength) + "|\n"))
		stream.Write([]byte("  +" + strings.Repeat("-", cardLength) + "+\n"))
	}
}

func PrintClanTop(conn *ConnectionData, world *World) {
	stream := conn.store

	stream.Write([]byte(color(conn, "magenta", "tp") + "\n  * " + color(conn, "black", "white") + "[ ID ]" + color(conn, "magenta", "reset") + " * " + color(conn, "black", "white") + "[  Clan Name  ]" + color(conn, "magenta", "reset") + " * " + color(conn, "black", "white") + "[ Status ]" + color(conn, "magenta", "reset") + " * " + color(conn, "black", "white") + "[  Owner  ]" + color(conn, "magenta", "reset") + " *\n" + color(conn, "reset", "reset")))
	sorted := make([]*Clan, 0, len(world.clans))
	for _, c := range world.clans {
		sorted = append(sorted, c)
	}
	slices.SortFunc(sorted, func(a, b *Clan) int {
		return len(a.members) - len(b.members)
	})

	for _, c := range sorted {
		n := c.name
		ou := world.playerList[c.ownerID].username
		id := strconv.Itoa(c.id)
		s := c.status
		stream.Write([]byte("     " + id + strings.Repeat(" ", 4-len(id)) + "  -  " + color(conn, "green", "tp") + n + strings.Repeat(" ", 15-len(n)) + color(conn, "reset", "reset") + "-  " + color(conn, "magenta", "tp") + s + strings.Repeat(" ", 10-len(s)) + color(conn, "reset", "reset") + "-  " + color(conn, "cyan", "tp") + ou + color(conn, "reset", "reset") + "\n"))
	}
	if len(sorted) == 0 {
		stream.Write([]byte(color(conn, "magenta", "tp") + "  Empty :(\n" + color(conn, "reset", "reset")))
	}
}
