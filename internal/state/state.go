package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/x/explorer"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/filemanager"
	p "github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/prompt"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const (
	defaultPlayerVol   = 0.5
	mLoadConflictTitle = "Markers loading conflict"
	mLoadConflictBody  = "These markers were initially saved for \"%s\", but currently loaded \"%s\".\nStill want to load them for this audio file?\n\nMarkers exceeding audio length will be set to 0 and have \"Redacted\" tag added."
)

func NewAppState(window *app.Window) AppState {
	th, err := theme.New()
	if err != nil {
		log.Fatal(err)
	}
	return AppState{
		Th:          th,
		Dialog:      common.Dialog{},
		Prompter:    prompt.NewPrompter(th),
		fileManager: filemanager.NewFileManager(window),
		TimeMarkers: tm.NewTimeMarkers(),
	}
}

type AppState struct {
	Th          *theme.RepeatTheme
	Dialog      common.Dialog
	Prompter    prompt.Prompter
	fileManager *filemanager.FileManager
	LoadedAFile string
	LoadedMFile string
	Player      *p.Player
	MonoSamples []float32
	AudioMeta   audio.AudioMeta
	MarkersMeta tm.MarkersMeta
	AFileMeta   filemanager.FileMeta
	MFileMeta   filemanager.FileMeta
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
	return a.LoadedAFile != ""
}
func (a *AppState) HasMarkersLoaded() bool {
	return a.LoadedMFile != ""
}

func (a *AppState) GetError() error {
	err := a.err
	a.err = nil
	return err
}

func (a *AppState) reset() {
	a.LoadedAFile = ""
	a.AFileMeta = filemanager.FileMeta{}
	a.AudioMeta = audio.AudioMeta{}
	a.err = nil
	if a.Player != nil {
		a.Player.Reset()
	}

	a.MonoSamples = a.MonoSamples[:0]
	a.TimeMarkers.MarkAllDead()
	a.TimeMarkers.DeleteDead()

	a.LoadedMFile = ""
	a.MFileMeta = filemanager.FileMeta{}
	a.MarkersMeta = tm.MarkersMeta{}
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
		if a.LoadedAFile == filePath {
			return
		}

		a.isLoading = true
		a.reset()
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
		a.AFileMeta = filemanager.NewFileMeta(filepath.Base(filePath), fileInfo.Size(), fileInfo.ModTime())
		a.LoadedAFile = filePath
	}, ".mp3", ".wav", ".flac")
}

func (a *AppState) MarkersLoad() {
	a.isChoosing = true
	a.fileManager.Load(func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				a.err = err
			}
			return
		}
		if a.LoadedMFile == filePath && !a.TimeMarkers.IsEmpty() {
			return
		}

		a.LoadedMFile = ""
		file, err := os.Open(filePath)
		if err != nil {
			a.err = err
			return
		}
		var saveStruct filemanager.MarkersSaveScheme
		decoder := json.NewDecoder(file)
		if err = decoder.Decode(&saveStruct); err != nil {
			a.err = err
			return
		}

		if a.AFileMeta.Name != saveStruct.FName {
			answer := a.Prompter.Ask(mLoadConflictTitle, fmt.Sprintf(mLoadConflictBody, saveStruct.FName, a.AFileMeta.Name))
			if answer == false {
				return
			}
			saveStruct.Markers.SanitizeSamples(a.AudioMeta.MaxMonoSamples())
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			a.err = err
			return
		}
		a.TimeMarkers = saveStruct.Markers
		a.MarkersMeta = tm.NewMarkersMeta(a.TimeMarkers)
		a.MFileMeta = filemanager.NewFileMeta(fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime())
		a.LoadedMFile = filePath
	}, ".rpt")

}

func (a *AppState) encodeMarkers() ([]byte, error) {
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	saveStruct := filemanager.MarkersSaveScheme{
		Version: 1,
		FName:   a.AFileMeta.Name,
		FSize:   a.AFileMeta.Size,
		FLen:    a.AudioMeta.Seconds,
		FSRate:  a.AudioMeta.SampleRate,
		Markers: a.TimeMarkers,
	}
	if err := encoder.Encode(saveStruct); err != nil {
		a.err = err
		return []byte{}, err
	}
	return data.Bytes(), nil
}

func (a *AppState) updateMarkersMeta(filePath string) {
	if filePath == "" {
		return
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		a.err = err
		return
	}
	a.MFileMeta = filemanager.NewFileMeta(fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime())
	a.MarkersMeta = tm.NewMarkersMeta(a.TimeMarkers)
}

func (a *AppState) MarkersSave() {
	if a.TimeMarkers.IsEmpty() || a.LoadedMFile == "" {
		return
	}
	data, err := a.encodeMarkers()
	if err != nil {
		a.err = err
		return
	}
	a.fileManager.Save(a.LoadedMFile, data, func(err error) {
		if err != nil {
			a.err = err
			return
		}
		a.updateMarkersMeta(a.LoadedMFile)
	})
}

func (a *AppState) MarkersSaveAs() {
	if a.TimeMarkers.IsEmpty() {
		return
	}
	data, err := a.encodeMarkers()
	if err != nil {
		a.err = err
		return
	}
	a.isChoosing = true
	a.fileManager.SaveAs("markers.rpt", data, func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				a.err = err
			}
			return
		}
		a.LoadedMFile = filePath
		a.updateMarkersMeta(filePath)
	})
}
