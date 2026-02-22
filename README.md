<div align="center">
  <h1>ForlornMUD</h1>
  <img src="https://res.cloudinary.com/dp7g5aflo/image/upload/v1771271258/Untitled197_20260216231412_rdbuqx.png">
  <h6><a href="https://youtube.com/@vishkaun">Click here for art credits!</a></h6>
  
  ![Time Tracking](https://img.shields.io/badge/ForlornMUD-87h%2014m-critical?logo=neovim&style=plastic)
  ![GitHub Stars](https://img.shields.io/github/stars/Snxhit/ForlornMUD?style=plastic)
  ![GitHub Forks](https://img.shields.io/github/forks/Snxhit/ForlornMUD?style=plastic)
  ![GitHub Issues](https://img.shields.io/github/issues/Snxhit/ForlornMUD?style=plastic)
  
</div>

---

## Index

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Usage](#usage)
  - [For Users](#for-users) <- Users click here!
  - [Self-Hosting](#self-hosting)
- [Related Projects](#related-projects)
- [Screenshots](#screenshots) <- For instructions video
- [What I Learned](#what-i-learned)
- [About Me](#about-me)

---

## Overview

This project was created in [FlavorTown](https://flavortown.hackclub.com/)!<br>
**Includes a Flavortown map with over 55 rooms!**

**Project Highlights**:
- Real-time multiplayer MUD engine with web-based and terminal clients!
- Dynamic world with entity, item spawners, merchants, and a flexible stats, leveling system!
- Thought out combat mechanics, including PvP, PvE, with looting and leveling!
- Automatic database setup and persistent world state.

---

## Features

- Flexible entity & item system using spawners for dynamic world population.
- Stat system for players and entities, including item buffs and modifiers.
- Combat system with scaling, random factors, dodging mechanics, and looting and item drops.
- Merchant entities for buying/selling items, simulating an in-game economy.
- Room-based world navigation with cardinal directions and descriptive areas, like in classic MUDs!
- Persistent player profiles, inventory, and stats.
- [Web client](https://snxhit.me/ForlornClientWeb) and terminal (Netcat/Ncat) support for gameplay.
- Helpfile system for in-game tutorials, topics & command documentation.
- PvP combat with coin looting and respawning.
- Periodic world ticks for health regen, spawner management & combat.
- Fully optional colors, beautiful unicodes for a modern terminal look and maximum compatibility!
- Profile, and help cards for enhanced terminal readability!
- Automatic database initialization and world setup on first run (Self-Hosting)

---

## Tech Stack

- **Engine** Written in Golang
- **Database** SQLite
- **Networking** TCP for real-time connections (WebSockets for web client)
- **UI** ASCII art and colored terminal output
- **Authentication** Username + hashed passwords
- **Deployment** VPS on Azure

---

## Architecture
```
- main.go: Entry point, owns all objects, handles networking, client connections, and world initialization.
- commands.go: Processes player input and executes game logic.
- combat.go: Combat mechanics for PvE and PvP.
- helpfiles.go: Defines, manages and sets up in-game helpfiles.
- ticks.go: Manages periodic world updates.
- utils.go: All sorts of utility functions.
- db.go: Contains loader for loading template dir.
- template/
  - 001_dbschema.sql: The basic schema of db.
  - 002_rooms.sql: Sets up rooms.
  - 003_spawners.sql: Sets up item and entity spawners.
  - 004_items.sql: Sets up items, item modifiers, item effects.
  - 005_entities.sql: Sets up entities and their drops
  - 006_merchants.sql: Sets up merchant entities, merchants table, and their lists.
  - flavortown_map.txt: an AI generated (poor quality) map of the Flavortown map included with the base engine.
```

---

## Usage

### For Users

**Follow these steps to play the instance hosted by Snxhit:**

Server status: OFFLINE

1. **Web Client**

   Visit the web client [here](https://snxhit.me/ForlornClientWeb)!

2. **Connect**

   A connection will instantly be established to the server. (If it's online)

3. **Login or Register**

   You will be prompted for a username and a password.

#### Or, alternatively:

1. You may connect to the instance by following the steps mentioned in the [Self-Hosting](#self-hosting) section,
    - But, replace Netcat's destination from `localhost` to `forlorn.snxhit.me`!

**Tip:**
For access to UI elements and a better gameplay experience, stick to the web client!

#### In-Game Help
- To get started with the game, use `help newplayer` for a guided tutorial!

---

### Self-Hosting

**If you want to run your own instance, here's how!**

#### Prerequisites

- Tested on `Go v1.25.7` (If compiling from source)
- `Git v2.50.0+` (If compiling from source)

#### Installation Steps

1. **Get the binary**
  Download the official released binaries from the [releases](https://github.com/Snxhit/ForlornMUD/releases) page,

    1.1 **Compile from source**
      Follow the below sequence of commands to alternatively compile from source instead of using release binaries.
        ```bash
        git clone https://github.com/Snxhit/ForlornMUD.git
        cd ForlornMUD
        go build .
        ```

2. **Setting up the database**
    1. Launch the `ForlornMUD` executable that you get from the release or by compiling.
    2. Wait till the console output says that the server successfully started.
    3. The server, on startup, automatically generates the database.

3. **Access your instance**

    - **Windows**:
      1. Download [Ncat](https://nmap.org/ncat/) from here
      2. Run `ncat locahost 8899`

    - **Linux**:
      1. Install Netcat or Ncat from your distro's package manager (apt, pacman, dnf, etc.) if not already installed
      2. Run `ncat localhost 8899` (Try `nc localhost 8899` if `ncat` doesn't work.)

    - **MacOS**:
      1. Usually, netcat is preinstalled on most MacOS systems
      2. Run `nc localhost 8899`

4. **Customizing your instance**
    - You start with the Flavortown map template, which includes a huge area, but not many entities and items.
    - You can customize anything to you want, add new items, entities, effects, spawners, rooms, et cetera.
    - After you're done customizing:
        - Delete `game.db` and,
        - Run the server once again,
        - Viola!
    - **You now have your very own customized instance of ForlornMUD!**

Support for custom instances on the web client is a work in progress!

Consider deploying your instance on a VPS instead of port forwarding locally, or use free tunnels! ([Cloudflared Tunnel](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/downloads/) or [Pinggy](https://pinggy.io/))

---

## Related Projects
- [ForlornClientWeb](https://github.com/Snxhit/ForlornClientWeb) <br>
A web-based client for ForlornMUD, built with xterm.js for a modern MUD experience by implementing reactive UI!<br>
It connects to the game via WebSockets, making it accessible from any browser.

- [ForlornWebsocketBridge](https://github.com/Snxhit/ForlornWebsocketBridge) <br>
A bridge that translates WebSocket connections from the web client into TCP connections for the MUD server.<br>
This enables playing from the browser without requiring direct TCP support.

---

## Screenshots
- Preview<br>
  [Preview](https://res.cloudinary.com/dp7g5aflo/video/upload/v1771735143/ForlornPreview1_ithyhp.mp4)

- **Combat system**<br>
Item drops, coins, exp, levels, and more!

  ![Combat System](https://res.cloudinary.com/dp7g5aflo/image/upload/v1771233402/2026-02-16_14-39_1_dxk4oq.png)<br>
Defeat, respawning system

  ![Combat System 2](https://res.cloudinary.com/dp7g5aflo/image/upload/v1771233402/2026-02-16_14-38_tw4ceh.png)

- **Merchants**<br>
Listing, buying and selling

  ![Merchants](https://res.cloudinary.com/dp7g5aflo/image/upload/v1771233402/2026-02-16_14-39_kxfarl.png)

- **Profile**

  ![Profile](https://res.cloudinary.com/dp7g5aflo/image/upload/v1771233402/2026-02-16_14-45_svk09j.png)

- **World exploration**

  ![Exploration](https://res.cloudinary.com/dp7g5aflo/image/upload/v1771233402/2026-02-16_14-40_gdf9uc.png)

- **Training system**

  ![Training](https://res.cloudinary.com/dp7g5aflo/image/upload/v1771233402/2026-02-16_14-45_1_zav7i2.png)


---

## What I Learned

Building this project taught me quite a lot:

### Technical Skills
- **Golang**: This was my first time using Go, and I loved it!
- **SQL**: Although it was long overdue, I finally learnt SQL and DBMS from scratch!

### Problem Solving
- Managing time working on projects because this time it actually mattered.
- Managing larger codebases than I'm used to.
- Structuring servers like this with a lot of reusable logic (Learnt this the hard way.)
- Although I tried my best to write good quality code, I still see a lot of design choices biting me. So, I also learned what **NOT** to do!
- A lot about concurrency and using goroutines with Go!
- Actually learning how to write READMEs and details.
- Properly using a VCS!

### Personal
1. Working on something till I'm satisfied with it.
2. Managing scope creep.
3. Keeping track of work status instead of winging it and working on whatever, whenever.
4. Building consistent and efficient work habits.

---

## About Me

**Snxhit**<br>
I'm passionate about building interactive multiplayer games and learning new languages.<br>
**ForlornMUD** is my very first major Go project (There will definitely be more!), and it reflects my love for simulated economies, text based multiplayer games alike!

- GitHub: [@Snxhit](https://github.com/Snxhit)
- LinkedIn: Snxhit (Unavailable for now)
- Email: [developer@snxhit.me](mailto:developer@snxhit.me)

---

## Project Status

**Version**: v0.1.3<br>
**Latest Release Version**: v0.1.3<br>
**Time Spent**: 87 Hours 14 Minutes<br>
**Status**: Actively developed, not stable yet!<br>

---

<div align="center">

  **Made with ❤️**

  If you found this project interesting, consider giving it a star :D

  [Report a Bug](https://github.com/Snxhit/ForlornMUD/issues) or [Suggest a new Feature](https://github.com/Snxhit/ForlornMUD/issues)

</div>
