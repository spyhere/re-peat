# ↺ re-peat

## Description

This is an application that can display sound waves of a given MP3 track, you can put/see/edit markers and able to play from them.

## Motivation

In theaters, philharmonics, and even amateur rehearsals, there is often a struggle to start playback from a specific point in the music. Since music usually serves as the backbone of a performance, and actors or dancers tend to repeat certain sections during rehearsals, they rely on music as a reference. The person responsible for the music then has to locate the correct point in the soundtrack, which is often a hassle.
The goal of this application is to eliminate that hassle by allowing users to mark those points once and then start the music directly from them.

## macOS users

This application is not signed with Apple Developer ID, so it won't run by default on macOS. If you still want to run it, open Terminal in the directory, where unzipped application is located, and run:
```
xattr -dr com.apple.quarantine re-peat.app
```
This will remove quarantine attribute from the binary, so macOS won't block its execution.

## Demo

https://github.com/user-attachments/assets/233cf3d8-69d5-40ba-94cf-6a9ac8e92d9e

https://github.com/user-attachments/assets/de938d8c-4bb0-42dc-b3c9-d479ff98eb7b

## Quick Start

1. Clone the repo
2. Make sure you have `Go` toolchain installed, if you don't have it then check out [this](https://go.dev/doc/install)
3. Inside the cloned repo run `go mod tidy` to install all needed dependencies
4. `go run .` to run the application

## Usage

### Project

- load mp3/wav/flac audio file
- view the audio file stats
- load and save markers
- view markers file stats

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
- add comment to the marker

### Editor

- view the waveform of a loaded MP3 file
- zoom, pan, and navigate through the waveform 
- set a playhead position by clicking on the waveform
- nudge the playhead position with arrow keys
- start the player from the set playhead position by pressing Space key
- create a time marker from the set playhead position
- edit, drag or delete an existing time marker
- set the playhead to a time marker's position

## Error logs

### Crash

When the application crashes, it creates a crash report on your Desktop. If the application detects at least 1 crash report on the Desktop, it will notify the user on the next startup.

### Error

There still a possibility of device or software faults. Such errors are not expected, but if they occur, an error log will be immediately saved to the Desktop and the user will be notified.


## Contributing

If you are good at math or Go and you spot that there is a room for improvement, you can create a PR.

## License

This product is licensed under Apache 2.0
