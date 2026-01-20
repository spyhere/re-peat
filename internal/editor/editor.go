package editor

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/ui/theme"
	"github.com/tosone/minimp3"
)

// TODO: Rename all constants to Go idiomatic way (no caps snake case)

const (
	WaveEdgePadding = 3 // Forced to add this padding otherwise waves left and right border's px is being clipped
	// TODO: Can we store it as ms duration?
	PLAYHEAD_INIT_DUR = 50
	MaxScrollLvl      = 5
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
		audio: audio{
			sampleRate: dec.SampleRate,
			channels:   dec.Channels,
			pcmLen:     int64(len(pcm)),
			pcmMonoLen: len(monoSamples),
			seconds:    float32(frames) / float32(dec.SampleRate),
		},
		margin:         400,
		padding:        90,
		playheadUpdate: time.Millisecond * PLAYHEAD_INIT_DUR,
		cache: cache{
			peakMap: make(map[int][][2]float32),
			levels:  make([]int, MaxScrollLvl+1),
			workers: make([]*cacheWorker, MaxScrollLvl+1),
		},
		scroll: scroll{
			maxLvl: MaxScrollLvl,
		},
		th: th,
	}, nil
}

type Editor struct {
	playhead       int64 // Shows amount of PCM bytes from the beginning (not samples)
	playheadUpdate time.Duration
	audio          audio
	monoSamples    []float32
	cache          cache
	p              *player.Player
	margin         int
	padding        int
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
	size.X += WaveEdgePadding
	ed.size = size
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

func (ed *Editor) handleClick(posX float32) {
	pxPerSec := float32(ed.audio.sampleRate) / float32(ed.scroll.samplesPerPx)
	seconds := (posX / pxPerSec) + (float32(ed.scroll.leftB) / float32(ed.audio.sampleRate))
	// TODO: handle error here
	seekVal, _ := ed.p.Search(seconds)
	ed.playhead = seekVal
}

const (
	ZOOM_RATE = 0.0008
	PAN_RATE  = 0.2
)

func (ed *Editor) handleScroll(scroll f32.Point, pos f32.Point) {
	// Pan
	curSamplesPerPx := ed.scroll.samplesPerPx
	panSamples := int(scroll.X * PAN_RATE * float32(curSamplesPerPx))
	ed.scroll.leftB += panSamples
	maxLeft := ed.audio.pcmMonoLen - int(curSamplesPerPx)*ed.size.X
	// TODO: Do I have to clamp it here?
	ed.scroll.leftB = clamp(0, ed.scroll.leftB, maxLeft)

	// Zoom
	// TODO: maybe zoom should be before pan?
	oldSPP := ed.scroll.samplesPerPx
	ed.scroll.samplesPerPx *= float32(math.Exp(float64(-scroll.Y * ZOOM_RATE)))
	ed.scroll.samplesPerPx = clamp(float32(ed.scroll.minSamplesPerPx), ed.scroll.samplesPerPx, float32(ed.scroll.maxSamplesPerPx))
	ed.scroll.leftB += int(pos.X * (oldSPP - ed.scroll.samplesPerPx))
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

func (ed *Editor) handlePointerEvents(gtx layout.Context) {
	event.Op(gtx.Ops, ed)
	for {
		evt, ok := gtx.Event(pointer.Filter{
			Target:  ed,
			Kinds:   pointer.Press | pointer.Scroll,
			ScrollX: pointer.ScrollRange{Min: -1e9, Max: 1e9},
			ScrollY: pointer.ScrollRange{Min: -1e9, Max: 1e9},
		})
		if !ok {
			break
		}
		e, ok := evt.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Kind {
		case pointer.Scroll:
			ed.handleScroll(e.Scroll, e.Position)
		case pointer.Press:
			ed.handleClick(e.Position.X)
		}
	}
}

// TODO: Move this method to separate file
func (ed *Editor) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	player := ed.p
	isPlaying := player.IsPlaying()
	ed.handlePointerEvents(gtx)
	ed.handleKey(gtx, isPlaying)

	backgroundComp(gtx, ed.th.Editor.Bg)

	yCenter := gtx.Constraints.Max.Y / 2
	offsetBy(gtx, image.Pt(0, ed.margin), func() {
		soundWavesComp(gtx, ed.th, float32(yCenter-ed.margin), ed.getRenderableWaves(), ed.scroll, ed.cache)
	})
	secondsRulerComp(gtx, ed.th, ed.margin-50, ed.audio, ed.scroll)

	playheadComp(gtx, ed.th, ed.playhead, ed.audio, ed.scroll)
	if isPlaying {
		if ed.playhead < ed.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playheadUpdate)})
		}
		ed.listenToPlayerUpdates()
	}
	return layout.Dimensions{}
}
