package state

import (
	"errors"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/x/explorer"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/filemanager"
	p "github.com/spyhere/re-peat/internal/player"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
)

const defaultPlayerVol = 0.5

func NewAppState(window *app.Window) AppState {
	return AppState{
		fileManager: filemanager.NewFileManager(window),
		TimeMarkers: tm.NewTimeMarkers(),
	}
}

type AppState struct {
	fileManager *filemanager.FileManager
	LoadedFile  string
	Player      *p.Player
	MonoSamples []float32
	AudioMeta   audio.AudioMeta
	FileMeta    filemanager.FileMeta
	TimeMarkers tm.TimeMarkers
	isChoosing  bool
	isLoading   bool
	err         error
}

func (a *AppState) IsChoosing() bool {
	return a.isChoosing
}
func (a *AppState) IsLoading() bool {
	return a.isLoading
}

func (a *AppState) HasAudioLoaded() bool {
	return a.LoadedFile != ""
}

func (a *AppState) GetError() error {
	err := a.err
	a.err = nil
	return err
}

func (a *AppState) Reset() {
	a.LoadedFile = ""
	a.err = nil
	if a.Player != nil {
		a.Player.Reset()
	}
	a.MonoSamples = a.MonoSamples[:0]
	a.TimeMarkers.MarkAllDead()
	a.TimeMarkers.DeleteDead()
}

func (a *AppState) AudioLoad() {
	a.isChoosing = true
	a.fileManager.Load(func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				a.err = err
			}
			return
		}
		if a.LoadedFile == filePath {
			return
		}

		a.isLoading = true
		a.Reset()
		defer func() {
			a.isLoading = false
		}()

		monoSamples, audioMeta, err := audio.LoadMonoSamples(filePath)
		if err != nil {
			a.err = err
			return
		}
		file, err := os.Open(filePath)
		if err != nil {
			a.err = err
			return
		}
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			a.err = err
			return
		}
		if a.Player == nil {
			a.Player = p.NewPlayer()
			a.Player.SetAudio(file)
			a.Player.SetVolume(defaultPlayerVol)
		} else {
			a.Player.SetAudio(file)
		}
		// NOTE: Is it safe to decode audio once?
		// Set everything at once only if it's happy path
		a.MonoSamples = monoSamples
		a.AudioMeta = audioMeta
		a.FileMeta = filemanager.NewFileMeta(filepath.Base(filePath), fileInfo.Size(), fileInfo.ModTime())
		a.LoadedFile = filePath
	}, ".mp3", ".wav", ".flac")
}
func (a *AppState) MarkersLoad()   {}
func (a *AppState) MarkersSave()   {}
func (a *AppState) MarkersSaveAs() {}
