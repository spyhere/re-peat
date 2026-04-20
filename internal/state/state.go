package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/x/explorer"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/filemanager"
	"github.com/spyhere/re-peat/internal/filters"
	"github.com/spyhere/re-peat/internal/i18n"
	p "github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/playhead"
	"github.com/spyhere/re-peat/internal/prompt"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const (
	defaultPlayerVol = 0.5
)

func NewAppState(window *app.Window, locale string) (AppState, error) {
	th, err := theme.New()
	if err != nil {
		return AppState{}, err
	}
	return AppState{
		I18n:        i18n.NewI18n(i18n.Parse(locale)),
		Th:          th,
		ChipsFilter: filters.NewChipsFilter(100), // TODO: think about centralized way of capacity constant
		Dialog:      common.Dialog{},
		Prompter:    prompt.NewPrompter(th),
		fileManager: filemanager.NewFileManager(window),
		TimeMarkers: tm.NewTimeMarkers(),
	}, nil
}

type AppState struct {
	I18n        i18n.State
	Th          *theme.RepeatTheme
	ChipsFilter filters.ChipsFilter
	SearchbarV  string
	Dialog      common.Dialog
	Prompter    prompt.Prompter
	Playhead    playhead.Transport
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

func (a *AppState) resetAudioDependantState() {
	a.Playhead.Set(0)
	a.TimeMarkers.MarkAllDead()
	a.TimeMarkers.DeleteDead()

	a.LoadedMFile = ""
	a.MFileMeta = filemanager.FileMeta{}
	a.MarkersMeta = tm.MarkersMeta{}
}

func (a *AppState) pausePlayer() {
	if a.Player != nil {
		a.Player.Pause()
	}
}

func (a *AppState) AudioLoad() error {
	a.pausePlayer()
	a.isChoosing = true
	var resultErr error
	a.fileManager.Load(func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				resultErr = err
			}
			return
		}
		if a.LoadedAFile == filePath {
			return
		}

		a.isLoading = true
		defer func() {
			a.isLoading = false
		}()

		monoSamples, audioMeta, err := audio.LoadMonoSamples(filePath)
		if err != nil {
			resultErr = err
			return
		}
		file, err := os.Open(filePath)
		if err != nil {
			resultErr = err
			return
		}
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			resultErr = err
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
		a.resetAudioDependantState()
	}, ".mp3", ".wav", ".flac")
	return resultErr
}

func (a *AppState) MarkersLoad() error {
	a.pausePlayer()
	a.isChoosing = true
	var resultErr error
	a.fileManager.Load(func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				resultErr = err
			}
			return
		}
		if a.LoadedMFile == filePath && !a.TimeMarkers.IsEmpty() {
			return
		}

		a.LoadedMFile = ""
		file, err := os.Open(filePath)
		if err != nil {
			resultErr = err
			return
		}
		var saveStruct filemanager.MarkersSaveScheme
		decoder := json.NewDecoder(file)
		if err = decoder.Decode(&saveStruct); err != nil {
			resultErr = err
			return
		}

		if a.AFileMeta.Name != saveStruct.FName {
			title := a.I18n.Project.MConflictLoadTitle
			body := a.I18n.Project.MConflictLoadBody
			answer := a.Prompter.Ask(title, fmt.Sprintf(body, saveStruct.FName, a.AFileMeta.Name))
			if answer == false {
				return
			}
			saveStruct.Markers.SanitizeSamples(a.AudioMeta.MaxMonoSamples())
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			resultErr = err
			return
		}
		a.TimeMarkers = saveStruct.Markers
		a.MarkersMeta = tm.NewMarkersMeta(a.TimeMarkers)
		a.ChipsFilter.Recreate(a.TimeMarkers)
		a.MFileMeta = filemanager.NewFileMeta(fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime())
		a.LoadedMFile = filePath
	}, ".rpt")
	return resultErr
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
		return []byte{}, err
	}
	return data.Bytes(), nil
}

func (a *AppState) updateMarkersMeta(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("File path is not specified")
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	a.MFileMeta = filemanager.NewFileMeta(fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime())
	a.MarkersMeta = tm.NewMarkersMeta(a.TimeMarkers)
	return nil
}

func (a *AppState) MarkersSave() error {
	if a.TimeMarkers.IsEmpty() || a.LoadedMFile == "" {
		return fmt.Errorf("Unreachable MarkersSave! markersLen: %v, loadedMFile: %v", len(a.TimeMarkers), a.LoadedMFile)
	}
	data, err := a.encodeMarkers()
	if err != nil {
		return err
	}
	var resultErr error
	a.fileManager.Save(a.LoadedMFile, data, func(err error) {
		if err != nil {
			resultErr = err
			return
		}
		a.updateMarkersMeta(a.LoadedMFile)
	})
	return resultErr
}

func (a *AppState) MarkersSaveAs() error {
	if a.TimeMarkers.IsEmpty() {
		return fmt.Errorf("Unreachable MarkersSaveAs! Markers are empty")
	}
	data, err := a.encodeMarkers()
	if err != nil {
		return err
	}
	a.isChoosing = true
	var resultErr error
	a.fileManager.SaveAs("markers.rpt", data, func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				resultErr = err
			}
			return
		}
		a.LoadedMFile = filePath
		a.updateMarkersMeta(filePath)
	})
	return resultErr
}
