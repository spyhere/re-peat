package editor

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/widget"
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
		p:           player,
		monoSamples: monoSamples,
		audio:       newAudio(dec, pcm, monoSamples, frames),
		playhead:    newPlayhead(playheadInitDur),
		cache:       newCache(),
		markers:     newMarkers(),
		renamer:     newRenamer(),
		scroll:      newScroll(),
		th:          th,
		tags:        newTags(),
	}, nil
}

type interactionMode int

const (
	modeIdle interactionMode = iota
	modeHitWave
	modeMLife
	modeMCreateIntent
	modeMDeleteIntent
	modeMHit
	modeMEditIntent
	modeMEdit
	modeMDrag
)

type Editor struct {
	mode        interactionMode
	cursor      pointer.Cursor
	playhead    *playhead
	audio       audio
	monoSamples []float32
	cache       cache
	markers     *markers
	renamer     *widget.Editor
	p           *player.Player
	waveM       int // wave margin
	tags        *tags
	size        image.Point
	scroll      scroll
	th          *theme.RepeatTheme
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

func (ed *Editor) playheadPosFromX(posX float32) {
	pxPerSec := float32(ed.audio.sampleRate) / float32(ed.scroll.samplesPerPx)
	seconds := (posX / pxPerSec) + (float32(ed.scroll.leftB) / float32(ed.audio.sampleRate))
	// TODO: handle error here
	seekVal, _ := ed.p.Search(seconds)
	ed.playhead.set(seekVal)
}

func (ed *Editor) setPlayhead(pcmValue int64) {
	// TODO: handle error here
	seekValue, _ := ed.p.Set(pcmValue)
	ed.playhead.set(seekValue)
}

func (ed *Editor) handleWaveClick(pCoords f32.Point, buttons pointer.Buttons) {
	switch buttons {
	case pointer.ButtonPrimary:
		switch ed.mode {
		case modeHitWave:
			ed.playheadPosFromX(pCoords.X)
		}
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
	ed.playhead.update = clamp(playheadMinDur, playheadUpdate, playheadInitDur)
	// Pan
	curSamplesPerPx := ed.scroll.samplesPerPx
	panSamples := int(scroll.X * panRate * float32(curSamplesPerPx))
	ed.scroll.leftB += panSamples
}

func (ed *Editor) confirmEdit(newName string) {
	ed.markers.editing.name = newName
	ed.markers.stopEdit()
	ed.renamer.SetText("")
	ed.mode = modeIdle
}

func (ed *Editor) handleRenamer(we widget.EditorEvent) {
	if e, ok := we.(widget.SubmitEvent); ok {
		if e.Text == "" {
			return
		}
		ed.confirmEdit(e.Text)
	}
}

func (ed *Editor) listenToPlayerUpdates() {
	player := ed.p
	select {
	case _ = <-player.IsDoneCh():
		ed.playhead.bytes = ed.audio.pcmLen
		// We need to pause it after it's done to mitigate the potential bug. See [player.IsDoneCh] comment.
		ed.p.Pause()
	default:
		ed.playhead.bytes = player.GetReadAmount()
	}
}

func (ed *Editor) isCreateButtonVisible() bool {
	return ed.mode == modeMLife || ed.mode == modeMCreateIntent || ed.mode == modeMDeleteIntent
}

func (ed *Editor) getMI9n(m *marker) mInteraction {
	isHovering := ed.markers.isHovering()
	hoveringOverThis := isHovering && ed.markers.hovering == m
	isDragging := ed.mode == modeMDrag
	return mInteraction{
		flag:  ed.mode == modeMLife || ed.mode == modeMDeleteIntent || ed.mode == modeMCreateIntent,
		pole:  (!isHovering || hoveringOverThis) && !isDragging,
		label: !isDragging,
	}
}

func (ed *Editor) setCursor(c pointer.Cursor) {
	ed.cursor = c
}

func (ed *Editor) updateDifferedState() {
	ed.markers.deleteDead()
}
