<p align="center">
 <img alt="kfdsl logo" src="https://raw.githubusercontent.com/K4rian/kfdsl/refs/heads/assets/icons/logo-kfdsl.svg" width="25%" align="center">
</p>

**KFDSL** is a custom launcher for the **Killing Floor Dedicated Server**.<br>
It automates server setup, updates, and management by integrating with **SteamCMD** and supporting **KFPatcher** and **MutLoader** out of the box.<br>

---

## Features
- **SteamCMD Integration**: Install, update, and validate server files effortlessly using SteamCMD.
- **Configuration Management**: Automatically updates server and mods configuration files.
- **Mod Support**: Built-in support for **KFPatcher** and **MutLoader**.
- **Process Management**: Start, stop and automatically restart the server process.

## Getting started
### Prerequisites
- A **Linux** environment with **SteamCMD** installed.
- A secondary **Steam Account** with **Steam Guard disabled** (see <a href="#environment-variables">Environment variables</a>).
- The following ports to be opened:
  - 7707 (UDP)
  - 7708 (UDP)
  - 28852 (UDP/TCP)
  - 8075 (TCP)
  - 20560 (TCP/UDP)

### Installation
#### Using Docker (✅ **Recommended**)
- See the __[docker-killingfloor][X]__ repository.

#### Manual
- __[Download the latest build][X]__ for your platform.
- Extract the archive.
- Run the launcher `./kfdsl` with the desired flags and arguments, see the section below.


## Environment variables
To download and update the server files, it is required to provide both a valid Steam username and password.  
It is **strongly recommended** to create a secondary Steam account **without Steam Guard** specifically for the server.  
Using your **main Steam account is NOT recommended**.

The following environment variables have to be set for the launcher to work:

Variable               | Default Value                     | Description
---                    | ---                               | ---
STEAMACC_USERNAME      | `anonymous`                       | Steam account username. 
STEAMACC_PASSWORD      | *(empty)*                         | Steam account password.

## Flags and Arguments
<details>
<summary>Click to expand</summary>

Flag                     | Default Argument Value          | Description
---                      | ---                             | ---
--config                 | `KillingFloor.ini`              | Server configuration file. 
--servername             | `KF Server`                     | Name of the server. 
--shortname              | `KFS`                           | Short name (alias) for the server. 
--port                   | `7707`                          | Game server port. 
--webadminport           | `8075`                          | Web admin panel port. 
--gamespyport            | `7717`                          | GameSpy query port. 
--gamemode               | `survival`                      | Game mode (`survival, objective, toymaster`). 
--map                    | `KF-BioticsLab`                 | Map to start the server on. 
--difficulty             | `hard`                          | Game difficulty level (`easy, normal, hard, suicidal, hell`). 
--length                 | `medium`                        | Game length (`short, medium, long`). 
--friendlyfire           | `0.0`                           | Friendly fire multiplier (`0.0` = off, `1.0` = full damage). 
--maxplayers             | `6`                             | Maximum number of players. 
--maxspectators          | `6`                             | Maximum number of spectators. 
--password               | *(empty)*                       | Server Password (`empty` = no password). 
--region                 | `1`                             | Server region. 
--adminname              | *(empty)*                       | Administrator name. 
--adminmail              | *(empty)*                       | Administrator email address. 
--adminpassword          | *(empty)*                       | Administrator password. 
--motd                   | *(empty)*                       | Message of the day. 
--specimentype           | `default`                       | ZEDs type (`default, summer, halloween, christmas`). 
--mutators               | *(empty)*                       | Command-line mutators list. 
--servermutators         | *(empty)*                       | Server-side mutators list (`ServerActors`). 
--redirecturl            | *(empty)*                       | URL for fast download redirection. 
--maplist                | `all`                           | List of available maps for the current game separated by a comma (`all` = all available maps). 
--webadmin               | `unset` *(disabled)*            | Enable the web admin panel. 
--mapvote                | `unset` *(disabled)*            | Enable map voting. 
--mapvote-repeatlimit    | `1`                             | Number of maps to be played before a repeat. 
--adminpause             | `unset` *(disabled)*            | Allow administrators to pause the game. 
--noweaponthrow          | `unset` *(disabled)*            | Prevent weapons from being thrown on the ground. 
--noweaponshake          | `unset` *(disabled)*            | Disable weapon shake effect. 
--thirdperson            | `unset` *(disabled)*            | Enable third-person view (F4). 
--lowgore                | `unset` *(disabled)*            | Disable gore system (no dismemberment). 
--uncap                  | `unset` *(disabled)*            | Uncap the framerate (requires client-side tweaks too). 
--unsecure               | `unset` *(disabled)*            | Start the server without Valve Anti-Cheat (VAC). 
--nosteam                | `unset` *(disabled)*            | Bypass SteamCMD and start the server immediately. 
--novalidate             | `unset` *(disabled)*            | Skip server files integrity check. 
--autorestart            | `unset` *(disabled)*            | Automatically restart the server if it crashes. 
--mutloader              | `unset` *(disabled)*            | Enable MutLoader (inline mutator). 
--kfpatcher              | `unset` *(disabled)*            | Enable KFPatcher (server mutator). 
--hideperks              | `unset` *(disabled)*            | KFPatcher: Hide perks. 
--nozedtime              | `unset` *(disabled)*            | KFPatcher: Disable ZED Time (slow-motion). 
--buyeverywhere          | `unset` *(disabled)*            | KFPatcher: Allow buying weapons anywhere. 
--alltraders             | `unset` *(disabled)*            | KFPatcher: Keep all traders open. 
--alltraders-message     | `"^wAll traders are ^ropen^w!"` | KFPatcher: Message displayed when all traders are open. 
--log-to-file            | `unset` *(disabled)*            | Enable logging to a file. 
--log-level              | `info`                          | Logging level (`info, debug, warn, error`). 
--log-file               | `./kfdsl.log`                   | Path to the log file. 
--log-file-format        | `text`                          | Format of the log file (`text, json`). 
--log-max-size           | `10`                            | Maximum log file size in MB. 
--log-max-backups        | `5`                             | Maximum number of old log files to retain. 
--log-max-age            | `28`                            | Maximum log file age in days. 
--steamcmd-root          | `$HOME/steamcmd`                | SteamCMD root directory.
--steamcmd-appinstalldir | `$HOME/gameserver`              | Server root directory.

