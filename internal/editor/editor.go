package editor

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/ui/theme"
	"github.com/tosone/minimp3"
)

const (
	waveEdgePadding = 3 // Forced to add this padding otherwise waves left and right border's px is being clipped
	playheadInitDur = time.Millisecond * 50
	playheadMinDur  = time.Millisecond * 20
)

func NewEditor(th *theme.RepeatTheme, dec *minimp3.Decoder, pcm []byte, player *player.Player) (*Editor, error) {
	normSamples, err := getNormalisedSamples(pcm)
	if err != nil {
		return &Editor{}, err
	}
	fmt.Println("Audio data is normalised")
	frames := len(normSamples) / dec.Channels
	monoSamples := makeSamplesMono(normSamples, dec.Channels)
	fmt.Println("WaveRenderer received mono samples")
	return &Editor{
		p:              player,
		monoSamples:    monoSamples,
		audio:          newAudio(dec, pcm, monoSamples, frames),
		playheadUpdate: playheadInitDur,
		cache:          newCache(),
		markers:        newMarkers(),
		scroll:         newScroll(),
		th:             th,
	}, nil
}

type interactionMode int

const (
	modeIdle interactionMode = iota
	modeHitWave
	modeSetMarker
	modeHitMarker
	modeDragMarker
)

type Editor struct {
	mode           interactionMode
	cursor         pointer.Cursor
	playhead       int64 // Shows amount of PCM bytes from the beginning (not samples)
	playheadUpdate time.Duration
	audio          audio
	monoSamples    []float32
	cache          cache
	markers        *markers
	p              *player.Player
	waveM          int // wave margin
	waveTag        struct{}
	size           image.Point
	scroll         scroll
	th             *theme.RepeatTheme
}

func (ed *Editor) getRenderableWaves() [][2]float32 {
	samplesPerPx := ed.scroll.samplesPerPx
	visibleSamples := int(samplesPerPx * float32(ed.size.X))
	leftB := clamp(0, ed.scroll.leftB, ed.audio.pcmMonoLen-visibleSamples)
	rightB := leftB + visibleSamples
	if leftB == ed.scroll.leftB && rightB == ed.scroll.rightB {
		return ed.cache.curSlice
	}
	ed.scroll.leftB = leftB
	ed.scroll.rightB = rightB

	cacheSPP := ed.cache.getLevel(samplesPerPx)
	cacheLeftB := leftB / cacheSPP
	cacheRightB := rightB / cacheSPP

	ed.cache.curSlice = ed.cache.peakMap[cacheSPP][cacheLeftB:cacheRightB]
	ed.cache.curLvl = cacheSPP
	ed.cache.leftB = cacheLeftB
	return ed.cache.curSlice
}

func (ed *Editor) SetSize(size image.Point) {
	prev := ed.size
	size.X += waveEdgePadding
	ed.size = size
	if !prev.Eq(ed.size) {
		ed.waveM = prcToPx(size.Y, ed.th.Sizing.Editor.WaveM)
		ed.cache.isPopulated = false
	}
}

// TODO: optimisation - debounce on window resize
func (ed *Editor) MakePeakMap() {
	if ed.cache.isPopulated {
		return
	}
	ed.scroll.maxSamplesPerPx = float32(ed.audio.sampleRate) / (float32(ed.size.X) / ed.audio.seconds)
	ed.scroll.minSamplesPerPx = ed.scroll.maxSamplesPerPx / float32(math.Exp2(float64(ed.scroll.maxLvl)))
	ed.scroll.samplesPerPx = float32(ed.scroll.maxSamplesPerPx)

	idx := 0
	maxSamplesPerPx := int(ed.scroll.maxSamplesPerPx)
	minSamplesPerPx := int(ed.scroll.minSamplesPerPx)
	for i := maxSamplesPerPx; i >= minSamplesPerPx; i /= 2 {
		ed.cache.workers[idx] = &cacheWorker{
			samplesPerPx: i,
			min:          1,
			max:          -1,
			count:        i,
		}
		ed.cache.peakMap[i] = make([][2]float32, len(ed.monoSamples)/i)
		ed.cache.levels[idx] = i
		idx++
	}
	populateCache(ed.cache.peakMap, ed.monoSamples, ed.cache.workers)
	ed.cache.isPopulated = true
}

