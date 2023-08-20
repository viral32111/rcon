# RCON

[![CI](https://github.com/viral32111/rcon/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/viral32111/rcon/actions/workflows/ci.yml)
[![CodeQL](https://github.com/viral32111/rcon/actions/workflows/codeql.yml/badge.svg)](https://github.com/viral32111/rcon/actions/workflows/codeql.yml)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/viral32111/rcon?label=Latest)
![GitHub repository size](https://img.shields.io/github/repo-size/viral32111/rcon?label=Size)
![GitHub release downloads](https://img.shields.io/github/downloads/viral32111/rcon/total?label=Downloads)
![GitHub commit activity](https://img.shields.io/github/commit-activity/m/viral32111/rcon?label=Commits)

This is a command-line RCON (remote console) client for [Minecraft](https://minecraft.net) and the [Source Engine](https://wikipedia.org/wiki/Source_(game_engine)).

## 📜 Background

I host game servers for my community and friends, thus I require a reliable way to remotely control them over command-line on the host server. There are many tools available to do this already, such as [mcrcon](https://github.com/Tiiffi/mcrcon) which served as an inspiration for this project. However, once I started needing one for the Source Engine too, I felt I should make my own as I did not want to be different multiple utilities for each game.

I decided to create this project in Go as it is perfect for small single-executable utilities that need to work across a variety of platforms. Go has a vast standard library too, eliminating the hastle of downloading and importing third-party libraries.

## 📥 Usage

Download the [latest release](https://github.com/viral32111/rcon/releases/latest) for your platform. There are builds available for Linux and Windows, on 32-bit and 64-bit architectures of x86 and ARM. There are extra Linux builds to accommodate glibc and musl libraries. This should cover the majority of Docker images.

The utility expects, at minimum, a protocol and command to be provided. The server's IP address, password, and more, can be specified using optional flags. The server response is displayed as the output, so long as the connection and authentication was successful.

Each argument will be treated as a separate command, so wrap commands in quotation that contain spaces. For example, `"sv_cheats 1"` would be considered a single command but `sv_cheats 1` would be considered as two commands.

### ⚙️ Flags

There are different protocols implemented. Exactly one must be chosen via a flag:

* `--minecraft` to use the [Minecraft protocol](https://wiki.vg/RCON) and set the default port to `25575`.
* `--sourceengine` to use the [Source Engine protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol), and set the default port to `27015`.

There are additional optional flags for fine-tuning functionality:

* `--address <string>`: The remote server's IPv4 address. Defaults to `127.0.0.1`.
* `--port <number>`: The remote server's port number (e.g. `-port 27020`, defaults to the protocol is in use).
* `--password <string>`: The remote console password (e.g. `-password verySecurePassword123`, defaults to an empty string),
* `--interval <number>`: The time to wait in seconds between sending commands, only useful when multiple commands are specified (defaults to 1 second).

These flags can be prefixed with either a single (`-`) or double (`--`) hyphen.

The flags can be provided in any order, but the arguments (the commands to execute) must come last.

Use the `--help` (`-h`) flag for more information.

## 🖼️ Examples

Viewing the status of a Garry's Mod server at `192.168.0.5` using the default port `27015`:

```
$ rcon -sourceengine -address 192.168.0.5 -password verySecurePassword123 status
hostname: Example Server
version : 2022.06.08/24 8606 insecure
udp/ip  : 192.168.0.5:27015
map     : gm_construct at: 0 x, 0 y, 0 z
players : 0 (10 max)

# userid name                uniqueid            connected ping loss state  adr
```

Enabling cheats on a Team Fortress 2 server at `127.0.0.1` using the custom port `27020`:

```
$ rcon -sourceengine -port 27020 -password superRealPassword567 sv_cheats 1
L 07/29/2022 - 20:59:32: server_cvar: "sv_cheats" "1"
```

Banning an IP address on a Team Fortress 2 server at `127.0.0.1` using the default port `27015`:

```
$ rcon --password aw3s0meP4ssw0rd --sourceengine addip 60 192.168.0.100
L 07/29/2022 - 21:00:54: Addip: "<><><>" was banned by IP "for 60.00 minutes" by "Console" (IP "192.168.0.100")
```

Listing online players on a Minecraft server at `192.168.0.10` using the default port `25575`:

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

## ⚖️ License

Copyright (C) 2022-2023 [viral32111](https://viral32111.com).

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