> **All flags can also be set using environment variables.**<br>
> For example, `--config` can be set using the `KF_CONFIG` environment variable.<br>
> **Note**: All environment variables must be prefixed with `KF_`, except for `STEAMCMD_ROOT` and `STEAMCMD_APPINSTALLDIR`, which do not use a prefix.
</details>

## Usage
> *In all examples, the required `environment variables` are stored in the `kfdsl.env` file located in the current working directory.*

__Example 1:__<br>
Run a public `Survival` server with custom `names`, set to `Suicidal` difficulty on a `long-length` game, and starting on `KF-WestLondon`: 
```bash
source kfdsl.env && ./kfdsl \
  --servername "KF Server [Suidical] [Long]" \
  --shortname "KFS" \
  --map "KF-WestLondon" \
  --difficulty "suicidal" \
  --length "long"
```

__Example 2:__<br>
Run a password-protected server in `Objective` mode, with `map voting` enabled, set to `Hard` difficulty on a `medium-length` game, and starting on `KFO-Steamland`:
```bash
source kfdsl.env && ./kfdsl \
  --servername "KF Server [Objective] [Hard] [Medium]" \
  --shortname "OKFS" \
  --gamemode "objective" \
  --map "KFO-Steamland" \
  --difficulty "hard" \
  --length "medium" \
  --password "<16_CHARACTERS_MAX_PASSWORD>" \
  --mapvote
```

__Example 3:__<br />
Run a public `Toy Master` server using a custom `configuration file`, `map voting` and `web admin panel` enabled, set to `Hell on Earth` difficulty on a `short-length` game, and using a custom `server directory`:
```bash
source kfdsl.env && ./kfdsl \
  --config "ToyGame.ini" \
  --servername "KF Server [Toy Master] [HoE] [Short]" \
  --shortname "TMKFS" \
  --gamemode "toymaster" \
  --map "TOY-DevilsDollhouse" \
  --difficulty "hell" \
  --length "short" \
  --adminpassword "<16_CHARACTERS_MAX_PASSWORD>" \
  --webadmin \
  --mapvote \
  --steamcmd-appinstalldir "/opt/kfserver"
```

## Building
Building is done with the `go` tool. If you have setup your `GOPATH` correctly, the following should work:
```bash
go get github.com/k4rian/kfdsl
go build -ldflags "-w -s" github.com/k4rian/kfdsl
```

## See also
- **[docker-steamcmd][5]**: A Docker image used to deploy a KF Dedicated Server with KFDSL included.
- **[kfrs][6]**: A secure HTTP redirect server for the KF Dedicated Server, written in Go.

## License
[MIT][7]

<p align="center"><small>Made with 💀 for the Killing Floor community.</small></p>

[1]: https://github.com/K4rian/kfdsl "Killing Floor Dedicated Server Launcher (KFDSL)"
[2]: https://hub.docker.com/_/debian "Debian Docker Image on Docker Hub"
[3]: https://github.com/K4rian/docker-killingfloor/blob/main/Dockerfile "Latest Dockerfile"
[4]: https://github.com/K4rian/docker-killingfloor/tree/main/compose "Compose Files"
[5]: https://github.com/K4rian/docker-steamcmd "SteamCMD Docker Image"
[6]: https://github.com/K4rian/kfrs "Killing Floor Redirect Server (KFRS)"
[7]: https://github.com/K4rian/docker-killingfloor/blob/main/LICENSE