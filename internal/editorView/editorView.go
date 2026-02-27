package editorview

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/widget"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/ui/theme"
	"github.com/tosone/minimp3"
)

const (
	waveEdgePadding = 3 // Forced to add this padding otherwise waves left and right border's px is being clipped
	playheadInitDur = time.Millisecond * 50
	playheadMinDur  = time.Millisecond * 20
)

type EditorProps struct {
	Dec           *minimp3.Decoder
	Player        *player.Player
	Th            *theme.RepeatTheme
	OnStartEditCb func()
	OnStopEditCb  func()
	Pcm           []byte
}

func NewEditor(props EditorProps) (*Editor, error) {
	normSamples, err := getNormalisedSamples(props.Pcm)
	if err != nil {
		return &Editor{}, err
	}
	fmt.Println("Audio data is normalised")
	frames := len(normSamples) / props.Dec.Channels
	monoSamples := makeSamplesMono(normSamples, props.Dec.Channels)
	fmt.Println("WaveRenderer received mono samples")
	return &Editor{
		p:             props.Player,
		monoSamples:   monoSamples,
		audio:         newAudio(props.Dec, props.Pcm, monoSamples, frames),
		playhead:      newPlayhead(playheadInitDur),
		cache:         newCache(),
		markers:       newMarkers(),
		mEditor:       newMEditor(),
		scroll:        newScroll(),
		th:            props.Th,
		tags:          newTags(),
		onStartEditCb: props.OnStartEditCb,
		onStopEditCb:  props.OnStopEditCb,
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
	mode          interactionMode
	cursor        pointer.Cursor
	playhead      *playhead
	audio         audio
	monoSamples   []float32
	cache         cache
	markers       *markers
	mEditor       *widget.Editor
	p             *player.Player
	waveM         int // wave margin
	tags          *tags
	size          image.Point
	scroll        scroll
	th            *theme.RepeatTheme
	onStartEditCb func()
	onStopEditCb  func()
}

func (ed *Editor) getRenderableWaves() [][2]float32 {
	samplesPerPx := ed.scroll.samplesPerPx
	visibleSamples := int(samplesPerPx * float32(ed.size.X))
	leftB := common.Clamp(0, ed.scroll.leftB, ed.audio.pcmMonoLen-visibleSamples)
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
		ed.waveM = common.PrcToPx(size.Y, ed.th.Sizing.Editor.WaveM)
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
	ed.scroll.samplesPerPx = common.Clamp(float32(ed.scroll.minSamplesPerPx), ed.scroll.samplesPerPx, float32(ed.scroll.maxSamplesPerPx))
	ed.scroll.leftB += int(pos.X * (oldSPP - ed.scroll.samplesPerPx))
	zoomFactor := ed.scroll.maxSamplesPerPx / ed.scroll.samplesPerPx
	playheadUpdate := time.Duration(float32(playheadInitDur) / zoomFactor)
	ed.playhead.update = common.Clamp(playheadMinDur, playheadUpdate, playheadInitDur)
	// Pan
	curSamplesPerPx := ed.scroll.samplesPerPx
	panSamples := int(scroll.X * panRate * float32(curSamplesPerPx))
	ed.scroll.leftB += panSamples
}

func (ed *Editor) startEdit(m *marker) {
	ed.mode = modeMEdit
	if m == nil {
		ed.markers.newMarker(ed.playhead.bytes)
	} else {
		ed.mEditor.SetText(m.name)
		ed.mEditor.SetCaret(len(m.name), 0)
		ed.markers.startEdit(m)
	}
	ed.onStartEditCb()
}

func (ed *Editor) cancelEdit() {
	if !ed.markers.isEditing() {
		return
	}
	if ed.markers.editing.name == "" {
		ed.markers.editing.markDead()
	}
	ed.markers.stopEdit()
	ed.mEditor.SetText("")
	ed.mode = modeIdle
	ed.onStopEditCb()
}

func (ed *Editor) confirmEdit(newName string) {
	ed.markers.editing.name = newName
	ed.markers.stopEdit()
	ed.mEditor.SetText("")
	ed.mode = modeIdle
	ed.onStopEditCb()
}

func (ed *Editor) handleMEditor(we widget.EditorEvent) {
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
	isEditing := ed.mode == modeMEdit
	return mInteraction{
		flag:    (ed.mode == modeMLife || ed.mode == modeMDeleteIntent || ed.mode == modeMCreateIntent) && !isEditing,
		pole:    (!isHovering || hoveringOverThis) && !isDragging && !isEditing,
		label:   !isDragging,
		hovered: hoveringOverThis,
	}
}

func (ed *Editor) setCursor(c pointer.Cursor) {
	ed.cursor = c
}

func (ed *Editor) updateDifferedState() {
	ed.markers.deleteDead()
}
