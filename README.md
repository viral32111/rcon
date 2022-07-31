# RCON

This is a command-line RCON client for [Minecraft](https://minecraft.net) and the [Source Engine](https://wikipedia.org/wiki/Source_(game_engine)).

The [Minecraft protocol](https://wiki.vg/RCON) is an implementation of the [Source Engine protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol), with a few minor bug fixes.

## Background

Since I run game servers for my community and friends, I require a reliable way to remotely control them via command-line on the host server. There are many tools available to do this already, such as [mcrcon](https://github.com/Tiiffi/mcrcon), which is what I used before making this project, but I wanted to make my own once I started needing one for the Source Engine too, as I did not want to have multiple utilities for each game.

I decided to use Go because it is perfect for these kinds of small single-executable utilities that need to work on multiple platforms, which in my case is Linux (glibc & musl) and Windows. Go also has a huge standard library, which makes development easier as third-party libraries do not have to be imported and managed.

## Usage

The program is simple to use, just specify a protocol, server IP address, password and a command then the server response will be displayed.

### Flags

These flags can be prefixed with either a single hyphen (`-`) or a double hyphen (`--`), the choice is yours.

Use the `-help` flag to show a list of these flags with descriptions and default values.

#### Required

Exactly one protocol must be used.

* `-minecraft` to use the [Minecraft protocol](https://wiki.vg/RCON), and set `-port` default to `25575`.
* `-sourceengine` to use the [Source Engine protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol), and set `-port` default to `27015`.

#### Optional

* `-address <string>` to specify the remote server's IPv4 address (e.g. `-address 192.168.0.5`, defaults to `127.0.0.1`).
* `-port <number>` to specify the remote server's port number (e.g. `-port 27020`, defaults to the protocol is in use).
* `-password <string>` to specify the remote console password (e.g. `-password verySecurePassword123`, defaults to an empty string),

## Arguments

All arguments that are not flags will be combined to become the command to execute.

The flags can be provided in any order, but the arguments (the command to execute) must come last.

### Examples

A Garry's Mod server at `192.168.0.5` using the default port `27015`:

```
$ rcon -sourceengine -address 192.168.0.5 -password verySecurePassword123 status
hostname: Example Server
version : 2022.06.08/24 8606 insecure
udp/ip  : 192.168.0.5:27015
map     : gm_construct at: 0 x, 0 y, 0 z
players : 0 (10 max)

# userid name                uniqueid            connected ping loss state  adr
```

A Source Engine server at `127.0.0.1` using the custom port `27020`:

```
$ rcon -sourceengine -port 27020 -password superRealPassword567 sv_cheats 1
L 07/29/2022 - 20:59:32: server_cvar: "sv_cheats" "1"
```

A Source Engine server at `127.0.0.1` using the default port `27015`:

```
$ rcon --password aw3s0meP4ssw0rd --sourceengine addip 60 192.168.0.100
L 07/29/2022 - 21:00:54: Addip: "<><><>" was banned by IP "for 60.00 minutes" by "Console" (IP "192.168.0.100")
```

A Minecraft server at `192.168.0.10` using the default port `25575`:

```
$ rcon -address 192.168.0.10 -minecraft -password reallyG00dPassword list
There are 0 of a max of 20 players online:
```

## To-Do List

* Check if request/response packet identifiers match.
* Multi-packet/fragmented responses.
* More error handling.
* Environment variables as fallback for flags & arguments.
  * `RCON_ADDRESS=192.168.0.5`
  * `RCON_PORT=27015`
  * `RCON_PASSWORD=abcxyz`
  * `RCON_COMMAND=status`
* Run multiple commands with a specified delay between each one (mcrcon `-w <seconds> command1 command2...` equivalent)

## License

Copyright (C) 2022 [viral32111](https://viral32111.com).

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see https://www.gnu.org/licenses.
