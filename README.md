# ntkpr

![Demo](assets/screenshot.png)
(Screenshot for ntkpr v0.3)

## Important Notices

A straight forward way to install and use this app is cloning this repo, then go to /ntkpr/scripts folder. build.sh will build the executable, and add_to_path.sh will add it to system path. You might wanna configure the shell commands a bit to fit it to your computer.

## Introduction

`ntkpr` is a terminal journal management tool that provides TUI interface for note taking and design (more to come) for journal management.

I enjoy writing down plans, thoughts and ideas during daily work and life, and I also like the general style of terminal applications. `ntkpr` is an attempt to digitalize and automate the whole workflow, from which I hope to explore different patterns of journal taking and managing.

## Development

- Stable version: v0.2.0. Simple, 1-layer note managing.
- developing: v0.3.0. Thread-Branch-Note version-control style structure, more advanced UI.

## Installation Guide

Currently `ntkpr` runs on Linux and macOS.

### macOS

```bash
cd ~
curl -L https://github.com/haochend413/ntkpr/releases/latest/download/ntkpr_darwin_arm64 \
  -o ntkpr

chmod +x ntkpr
sudo mv ntkpr /usr/local/bin/
```

### Linux

```bash
cd ~
curl -L https://github.com/haochend413/ntkpr/releases/latest/download/ntkpr_linux_amd64 \
  -o ntkpr

chmod +x ntkpr
sudo mv ntkpr /usr/local/bin/
```

### Windows

Windows is currently not supported. You can run `ntkpr` on WSL if you're using Windows.

### Local Build

You can also clone the git repo and build it locally with `go build -o ntkpr`. This will allow you to try the locally hosted GUI interface and the LLM agent. This should work on any OS.

## Keymaps

### Global Keymaps

- `Ctrl+c`: quit the application.
- `Tab`: switch focusing window.
- `Ctrl+q`: sync with database and flush the caches.

### Table Keymaps

- `n`: create new note.
- `Ctrl+d`: delete current note.
- `Ctrl+z`: undo last deletion.
- `Ctrl+h`: highlight current note.
- `Ctrl+p`: make current note private (invisible on GUI).
- `A`: switch to Default context.
- `R`: switch to Recent context.
- `S`: open up search bar.
- `enter/Tab`: go to text area.

### Textarea Keymaps

- `Ctrl+s`: save current note content.
- Other shortcuts included by default.

## Commands

```bash
ntkpr # launch TUI
```

### GUI commands

These now only works if you clone the git repo and build/run it locally.

```bash
ntkpr gui # launch GUI
```

```bash
ntkpr export # sync GUI data with database
```

### Data commands

```bash
ntkpr backup [path/to/backup/folder] # backup the config, state and your database to a folder. Default to cwd.
```

## Program Config

Program configs are stored by default in:

```bash
"~/Library/Application Support/ntkpr/" # macOS
"~/.local/state/ntkpr/" # Linux
```
