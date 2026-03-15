package main

import (
	"slices"
	"time"
)

type NPC struct {
	id          int
	name        string
	description string
	locationID  int
	script      []NPCScriptStage
}

type scriptOp int

const (
	opSay scriptOp = iota
	opSayDiffRoom
	opAlert
	opWait
	opWaitDiffRoom
	opWaitCommand
	opWaitKill
	opWaitForLvl
	opWaitForReturn
	opIfClientWeb
)

type NPCScriptStage struct {
	op scriptOp

	text             string
	duration         time.Duration
	command          []string
	targetTemplateID int
}

type NPCDialogue struct {
	text  string
	delay time.Duration
}

type NPCConversation struct {
	NPC   *NPC
	stage int
}

var npcs = map[int]*NPC{
	1: &NPC{
		id:         1,
		name:       "Xamien, the Guide",
		locationID: 100,
		script: []NPCScriptStage{
			{op: opSay, text: "Welcome to the world of Forlorn!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "I'm your guide, Xamien, and I'll be explaining the basics of the game to you!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "I'll also be pausing in between to give you enough time to read everything :)"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "If you're a returning player or don't want to do the tutorial, you can go `north` from here till you get to the Welcome Plaza!"},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "That is where the Flavortown map starts!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "If you want a less detailed introduction or if you're in a hurry, use the command 'help newplayer' to see all the features at once!"},
			{op: opWait, duration: 7 * time.Second},
			{op: opSay, text: "If you leave the room mid conversation, the conversation is reset!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Anyways. Since you're still here..."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "In this world, you can move around with the commands: `north`, `south`, `west`, and `east`"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Those are also the only directions you can move in."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "When you move around, using the above commands you're automatically shown information about what you see every time."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "But, you can also take a look by using the command... You guessed it... `look`!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Try it right now! I'll wait for you. (Type: 'look')"},
			{op: opWaitCommand, command: []string{"look", "l"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Great job! Now you know how to perceive your surroundings."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "At the very top, the name of the room is displayed in green, followed by the description of the room!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "After that, you see the Exits section (Also green!), which displays all available paths you can go along!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "After that, the Surroundings section shows all items on the ground, entities in the room, players in the room and NPCs in the room."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "But, that's not to be worried about right now."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Now, this game has a system of helpfiles which are explanations written for basically anything and everything in the game."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Wanna know more about the look command? Go on, try `help look`! I'll wait. (Type: 'help look')"},
			{op: opWaitCommand, command: []string{"help look", "h look", "help l", "h l"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Take your time and read it! I'll wait."},
			{op: opWait, duration: 8 * time.Second},
			{op: opSay, text: "You're getting the hang of it now!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "That is not the only command which has a helpfile. Every command has one!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "If you're ever stuck, try `help topic/command`! Ex: `help command`"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Remember, you can always take the tutorials again by talking to me or any of the other instructors!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Now, let's get you to your first academy session, shall we?"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "I can't accompany you all the way there, but..."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Go `north` once, and I'll direct you futher. (Type: 'north')"},
			{op: opWaitCommand, command: []string{"north", "n", "go north", "go n"}},
			{op: opWaitDiffRoom, duration: 1 * time.Second},
			{op: opSayDiffRoom, text: "Good. Now go `east` once and you'll find your next teacher. Good luck!"},
			{op: opWaitDiffRoom, duration: 2 * time.Second},
			{op: opSayDiffRoom, text: "His name is Gwaarhar. Talk to him like you did to me. (`talk gwaarhar`)"},
		},
	},
	2: &NPC{
		id:         2,
		name:       "Gwaarhar, the Dwarf",
		locationID: 102,
		script: []NPCScriptStage{
			{op: opSay, text: "Oh. Welcome."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "My name is Gwaarhar."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "I'm going to teach you about addressing items, entities, and other players."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "You talked to me by using `talk gwaarhar`. But, you could do the same by using `talk dwarf`."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "That is because you address NPCs by using any word from their name. The word you use isn't case senstive."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "So, for instance, `talk GWAARHAR` or `talk GwAaRhAr` would've also did the same."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Now, did you notice that there's 3 wooden swords on the ground?"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Take a look with `look` just in case you didn't notice. I'll wait. (Type: 'look')"},
			{op: opWaitCommand, command: []string{"look", "l"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "To pick it up, the same system as the one used on NPCs works. Which I also call the keyword method."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "So, you could do `pickup wooden` or `pickup sword`, etc."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "But, what if you want to specifically pick up the third wooden sword in the room?"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Entities and items also have another way to index them, which is by using numbers."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Which means, you can do `pickup x` where x is a number."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Basically, this command makes a list of ALL items in the room, and lets you target any one of them by their number on the list."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "So for example, if there were: 2 Item A and 1 Item B in this room..."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "`pickup 2` would pickup the 3rd item, which would be the Item B!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Similarly, `pickup 0` would pickup the very 1st item, which would be the Item A!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Simple enough, right?"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Oh, and don't worry. Every item and entity has a respawn time. The wooden swords respawn in 3 seconds. Grab as many as you want after the class."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "Why don't you try pickup up a wooden sword with the keyword method? I'll wait. (Hint: `take <keyword>`)"},
			{op: opWaitCommand, command: []string{"pickup broken", "pickup wooden", "pickup sword", "take broken", "take wooden", "take sword", "pick broken", "pick wooden", "pick sword", "collect broken", "collect wooden", "collect sword"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Good job. Now, use numbers, and pickup specifically the 2nd Broken Wooden Sword. Remember, indexes start from 0! I'll be waiting. (Hint: `take <index>`)"},
			{op: opWaitCommand, command: []string{"pickup 1", "pickup 1", "pickup 1", "take 1", "take 1", "take 1", "pick 1", "pick 1", "pick 1", "collect 1", "collect 1", "collect 1"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Hey! You did it!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "You can view your inventory by using `inventory` or `inv` or even `i`!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Go on. Try it."},
			{op: opWaitCommand, command: []string{"inventory", "inv", "i"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "There's a limit to how many items you can carry, and you can see that limit in the inventory."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "You can increase your limit by gaining more dexterity stat. You'll learn about stats soon."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Just like you picked up items, you can also drop them. (`help drop` if needed)"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Now that you have basic movement and items down, what say we practice some combat, eh?"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "But first, you need to equip that Sword you picked up."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "As you can see, your inventory also lists out items in an order just like the list of items in the room that you see using the `look` command."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "And because of that, you can access items in your inventory the same way!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "To equip, you use the `equip` command!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "So, you can either do: `equip sword` or `equip 0`!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Try it, I'll wait!"},
			{op: opWaitCommand, command: []string{"equip broken", "equip wooden", "equip sword", "wear broken", "wear wooden", "wear sword", "eq broken", "eq wooden", "eq sword", "equip 0", "wear 0", "eq 0", "equip 1", "wear 1", "eq 1"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Great! Now, you can view your worn items by using `equipped` or `worn` commands! Try it, I'll wait!"},
			{op: opWaitCommand, command: []string{"equipped", "worn", "armor"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Anyway, this class has a training area attached to it towards the east."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "When you get there, you'll see some training dummies."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Look at their names, and use the command: `fight <keyword>` without the <>, where keyword is any word of their name, like I taught you."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "Alternatively, you may use the number system."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "If you get stuck, come back to me to repeat the tutorial or take a look at `help indexing`!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "TIP: Almost every frequently used command has a shorthand. Check `help <command>` without the <> to see if there are shorthands for any command you want!"},
			{op: opWait, duration: 7 * time.Second},
			{op: opSay, text: "Now, go `east` of here to explore the training grounds and find, kill a Small Training Dummy! Remember, you don't have to fight the others."},
			{op: opSay, text: "You might need to move around a bit to find your target. Now go. (Go `east`)"},
			{op: opWaitCommand, command: []string{"east", "e", "go east", "go e"}},
			{op: opWaitDiffRoom, duration: 2 * time.Second},
			{op: opSayDiffRoom, text: "Good luck."},
			{op: opWaitKill, targetTemplateID: 19},
			{op: opWaitDiffRoom, duration: 1 * time.Second},
			{op: opAlert, text: "Good job. Return to Gwaarhar to finish the quest."},
			{op: opWaitForReturn},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Hey! You did it!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Now, go back to the training ground. Fight whatever you want, but this time..."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Your motive is to gain 1 level. You already gained some exp after fighting the Small Dummy."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "Take a look at the `profile` or `pf` command to check your progress, along with a lot of other information! I'll wait."},
			{op: opWaitCommand, command: []string{"profile", "pf"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Great. Now that you know where you stand, go, break a leg. Come back when you gain a level."},
			{op: opWaitForLvl},
			{op: opWaitDiffRoom, duration: 1 * time.Second},
			{op: opAlert, text: "Good job. Return to Gwaarhar to finish the quest."},
			{op: opWaitForReturn},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Hey! You did it!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "As you saw, when you gain a level, you also gain 5 'trains' along with it."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Each train allows you to increase any one of your 5 stats by 1."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "You can see your current stats by using the `profile` or `pf` command I taught you earlier."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "This is about all I can teach you, kid. It's time for your next class."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Then... Once you're ready, go `west` twice and `talk` to your next instructor."},
			{op: opWaitCommand, command: []string{"west", "w", "go west", "go w"}},
			{op: opWaitDiffRoom, duration: 2 * time.Second},
			{op: opSayDiffRoom, text: "Until next time, kid."},
		},
	},
	3: &NPC{
		name:       "Bulgan, the Sturdy",
		locationID: 112,
		script: []NPCScriptStage{
			{op: opSay, text: "Harr. Harr. Oh. You're here."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Well... You're rather small..."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "I see you've gained a level before coming here. Atleast you did that."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Harr. Harr."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Anyway."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Let's get you in shape, shall we?"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "As Gwaarhar must've told you, you gain trains when you level up."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Then, these trains can be used to improve your stats. We're gonna do just that."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "You have 6 different options to spend your trains."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "`str`, `dex`, `agi`, `stam`, `int` and finally, `hp`."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "For now, I would suggest not worrying about `stam` and `int`."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Just focus on `str`, `agi` and `hp` as these three directly affect combat."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Harr. Let me explain."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Strength (`str`) makes you deal more damage to the enemy."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Agility (`agi`) increases the chance of you dodging your enemy's attacks and increases the chance of you landing a hit on the enemy."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "Health Points (`hp`) increases your maximum health, hence allowing you to tank stronger enemies!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Additionally, dexterity (`dex`) allows you to carry more items. You'll learn about it soon."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "To train, you use the command `train`, with the following syntax: `train <quantity> <stat_name>`"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Your turn. Train your `str` stat once. I'll wait. (Quantity would be 1, name would be `str` or `strength`!)"},
			{op: opWait, duration: 6 * time.Second},
			{op: opWaitCommand, command: []string{"train 1 str", "train 1 strength"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Arr Harr. You already look a little stronger. Good job kid."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "On that note, I'd like to tell you more about items."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "As you can see in your `profile` card, or with the `worn` command..."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Items can be equipped to a total of 6 slots."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "But, that's only for items that are of the type Equipment."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "There are also Consumable type items, which give you boosts, or have a one-time effect!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "For example, the Cookie item heals 10 hp and then is used up!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "To check the type of the item, as well as it's effect, stats, et cetera..."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "You can use the `examine` command!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "First, check your inventory, then use `examine <item>`, either index it by keyword or using the number system."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "To consume Consumable items, you can use the `use <item>` command."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Well then, that's all I had for you."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Go `east` once, `north` once, then `east` again to talk to Cedriah about her boring economics..."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Or, you can just skip her and read `help merchants` instead. Wink."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Happy exploring, kiddo."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "See ya."},
		},
	},
	4: &NPC{
		name:       "Cedriah, the Enchantress",
		locationID: 114,
		script: []NPCScriptStage{
			{op: opSay, text: "Hey~ Nice to meet you."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "My name is Cedriah, would you like to learn about currency and merchants~"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Ahem."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "The currency used in the world of Forlorn is gold."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "You can get gold by defeating entities or selling items."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Merchants are just entities that also sell items."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "All merchants have a cyan-colored 'M' next to their name."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Did you notice? There's a merchant here in this room right now!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Take a `look`, I'll wait. (Type 'look')"},
			{op: opWaitCommand, command: []string{"l", "look"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "All merchants have a list of items that they deal in (either buy or sell)!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "All merchants accept only gold."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "To take a look at a merchant's item list..."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "You use the `list <merchant>` command, without the <>, where `merchant` is a word from the merchant entity's name."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "It's just like you learnt at the start, except numbers don't work on merchants!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Since merchants are just entities, you can fight them! But, that's not recommended..."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "You see, it is said that every merchant was once a brave adventurer. Having seen a lot of this world..."},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "They've become insanely strong. Normal people like us can't fathom their strength."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "But perhaps, one day... You might be strong enough to beat one of them."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Regardless, then, go on, try it."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Do `list dummy` or `list merchant`, whichever you prefer!"},
			{op: opWaitCommand, command: []string{"list dummy", "list merchant"}},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Now that you've seen the items the merchant has to sell, it's time to buy something!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "You can see the command to buy an item right below the merchant's list."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Now all you have to do is insert the ID of the item you wish to buy in place of <id> in the command!"},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "You can find the item ID in the first column of the list."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Then, go on. Buy the Wooden Stick that the merchant is selling! ('buy 20 merchant')"},
			{op: opWaitCommand, command: []string{"buy 20 dummy", "buy 20 merchant"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Hey! You did it! Now, let's try selling."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "When you try to sell, it checks if you have the required item in your inventory."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "If you do, then that item is sold and you get the money!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Use the same command, but replace buy with sell this time! ('sell 20 dummy')"},
			{op: opWait, duration: 3 * time.Second},
			{op: opWaitCommand, command: []string{"sell 20 dummy", "sell 20 merchant"}},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Notice that the item you bought and sold is completely free (0 gold)! Real items aren't like this, like the cookie that you see!"},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "Now you know how to deal with merchants and use gold!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Gold is also used in other places, for example, creating a clan!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Speaking of clans, your next teacher must be quite eager to teach you about the social part of Forlorn!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "It's about time you went to him."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Go `west` twice to meet him!"},
			{op: opWaitCommand, command: []string{"west", "go west", "w", "go w"}},
			{op: opWaitDiffRoom, duration: 2 * time.Second},
			{op: opSayDiffRoom, text: "See you later~"},
		},
	},
	5: &NPC{
		name:       "Gungnir, the Tribal Chief",
		locationID: 115,
		script: []NPCScriptStage{
			{op: opSay, text: "Ah! New kid, you're finally here!"},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "I'm Gungnir, and I'm gonna teach you about the social aspect of Forlorn!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Firstly, as you've learned before, you can identify players with the '*' before their names."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "You can fight with them, but you can also look at their profiles!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "The `user <player_name>` command shows you other players' profiles!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "This command shows you a user's `profile` exactly like you'd see your own!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Next up: You can message other players as well!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Use the `sayto <player_name> <msg>` (without the <>) command to do this."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Finally, Forlorn also has a clans system!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Clans are basically groups you can create/join, to enjoy the game with your clanmates!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "Clans give you clan tags which are 4-lettered tags that appear next to your name whenever other players see you!"},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "They're visible in your profile as well!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "You can talk to ALL of your clanmates at the same time with the `clan say <msg>` command!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "The clan helpfile has all the information about clans!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Forgive me for dumping a huge helpfile on you, but the clans feature is too big to be explained!"},
			{op: opWait, duration: 6 * time.Second},
			{op: opSay, text: "Please take a look at `help clans`! I'll wait."},
			{op: opWaitCommand, command: []string{"help clans", "h clans"}},
			{op: opWait, duration: 1 * time.Second},
			{op: opSay, text: "Take your time and read it! I'll be waiting."},
			{op: opWait, duration: 9 * time.Second},
			{op: opSay, text: "That's all I had to teach you!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "If you want, you should head to the library."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "Go `east` then `north` and you'll find your way to the library."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Good luck out there, kid."},
		},
	},
	6: &NPC{
		name:       "Arbara, the Librarian",
		locationID: 117,
		script: []NPCScriptStage{
			{op: opIfClientWeb},
			{op: opSay, text: "Oh! Hey there."},
			{op: opWait, duration: 2 * time.Second},
			{op: opSay, text: "I'm Arbara and I would like to tell you about the features of the web client. (The website that you're on right now!)"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "On the right, the side bar shows all the directions you can move in."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "It also shows your gold, your level, your exp, and it even has a exp bar!"},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "The server sends special strings which the client sees and uses to update the UI."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "Cool, right? Well, not gonna bore you with all that nerdy stuff..."},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "The two bars that are currently grayed out only activate when you enter combat, and reset once you defeat the enemy or die!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "They reflect the real-time fight data."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Then, there's also a map button that opens a map of the entirety of Forlorn!"},
			{op: opWait, duration: 4 * time.Second},
			{op: opSay, text: "You can use the map to navigate properly!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Finally, we have the settings."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "You can access the settings by clicking on the settings button in the side bar!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "The settings page allows you add 'aliases' which the client remembers."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "An alias is already added for you as an examaple."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Basically, as you can see, aliases are used to shorten commands."},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "The example alias turns the `help flavortown` command into `hf`."},
			{op: opWait, duration: 5 * time.Second},
			{op: opSay, text: "So, if you were to type `hf` in your input bar, the client automatically turns it into `help flavortown`!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "Upgrades for this feature are a work in progress!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "That's it!"},
			{op: opWait, duration: 3 * time.Second},
			{op: opSay, text: "I'm gonna keep this short and sweet. So, cya out there!"},
		},
	},
}

func defineNPCs(world *World) {
	world.npcs = npcs
}

func runNPCScript(char *Character) {
	for char.activeConvo != nil && char.activeConvo.stage < len(char.activeConvo.NPC.script) {
		stage := char.activeConvo.NPC.script[char.activeConvo.stage]

		switch stage.op {
		case opSay:
			if char.locationID == char.activeConvo.NPC.locationID {
				msg := "\x1b[2K\r  " + color(char.conn, "red", "tp") + "! " + color(char.conn, "yellow", "tp") + char.activeConvo.NPC.name + " " + color(char.conn, "reset", "reset") + "says, \""
				char.conn.store.Write([]byte(msg + color(char.conn, "cyan", "tp") + stage.text + color(char.conn, "reset", "reset") + "\"\n\n> "))
				char.activeConvo.stage++
			} else {
				return
			}

		case opSayDiffRoom:
			msg := "\x1b[2K\r  " + color(char.conn, "red", "tp") + "! " + color(char.conn, "yellow", "tp") + char.activeConvo.NPC.name + " " + color(char.conn, "reset", "reset") + "says, \""
			char.conn.store.Write([]byte(msg + color(char.conn, "cyan", "tp") + stage.text + color(char.conn, "reset", "reset") + "\"\n\n> "))
			char.activeConvo.stage++

		case opAlert:
			msg := "\x1b[2K\r  " + color(char.conn, "red", "tp") + "!!! " + color(char.conn, "reset", "reset")
			char.conn.store.Write([]byte(msg + color(char.conn, "cyan", "tp") + stage.text + color(char.conn, "reset", "reset") + "\n\n> "))
			char.activeConvo.stage++

		case opWait:
			if char.locationID == char.activeConvo.NPC.locationID {
				time.Sleep(stage.duration)
				char.activeConvo.stage++
			} else {
				return
			}

		case opWaitDiffRoom:
			time.Sleep(stage.duration)
			char.activeConvo.stage++

		case opWaitCommand:
			return

		case opWaitForLvl:
			return

		case opWaitKill:
			return

		case opWaitForReturn:
			return

		case opIfClientWeb:
			if char.conn.isClientWeb {
				char.activeConvo.stage++
			} else {
				msg := "\x1b[2K\r  " + color(char.conn, "red", "tp") + "! " + color(char.conn, "yellow", "tp") + char.activeConvo.NPC.name + " " + color(char.conn, "reset", "reset") + "says, \""
				char.conn.store.Write([]byte(msg + color(char.conn, "cyan", "tp") + "You're not on the web client!" + color(char.conn, "reset", "reset") + "\"\n\n> "))
				return
			}
		}
	}
}

func checkCmd(char *Character, cmd string) {
	if char.activeConvo == nil {
		return
	}

	if char.activeConvo.stage >= len(char.activeConvo.NPC.script) {
		return
	}

	stage := char.activeConvo.NPC.script[char.activeConvo.stage]

	if stage.op == opWaitCommand && slices.Contains(stage.command, cmd) {
		char.activeConvo.stage++
		go runNPCScript(char)
	}
}

func checkKill(char *Character, templateID int) {
	if char.activeConvo == nil {
		return
	}

	stage := char.activeConvo.NPC.script[char.activeConvo.stage]

	if char.activeConvo.stage >= len(char.activeConvo.NPC.script) {
		return
	}

	if stage.op != opWaitKill {
		return
	}

	if stage.targetTemplateID == templateID {
		char.activeConvo.stage++
		checkReturn(char)
		go runNPCScript(char)
	}
}

func checkReturn(char *Character) {
	if char.activeConvo == nil {
		return
	}

	if char.activeConvo.stage >= len(char.activeConvo.NPC.script) {
		return
	}

	stage := char.activeConvo.NPC.script[char.activeConvo.stage]

	if stage.op != opWaitForReturn {
		return
	}

	if char.locationID == char.activeConvo.NPC.locationID {
		char.activeConvo.stage++
		go runNPCScript(char)
	}
}

func gainLvl(char *Character) {
	if char.activeConvo == nil {
		return
	}

	if char.activeConvo.stage >= len(char.activeConvo.NPC.script) {
		return
	}

	stage := char.activeConvo.NPC.script[char.activeConvo.stage]

	if stage.op != opWaitForLvl {
		return
	}

	char.activeConvo.stage++
	go runNPCScript(char)
}
