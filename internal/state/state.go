package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/x/explorer"
	"github.com/spyhere/re-peat/internal/audio"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/configs"
	"github.com/spyhere/re-peat/internal/filemanager"
	"github.com/spyhere/re-peat/internal/filters"
	"github.com/spyhere/re-peat/internal/i18n"
	"github.com/spyhere/re-peat/internal/logging"
	p "github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/playhead"
	"github.com/spyhere/re-peat/internal/prompt"
	tm "github.com/spyhere/re-peat/internal/timeMarkers"
	"github.com/spyhere/re-peat/internal/ui/theme"
)

const (
	defaultPlayerVol = 0.5
)

func NewAppState(window *app.Window, lg logging.Logger, cfgs *configs.Configs) (AppState, error) {
	th, err := theme.New()
	if err != nil {
		return AppState{}, err
	}
	locale, err := cfgs.GetLocale()
	if err != nil {
		lg.Warn("Failed to get locale", "err", err)
	}
	return AppState{
		Cfgs:        cfgs,
		Lg:          lg,
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
	Cfgs        *configs.Configs
	Lg          logging.Logger
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
	MonoSamples []float32 // NOTE: Should it stay in state or moved to Editor?
	AudioMeta   audio.AudioMeta
	MarkersMeta tm.MarkersMeta
	AFileMeta   filemanager.FileMeta
	MFileMeta   filemanager.FileMeta
	TimeMarkers tm.TimeMarkers
	isChoosing  bool
	isLoading   bool
	isDecoding  bool
}

func (a *AppState) IsChoosing() bool {
	return a.isChoosing
}
func (a *AppState) IsLoading() bool {
	return a.isLoading
}
func (a *AppState) IsDecoding() bool {
	return a.isDecoding
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

func (a *AppState) DecodeAllSamples() {
	if a.isDecoding {
		return
	}
	a.isDecoding = true
	go func() {
		monoSamples, err := audio.FileToMonosamples(a.LoadedAFile)
		if err != nil {
			a.Lg.Error("Decoding samples", err)
		}
		a.Lg.Info("Decoded samples", "file", a.LoadedAFile)
		a.MonoSamples = monoSamples
		a.isDecoding = false
	}()
}

func (a *AppState) AudioLoad() {
	a.pausePlayer()
	a.isChoosing = true
	a.fileManager.Load(func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				a.Lg.Error("AudioLoad", err)
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

		file, err := os.Open(filePath)
		if err != nil {
			a.Lg.Error("AudioLoad", err)
			return
		}
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			a.Lg.Error("AudioLoad", err)
			return
		}
		var audioMeta audio.AudioMeta
		if a.Player == nil {
			a.Player = p.NewPlayer()
			audioMeta, err = a.Player.SetAudio(file)
			a.Player.SetVolume(defaultPlayerVol)
		} else {
			audioMeta, err = a.Player.SetAudio(file)
		}

		if err != nil {
			a.Lg.Error("AudioLoad", err)
			return
		}
		// Set everything at once only if it's happy path
		a.MonoSamples = a.MonoSamples[:0]
		a.AudioMeta = audioMeta
		a.AFileMeta = filemanager.NewFileMeta(filepath.Base(filePath), fileInfo.Size(), fileInfo.ModTime())
		a.LoadedAFile = filePath
		a.resetAudioDependantState()
		a.Lg.Info("Audio loaded")
	}, ".mp3", ".wav", ".flac")
}

func (a *AppState) MarkersLoad() {
	a.pausePlayer()
	a.isChoosing = true
	a.fileManager.Load(func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				a.Lg.Error("MarkersLoad", err)
			}
			return
		}
		if a.LoadedMFile == filePath && !a.TimeMarkers.IsEmpty() {
			return
		}

		a.LoadedMFile = ""
		file, err := os.Open(filePath)
		if err != nil {
			a.Lg.Error("MarkersLoad", err)
			return
		}
		var saveStruct filemanager.MarkersSaveScheme
		decoder := json.NewDecoder(file)
		if err = decoder.Decode(&saveStruct); err != nil {
			a.Lg.Error("MarkersLoad", err)
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
			a.Lg.Error("MarkersLoad", err)
			return
		}
		a.TimeMarkers = saveStruct.Markers
		a.MarkersMeta = tm.NewMarkersMeta(a.TimeMarkers)
		a.ChipsFilter.Recreate(a.TimeMarkers)
		a.MFileMeta = filemanager.NewFileMeta(fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime())
		a.LoadedMFile = filePath
		a.Lg.Info("Markers loaded")
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
		return []byte{}, err
	}
	return data.Bytes(), nil
}

func (a *AppState) updateMarkersMeta(filePath string) {
	if filePath == "" {
		a.Lg.Warn("updateMarkersMeta: file path is not specified")
		return
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		a.Lg.Error("updateMarkersMeta", err)
		return
	}
	a.MFileMeta = filemanager.NewFileMeta(fileInfo.Name(), fileInfo.Size(), fileInfo.ModTime())
	a.MarkersMeta = tm.NewMarkersMeta(a.TimeMarkers)
}

func (a *AppState) MarkersSave() {
	if a.TimeMarkers.IsEmpty() || a.LoadedMFile == "" {
		a.Lg.Warn("MarkersSave: unreachable", "markersLen", len(a.TimeMarkers), "loadedMFile", a.LoadedMFile)
		return
	}
	data, err := a.encodeMarkers()
	if err != nil {
		a.Lg.Error("MarkersSave", err)
		return
	}
	a.fileManager.Save(a.LoadedMFile, data, func(err error) {
		if err != nil {
			a.Lg.Error("MarkersSave", err)
			return
		}
		a.updateMarkersMeta(a.LoadedMFile)
		a.Lg.Info("Markers save")
	})
}

func (a *AppState) MarkersSaveAs() {
	if a.TimeMarkers.IsEmpty() {
		a.Lg.Warn("MarkersSaveAs: unreachable. Markers are empty")
		return
	}
	data, err := a.encodeMarkers()
	if err != nil {
		a.Lg.Error("MarkersSaveAs", err)
		return
	}
	a.isChoosing = true
	a.fileManager.SaveAs("markers.rpt", data, func(filePath string, err error) {
		a.isChoosing = false
		if err != nil {
			if !errors.Is(err, explorer.ErrUserDecline) {
				a.Lg.Error("MarkersSaveAs", err)
			}
			return
		}
		a.LoadedMFile = filePath
		a.updateMarkersMeta(filePath)
		a.Lg.Info("Markers saved as")
	})
}

func (a *AppState) NotifyCrashReportsOnStartup() {
	home, err := os.UserHomeDir()
	if err != nil {
		a.Lg.Error("Checking for crash", err)
		return
	}
	desktop := filepath.Join(home, "Desktop")
	matches, err := filepath.Glob(filepath.Join(desktop, logging.CrashReportFileName+"*.txt"))
	if err != nil {
		a.Lg.Error("Unreachable", err)
		return
	}
	if len(matches) == 0 {
		return
	}
	commonI18n := a.I18n.Common
	reports := strings.Join(matches, "\n")
	body := fmt.Sprintf(commonI18n.CrashFoundBody, len(matches), reports)
	a.Prompter.Tell(commonI18n.CrashFoundTitle, body, commonI18n.InfoDialogOk)
}
