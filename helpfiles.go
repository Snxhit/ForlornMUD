package main

import (
	"strconv"
	"strings"
)

type HelpEntry struct {
	id          int
	title       string
	keywords    []string
	instruction []string
	category    string
	text        []string
}

type HelpIndex struct {
	entries []HelpEntry
}

var HelpEntries = []HelpEntry{
	{
		id:          0,
		title:       "New player",
		keywords:    []string{"newplayer", "newbie", "tutorial", "new"},
		instruction: []string{""},
		category:    "Tutorial",
		text: []string{
			//"is command lists out information about any topic you'd",
			"Hey there! Welcome to ForlornMUD!",
			"To get started, read these helpfiles in order:",
			"I promise it won't take too long :)",
			"",
			"- help help",
			"- help commands",
			"- help movement",
			"- help rooms",
			"- help indexing",
			"- help items",
			"- help mobs",
			"- help profile",
			"- help level",
			"- help merchant",
			"- help flavortown",
			"",
			"You can always come back to this tutorial if you get",
			"lost, using help newplayer!",
			"",
			"If you don't wanna read all of those, then:",
			"",
			"- help basics",
			"",
			"If you're confused about any command, try:",
			"- help <command_name>",
			"There's a helpfile for almsot every command!",
			"",
			"At the very least, it is recommended to read",
			"the helpfiles of movement, rooms, indexing, mobs!",
			"",
			"You may also choose to watch the preview video that",
			"that goes through all the basic mechanics.",
			"You may find said video in the screenshots section",
			"of the GitHub README!",
			"",
			"Should sum it all up pretty nicely. It's like a TLDR!",
			"Enjoy :D!",
		},
	},
	{
		id:          1,
		title:       "Help",
		keywords:    []string{"help", "helpfiles", "query", "question", "stuck"},
		instruction: []string{"help <query>"},
		category:    "Command",
		text: []string{
			"This command lists out information about anything you'd",
			"like! Whether it's a command, a topic, or something else.",
			"",
			"Helpfiles can be accessed by their names or their",
			"index!",
			"",
			"For example, the name of this file is help. But, it can",
			"also be accessed by using `help 1`, because it's index",
			"is 1!",
			"Similarly, the newplayer helpfile can be accessed by",
			"`help 0` because it's index is 0!",
		},
	},
	{
		id:          2,
		title:       "Movement",
		keywords:    []string{"move", "movement", "directions"},
		instruction: []string{"<north|south|west|east>", "m <n|s|w|e>"},
		category:    "Command",
		text: []string{
			"Choose one of the four cardinal directions to move",
			"around the world of ForlornMUD!",
			"",
			"Every room has 4 possible exits, North, South, West,",
			"and East. You can go in any direction you'd like,",
			"as long as there is an exit in that direction,",
			"that you can see in the `look` command.",
		},
	},
	{
		id:          3,
		title:       "Rooms",
		keywords:    []string{"rooms", "room", "directions", "world"},
		instruction: []string{""},
		category:    "Topic",
		text: []string{
			"Rooms are singular units that together make up the world",
			"",
			"Rooms can have 4 possible exits in all 4 cardinal",
			"directions.",
			"",
			"Rooms can contain entities, items and players. You can",
			"look around in a room by using the `look`",
			"command, which lists the room name, description and",
			"the entities, items and players present in the room!",
			"",
			"You move around by moving between rooms. Try it!",
		},
	},
	{
		id:          4,
		title:       "Items",
		keywords:    []string{"items", "item", "drops", "armor"},
		instruction: []string{""},
		category:    "Topic",
		text: []string{
			"Items are pretty self-explanatory.",
			"",
			"Items can be found on the ground in rooms, and can be",
			"dropped by entities. They can be picked up with",
			"`pickup <item>` or `take <item>`.",
			"",
			"You can find out more about items using the `examine`",
			"command on them, as done in `help indexing`",
			"",
			"Different item types serve different purposes:",
			"",
			"- consumable: Items you can consume for benefits.",
			"- item: Items that do not serve a purpose.",
			"- mainhand, offhand, head, body, legs, ring",
			"  These correspond to equippable items!",
			"",
			"You can wear equippables using the `wear` command!",
		},
	},
	{
		id:          5,
		title:       "Entities",
		keywords:    []string{"mobs", "entities", "entity", "mob"},
		instruction: []string{""},
		category:    "Topic",
		text: []string{
			"Entities have stats, base damage, base defense, and",
			"you can engage in combat with them!",
			"",
			"You can engage in combat by using `fight <entity>`",
			"Where <entity> is the addressed entity as you learnt",
			"in `help indexing`.",
			"",
			"When defeated, entities give coins, exp, and even items!",
			"",
			"There are also merchant entities.",
		},
	},
	{
		id:          6,
		title:       "Profile",
		keywords:    []string{"profile", "pf", "info"},
		instruction: []string{"profile", "pf"},
		category:    "Command",
		text: []string{
			"The profile command lists out a lot of information",
			"in a prettily formatted card.",
		},
	},
	{
		id:          7,
		title:       "Level",
		keywords:    []string{"level", "levels", "leveling", "exp"},
		instruction: []string{""},
		category:    "Topic",
		text: []string{
			"You get exp by defeating entities, and for each 100 exp",
			"you gain 1 level. Each level you gain gives you",
			"5 trains, which can be used to increase your stats,",
			"and make your character stronger with `train <number> <stat>`!",
			"",
			"You can train str, dex, agi, stam, int, and hp!",
			"",
			"Use the `profile` command to view everything in one",
			"place!",
		},
	},
	{
		id:          8,
		title:       "Merchant",
		keywords:    []string{"merchant", "shop", "buy", "sell", "merchants"},
		instruction: []string{"list <entity>"},
		category:    "Topic",
		text: []string{
			"Merchants are just entities that buy and sell items.",
			"you can engage in combat with merchants too, but, that's",
			"not recommended... find out why.",
			"",
			"You can address a merchant the same way you'd address an",
			"entity, by using either a keyword from it's name or it's",
			"index in the room's `look` list.",
			"",
			"Merchants have a 'M' next to their name. Look out!",
			"",
			"Please read \"help indexing\" to understand how to",
			"address merchants!",
		},
	},
	{
		id:          9,
		title:       "Indexing",
		keywords:    []string{"addressing", "indexing", "access", "keywords"},
		instruction: []string{""},
		category:    "Topic",
		text: []string{
			"Entities and items are addressed the same way. Both work",
			"by keyword, and also by their number in the `look` list.",
			"Both have completely different commands, and as such,",
			"there is no confusion between the two.",
			"",
			"Imagine this. You're in a room, this is what you see:",
			"",
			"Surrounding",
			"  Dummy Entity One",
			"     Looks beatable.",
			"  Dummy Entity Two",
			"     Looks beatable.",
			"  Dummy Merchant M",
			"     Has something to sell.",
			"  2   | Dummy Item",
			"",
			"Then, this is how you'd address the first entity:",
			"`attack dummy`. This defaults to the first entity that",
			"has the keyword you specify (in this case, dummy) in",
			"their name.",
			"And, similarly:",
			"`take 0`. This specifies the first item that you see.",
			"`attack 1`. This specifies the second entity that you see.",
			"`list dummy`. This lists the shop of the first merchant.",
			"`list merchant` would also refer to the first merchant,",
			"because dummy and merchant are both in it's name!",
			"",
			"P.S. Indexes start from 0 instead of 1!",
			"P.S. For merchants, numbers do NOT work!",
			"",
			"Easy enough, no? This works for `list`, `drop`,",
			"and every other command :)",
		},
	},
	{
		id:          10,
		title:       "Flavortown",
		keywords:    []string{"flavortown", "flavor", "ft"},
		instruction: []string{""},
		category:    "Topic",
		text: []string{
			"This game includes a Flavortown-based map!",
			"",
			"`help move` to start exploring it!",
			"",
			"Explore around, and don't hesitate to fight any",
			"enemies that you see, or pick up any items you",
			"come across.",
			"",
			"Try fighting Flavorpheus!",
			"Hint: The Framework Laptop is a great weapon!",
			"Hint: Go North thrice.",
			"",
			"Don't worry, the items and entities respawn :)",
		},
	},
	{
		id:          11,
		title:       "Basics",
		keywords:    []string{"basic", "basics", "tldr"},
		instruction: []string{""},
		category:    "Topic",
		text: []string{
			"TLDR of basics! For detailed info, read help newplayer",
			"",
			"HELP: Use `help <query>` to get info on any topic. Access",
			"helpfiles by name or index number (e.g., `help 1`).",
			"",
			"MOVEMENT: Move using north/south/east/west or n/s/e/w.",
			"Rooms have 4 possible exits in cardinal directions.",
			"Use `look` to see exits, entities, items, and players.",
			"",
			"INDEXING: Too long to sum up. Use `help indexing`!",
			"",
			"ITEMS: Pick up with `take` or `pickup`. Use `examine`",
			"for details. Types: consumable, equippables (mainhand,",
			"offhand, head, body, legs, ring). Equip with `wear`.",
			"",
			"ENTITIES: Have stats and can be fought with `fight",
			"<entity>`. Drop coins, exp, and items when defeated.",
			"Some are merchants (use `list <entity>` to see items).",
			"",
			"LEVELING: Gain exp from combat. 100 exp = 1 level.",
			"Each level gives 5 trains to boost stats with `train",
			"<number> <stat>`. View progress with `profile` or `pf`.",
			"",
			"MERCHANTS: Entities that buy/sell. Address like normal",
			"entities. Use `list <entity>` to see their inventory.",
		},
	},
	{
		id:          12,
		title:       "Quit",
		keywords:    []string{"exit", "leave", "quit"},
		instruction: []string{"quit", "exit"},
		category:    "Command",
		text: []string{
			"You can exit the game by using the `quit` or `exit`",
			"commands.",
			"Upon exit, your character is automatically saved.",
		},
	},
	{
		id:          13,
		title:       "Inventory",
		keywords:    []string{"inventory", "inv", "storage"},
		instruction: []string{"inventory", "inv", "i"},
		category:    "Command",
		text: []string{
			"You can use the `inventory` or `inv` or `i` command to",
			"get a list of all the items you're currently carrying.",
			"",
			"Items are stackable, and are shown like this:",
			"  q   | name",
			"Where q is the quantity of the item, and name is the",
			"name of the item.",
			"",
			"If the output is empty, then your inventory is empty.",
		},
	},
	{
		id:          14,
		title:       "Effects",
		keywords:    []string{"effects", "effs", "buffs", "itembuffs"},
		instruction: []string{"effects", "effs"},
		category:    "Command",
		text: []string{
			"The `effects` or `effs` command can be used to get a list",
			"of all active effects that you have.",
			"",
			"Some equippable items provide 'modifiers' which increase",
			"(or decrease) your stats. These modifiers can be checked",
			"for any item in particular via the `examine` command.",
			"But, the `effects` command gives you a list of al active",
			"modifiers currently active on your character.",
		},
	},
	{
		id:          15,
		title:       "Equipped",
		keywords:    []string{"equipped", "worn", "armor"},
		instruction: []string{"equipped", "worn", "armor"},
		category:    "Command",
		text: []string{
			"The `equipped`, `worn`, and `armor` commands show you a",
			"list of all your usable armor slots and the item",
			"equipped in each slot (if any.)",
			"",
			"Example:",
			"  + head     - <empty>",
			"  - mainhand - Rusted Spoon",
		},
	},
	{
		id:          16,
		title:       "Pickup",
		keywords:    []string{"pickup", "collect", "take", "pick"},
		instruction: []string{"pickup <item>", "take <item>", "collect <item>"},
		category:    "Command",
		text: []string{
			"Used to put an item into your inventory from the room.",
			"",
			"Where <item> is the item you want to pick up, indexed",
			"as per `help indexing`!",
		},
	},
	{
		id:          17,
		title:       "Drop",
		keywords:    []string{"drop", "put", "throw"},
		instruction: []string{"drop <item>", "throw <item>"},
		category:    "Command",
		text: []string{
			"Used to remove an item from your inventory and put it in",
			"the room you're currently in.",
			"",
			"Where <item> is the item you want to drop, indexed",
			"as per `help indexing`!",
		},
	},
	{
		id:          18,
		title:       "Examine",
		keywords:    []string{"examine", "lookat", "observe"},
		instruction: []string{"examine <item>"},
		category:    "Command",
		text: []string{
			"This command is used to display detailed information",
			"about an item that is in YOUR inventory.",
			"",
			"You cannot use this command on items that are on the",
			"ground!",
		},
	},
	{
		id:          19,
		title:       "Equip",
		keywords:    []string{"equip", "wear"},
		instruction: []string{"equip <item>", "wear <item>"},
		category:    "Command",
		text: []string{
			"This command is used to wear an equippable item.",
			"",
			"If the slot of the item you're trying to wear is already",
			"equipped, then the currently equipped item is unequipped,",
			"and put in your inventory, and the item you addressed is",
			"equipped in its place.",
		},
	},
	{
		id:          20,
		title:       "Unequip",
		keywords:    []string{"unequip", "remove"},
		instruction: []string{"unequip <slot>", "remove <slot>"},
		category:    "Command",
		text: []string{
			"This command is used to free an equippable item slot.",
			"",
			"Where <slot> is any of the following:",
			"mainhand, offhand, head, body, legs, ring",
		},
	},
	{
		id:          21,
		title:       "Use",
		keywords:    []string{"use", "u"},
		instruction: []string{"use <item>", "u <item>"},
		category:    "Command",
		text: []string{
			"This command is used to consume a consumable.",
			"",
			"After consumption, the item's effect will do their job.",
			"An item can also have negative effects (-10 hp, etc.)",
		},
	},
	{
		id:          22,
		title:       "Fight",
		keywords:    []string{"fight", "kick", "attack"},
		instruction: []string{"fight <entity>", "attack <entity>"},
		category:    "Command",
		text: []string{
			"This command is used to initiate combat with an entity.",
			"",
			"Where <entity> is the entity you want to attack, indexed",
			"as per `help indexing`!",
		},
	},
	{
		id:          23,
		title:       "List",
		keywords:    []string{"list", "shop", "browse"},
		instruction: []string{"list <entity>", "browse <entity>"},
		category:    "Command",
		text: []string{
			"HELP: Use `help <query>` to get info on any topic. Access",
			"This command is used to get a list of items in a merchant",
			"shop.",
			"",
			"Where <entity> is the entity you want to list, indexed",
			"as per `help indexing`!",
			"",
			"Any entity can be a merchant. Try `list <entity>` on any",
			"entity you suspect to be a merchant. This is by design.",
		},
	},
	{
		id:          25,
		title:       "Toggle",
		keywords:    []string{"toggle", "pretty", "colors", "disable"},
		instruction: []string{"toggle <colors|pretty>"},
		category:    "Command",
		text: []string{
			"This command is used to enable or disable these:",
			"- Pretty unicodes (disables seamless borders on cards)",
			"- Colors (disables colored output everywhere)",
			"This option exists to provide maximum compatibility to",
			"all kinds of terminals, and also just as a QoL choice :D",
		},
	},
	{
		id:          26,
		title:       "Character",
		keywords:    []string{"character", "char", "userprofile", "pvp"},
		instruction: []string{"char <char_name>", "character <char_name>"},
		category:    "Command",
		text: []string{
			"The profile command lists out a lot of information",
			"about another player in a prettily formatted card.",
		},
	},
	{
		id:          27,
		title:       "Communication",
		keywords:    []string{"sayto", "say", "tell", "talk", "dm"},
		instruction: []string{"sayto <char_name> <msg>", "tell <char_name> <msg>"},
		category:    "Command",
		text: []string{
			"This command allows you to directly, privately ",
			"communicate with another player!",
			"",
			"The target player can be anywhere in the game,",
			"but they must be online!",
		},
	},
	{
		id:          28,
		title:       "Commands",
		keywords:    []string{"commands", "command", "cmd", "cmds"},
		instruction: []string{},
		category:    "Topic",
		text: []string{
			"Commands are input you give to the game, which allows",
			"you to interact with it.",
			"Since you're here, you already know how to use commands",
			"so I won't explain further.",
			"",
			"Below are all the commands that exist in ForlornMUD.",
			"",
			"- help, h - Shows detailed explanations about anything",
			"- look, l - Look around to see entities, items,",
			"  players, and all exits.",
			"- north, n - Move North",
			"- south, s - Move South",
			"- west, w - Move West",
			"- east, e - Move East",
			"- take, pickup - Take an item from the ground",
			"- drop, throw - Drop an item from your inv on the ground",
			"- inv, i - Take a look at the items in your inv",
			"- profile, pf - Get a profile card detailing a lot of",
			"    useful info about your character",
			"- character, char - Get the profile card of another char",
			"- examine - Gives detailed info about an item in your inv",
			"- equipped - Shows equipment you have equipped",
			"- effects, effs - Shows all active item buffs",
			"- use, u - Uses a consumable item (healing)",
			"- equip, eq - Equips a wearable item (armor)",
			"- uneq, uneq - Unequips a worn item from the slot specified",
			"- fight, attack, f - Initiates a fight with a player or mob",
			"- train - Increases a selected stat by the given number of trains",
			"- list - Lists the buy/sell inventory of a merchant",
			"- buy - Buys an item of the specified ID from a merchant",
			"- sell - Sells an item from your inv to a merchant",
			"- sayto, tell - Allows you to talk to other players",
			"- toggle - Toggles pretty unicode and/or colors",

			"exit, quit - Disconnect from the game",
			"",
			"You can view detailed information about each command",
			"by using `help <command_name>!",
		},
	},
}