func (ed *Editor) setPlayhead(posX float32) {
	pxPerSec := float32(ed.audio.sampleRate) / float32(ed.scroll.samplesPerPx)
	seconds := (posX / pxPerSec) + (float32(ed.scroll.leftB) / float32(ed.audio.sampleRate))
	// TODO: handle error here
	seekVal, _ := ed.p.Search(seconds)
	ed.playhead = seekVal
}

func (ed *Editor) handleWaveClick(pCoords f32.Point, buttons pointer.Buttons) {
	switch buttons {
	case pointer.ButtonPrimary:
		switch ed.mode {
		case modeHitWave:
			ed.setPlayhead(pCoords.X)
		case modeSetMarker:
			samples := ed.scroll.getSamplesFromPx(ed.markers.draft.pointerX)
			ed.markers.NewMarker(samples)
		}
	case pointer.ButtonSecondary:
		switch ed.mode {
		case modeHitWave:
			ed.mode = modeSetMarker
			ed.markers.enableDraft(pCoords.X)
		case modeSetMarker:
			ed.mode = modeHitWave
			ed.markers.disableDraft()
		}
	}
}

func (ed *Editor) handleWaveMove(m *marker, pCoord f32.Point) {
	switch ed.mode {
	case modeSetMarker:
		ed.markers.enableDraft(pCoord.X)
	case modeDragMarker:
		dSamples := ed.scroll.samplesPerPx * pCoord.X
		m.Samples += int(dSamples)
		m.Samples = clamp(0, m.Samples, ed.scroll.rightB)
	}
}

const (
	zoomRate = 0.0008
	panRate  = 0.2
)

func (ed *Editor) handleWaveScroll(scroll f32.Point, pos f32.Point) {
	// Zoom
	oldSPP := ed.scroll.samplesPerPx
	ed.scroll.samplesPerPx *= float32(math.Exp(float64(-scroll.Y * zoomRate)))
	ed.scroll.samplesPerPx = clamp(float32(ed.scroll.minSamplesPerPx), ed.scroll.samplesPerPx, float32(ed.scroll.maxSamplesPerPx))
	ed.scroll.leftB += int(pos.X * (oldSPP - ed.scroll.samplesPerPx))
	zoomFactor := ed.scroll.maxSamplesPerPx / ed.scroll.samplesPerPx
	playheadUpdate := time.Duration(float32(playheadInitDur) / zoomFactor)
	ed.playheadUpdate = clamp(playheadMinDur, playheadUpdate, playheadInitDur)
	// Pan
	curSamplesPerPx := ed.scroll.samplesPerPx
	panSamples := int(scroll.X * panRate * float32(curSamplesPerPx))
	ed.scroll.leftB += panSamples
}

func (ed *Editor) handleKey(gtx layout.Context, isPlaying bool) {
	for {
		evt, ok := gtx.Event(key.Filter{
			Name: key.NameSpace,
		})
		if !ok {
			break
		}
		e, ok := evt.(key.Event)
		if !ok {
			continue
		}
		if e.State == key.Press {
			if e.Name == key.NameSpace {
				isPlaying = !isPlaying
				if isPlaying {
					if ed.playhead >= ed.audio.pcmLen {
						continue
					}
					ed.p.Play()
					ed.p.WaitUntilReady()
				} else {
					ed.p.Pause()
				}
			}
		}
	}
}

func (ed *Editor) listenToPlayerUpdates() {
	player := ed.p
	select {
	case _ = <-player.IsDoneCh():
		ed.playhead = ed.audio.pcmLen
		// We need to pause it after it's done to mitigate the potential bug. See [player.IsDoneCh] comment.
		ed.p.Pause()
	default:
		ed.playhead = player.GetReadAmount()
	}
}

func (ed *Editor) shouldMarkersInterest() bool {
	return ed.mode != modeSetMarker && ed.mode != modeDragMarker
}

func (ed *Editor) setCursor(c pointer.Cursor) {
	ed.cursor = c
}
