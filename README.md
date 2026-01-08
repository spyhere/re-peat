# ↺ re-peat

:construction: Work in progress :construction:

## Description

This is an application that can display sound waves of a given mp3 track, you can put/see/edit markers and able to play from them.

## Motivation

In theaters, philharmonics, and even amateur rehearsals, there is often a struggle to start playback from a specific point in the music. Since music usually serves as the backbone of a performance, and actors or dancers tend to repeat certain sections during rehearsals, they rely on music as a reference. The person responsible for the music then has to locate the correct point in the soundtrack, which is often a hassle.
The goal of this application is to eliminate that hassle by allowing users to mark those points once and then start the music directly from them.


## Quick Start

1. Clone the repo
2. Make sure you have `Go` toolchain installed, if you don't have it then check out [this](https://go.dev/doc/install)
3. Inside the cloned repo run `go mod tidy` to install all needed dependencies
4. `go run .` to run the application
5. **WIP:** Right now there is no way of opening an mp3 file from the application, so you have to put it in `./assets/test_song.mp3` (exactly this name of the file) before starting the application


## Usage

Right now you can scroll and pan the waves, as well as play with `spacebar` the audio track. Click on the sound waves to put a playhead there.