var h = HelpIndex{HelpEntries}

func searchHelpfiles(query string) []int {
	var result []int
	valid, i := validateInt(query)
	if !valid {
		for id, file := range h.entries {
			for _, k := range file.keywords {
				if query == k {
					result = append(result, id)
				}
			}
		}
	} else {
		for id, file := range h.entries {
			if file.id == i {
				result = append(result, id)
			}
		}
	}
	return result
}

func printHelpFile(query string, conn *ConnectionData) {
	hf := searchHelpfiles(query)
	if len(hf) == 0 {
		conn.store.Write([]byte("\n  No helpfiles found!\n"))
	} else if len(hf) != 1 {
		cardLength := 42
		stream := conn.store
		stream.Write([]byte("\n  " + glphys(conn, "tlc") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "trc") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "sl") + " Found multiple helpfiles!" + strings.Repeat(" ", cardLength-26) + glphys(conn, "sl") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "rtj") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "ltj") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "sl") + strings.Repeat(" ", cardLength) + glphys(conn, "sl") + "\n"))
		if !conn.isColorEnabled {
			stream.Write([]byte("  " + glphys(conn, "sl") + "  [ ID ] + [   TITLE   ] + [  KEYWORD  ]" + strings.Repeat(" ", cardLength-40) + glphys(conn, "sl") + "\n"))
		} else {
			stream.Write([]byte("  " + glphys(conn, "sl") + "  [\x1b[30;47m ID \x1b[39;49m] + [\x1b[30;47m   TITLE   \x1b[39;49m] + [\x1b[30;47m  KEYWORD  \x1b[39;49m]" + strings.Repeat(" ", cardLength-40) + glphys(conn, "sl") + "\n"))
		}
		for _, i := range hf {
			id := color(conn, "magenta", "tp") + strconv.Itoa(h.entries[i].id) + color(conn, "reset", "reset")
			t := color(conn, "cyan", "tp") + h.entries[i].title + color(conn, "reset", "reset")
			query = color(conn, "green", "tp") + query + color(conn, "reset", "reset")
			stream.Write([]byte("  " + glphys(conn, "sl") + "   " + strings.Repeat("0", 4-visibleLen(id)) + id + "     " + t + strings.Repeat(" ", 11-visibleLen(t)) + "     " + query + strings.Repeat(" ", 11-visibleLen(query)) + strings.Repeat(" ", cardLength-39) + glphys(conn, "sl") + "\n"))
		}
		stream.Write([]byte("  " + glphys(conn, "sl") + strings.Repeat(" ", cardLength) + glphys(conn, "sl") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "blc") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "brc") + "\n"))
	} else {
		cardLength := 60
		// id := strconv.Itoa(h.entries[hf[0]].id)
		t := color(conn, "cyan", "tp") + h.entries[hf[0]].title + color(conn, "reset", "reset")
		stream := conn.store
		stream.Write([]byte("\n  " + glphys(conn, "tlc") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "trc") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "sl") + " " + t + strings.Repeat(" ", cardLength-1-visibleLen(t)) + glphys(conn, "sl") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "rtj") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "ltj") + "\n"))
		var ks string = color(conn, "green", "tp")
		for _, k := range h.entries[hf[0]].keywords {
			ks += k + " "
		}
		if len(ks) > 50 {
			ks = ks[:50]
			ks += "..."
		}
		ks += color(conn, "reset", "reset")
		stream.Write([]byte("  " + glphys(conn, "sl") + " " + "Keywords: " + ks + strings.Repeat(" ", cardLength-11-visibleLen(ks)) + glphys(conn, "sl") + "\n"))
		ct := color(conn, "magenta", "tp") + h.entries[hf[0]].category + color(conn, "reset", "reset")
		stream.Write([]byte("  " + glphys(conn, "rtj") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "ltj") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "sl") + " " + "Category: " + ct + strings.Repeat(" ", cardLength-11-visibleLen(ct)) + glphys(conn, "sl") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "rtj") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "ltj") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "sl") + strings.Repeat(" ", cardLength) + glphys(conn, "sl") + "\n"))
		for i, it := range h.entries[hf[0]].instruction {
			if i == 0 && it == "" {
				break
			}
			stream.Write([]byte("  " + glphys(conn, "sl") + "  " + color(conn, "yellow", "tp") + it + color(conn, "reset", "reset") + strings.Repeat(" ", cardLength-2-len(it)) + glphys(conn, "sl") + "\n"))
		}
		for _, t := range h.entries[hf[0]].text {
			stream.Write([]byte("  " + glphys(conn, "sl") + "  " + t + strings.Repeat(" ", cardLength-2-len(t)) + glphys(conn, "sl") + "\n"))
		}
		stream.Write([]byte("  " + glphys(conn, "sl") + strings.Repeat(" ", cardLength) + glphys(conn, "sl") + "\n"))
		stream.Write([]byte("  " + glphys(conn, "blc") + strings.Repeat(glphys(conn, "sll"), cardLength) + glphys(conn, "brc") + "\n"))
	}
}

/*stream.Write([]byte("\n   [ ID ] + [QTY] + [ SELL ] + [  BUY  ] + [ NAME ]  \n"))
for _, item := range world.merchants[en].list {
	id := strconv.Itoa(world.ItemTemplates[item].id)
	sp := strconv.Itoa(int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].sellRate))
	bp := strconv.Itoa(int(float64(world.ItemTemplates[item].baseValue) * world.merchants[en].buyRate))
	stream.Write([]byte("    " + id + strings.Repeat(" ", 6-len(id)) + "   " + "inf     " + sp + strings.Repeat(" ", 6-len(sp)) + "     " + bp + strings.Repeat(" ", 7-len(bp)) + "     " + world.ItemTemplates[item].name + "\n"))
}*/
