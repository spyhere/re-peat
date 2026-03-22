# ↺ re-peat

:construction: Work in progress :construction:

## Description

This is an application that can display sound waves of a given MP3 track, you can put/see/edit markers and able to play from them.

## Motivation

In theaters, philharmonics, and even amateur rehearsals, there is often a struggle to start playback from a specific point in the music. Since music usually serves as the backbone of a performance, and actors or dancers tend to repeat certain sections during rehearsals, they rely on music as a reference. The person responsible for the music then has to locate the correct point in the soundtrack, which is often a hassle.
The goal of this application is to eliminate that hassle by allowing users to mark those points once and then start the music directly from them.


## Quick Start

1. Clone the repo
2. Make sure you have `Go` toolchain installed, if you don't have it then check out [this](https://go.dev/doc/install)
3. Inside the cloned repo run `go mod tidy` to install all needed dependencies
4. `go run .` to run the application
5. **WIP:** Right now there is no way of opening an MP3 file from the application, so you have to put it in `./assets/test_song.mp3` (exactly this name of the file) before starting the application


## Usage

### Project

WIP

### Markers

- start the player from the beginning by pressing Space key
- view the list of existing time markers
- filter time markers by name
- filter time markers by tags
- delete a specific time marker
- delete all time markers
- play from a specific time marker
- select a specific time marker with hotkeys (by pressing its list order number)
- edit a time marker (change name, time, add or remove category tags)
- use Tab and Enter key to interact with input fields and buttons without mouse
- create a new time marker

### Editor

- view the waveform of a loaded MP3 file
- zoom, pan, and navigate through the waveform 
- set a playhead position by clicking on the waveform
- nudge the playhead position with arrow keys
- start the player from the set playhead position by pressing Space key
- create a time marker from the set playhead position
- edit, drag or delete the name of an existing time marker
- set the playhead to a time marker's position


## Contributing

If you are good at math or Go and you spot that there is a room for improvement, you can create a PR.
