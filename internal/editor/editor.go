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
	"github.com/spyhere/re-peat/internal/constants"
	"github.com/spyhere/re-peat/internal/player"
	"github.com/spyhere/re-peat/internal/ui/theme"
	"github.com/tosone/minimp3"
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
			sampleRate:  dec.SampleRate,
			channels:    dec.Channels,
			pcmLen:      int64(len(pcm)),
			pcmMonoLen:  len(monoSamples),
			seconds:     float32(frames) / float32(dec.SampleRate),
			secsPerByte: 1.0 / (float32(dec.SampleRate) * constants.BYTES_PER_SAMPLE * float32(dec.Channels)),
		},
		margin:         400,
		padding:        90,
		playheadUpdate: time.Millisecond * 50,
		scroll: scroll{
			maxPxPerSec: 200,
		},
		th: th,
	}, nil
}

type Editor struct {
	playhead       int64 // Shows amount of PCM bytes from the beginning (not samples)
	playheadUpdate time.Duration
	audio          audio
	monoSamples    []float32
	cached         [][2]float32
	p              *player.Player
	margin         int
	padding        int
	size           image.Point
	scroll         scroll
	th             *theme.RepeatTheme
}

func makeSamplesMono(samples []float32, chanNum int) []float32 {
	if chanNum == 1 {
		return samples
	}
	if chanNum > 2 {
		fmt.Println("Not supported more than 2 channels")
		return []float32{}
	}
	res := make([]float32, len(samples)/chanNum)

	for i := 0; i < len(samples); i += 2 {
		lSample := samples[i]
		rSample := samples[i+1]
		res[i/2] = (lSample + rSample) * 0.5
	}
	return res
}

func (ed *Editor) getSamplesPerPx() int {
	pxPerSec := max(ed.scroll.minPxPerSec, ed.scroll.pxPerSec)
	return int(float32(ed.audio.sampleRate) / pxPerSec)
}

// TODO: optimisation - create multi-resolution downsampled samples map
func (ed *Editor) getRenderableWaves() [][2]float32 {
	prevSamplesPerPx := ed.audio.samplesPerPx
	samplesPerPx := ed.getSamplesPerPx()
	ed.audio.samplesPerPx = samplesPerPx
	maxSamples := samplesPerPx * ed.size.X
	sampleAtCursor := ed.scroll.leftB + int(ed.scroll.originX*float32(prevSamplesPerPx))
	leftB := sampleAtCursor - int(ed.scroll.originX*float32(samplesPerPx))
	leftB = clamp(0, leftB, ed.audio.pcmMonoLen-maxSamples)
	rightB := leftB + maxSamples
	if leftB == ed.scroll.leftB && rightB == ed.scroll.rightB {
		return ed.cached
	}
	ed.scroll.leftB = leftB
	ed.scroll.rightB = rightB

	monoSamples := ed.monoSamples[leftB:rightB]
	res := ed.cached[:ed.size.X]

	var idx int
	var min float32 = 1
	var max float32 = -1
	count := samplesPerPx
	for _, it := range monoSamples {
		if it < min {
			min = it
		}
		if it > max {
			max = it
		}
		count--
		if count == 0 {
			res[idx] = [2]float32{min, max}
			idx++
			min = 1
			max = -1
			count = samplesPerPx
		}
	}
	if count != samplesPerPx && idx < len(res) {
		res[idx] = [2]float32{min, max}
	}
	ed.cached = res
	return res
}

func (ed *Editor) SetSize(size image.Point) {
	ed.size = size
	if cap(ed.cached) < size.X {
		ed.cached = make([][2]float32, size.X, size.X*2)
	}
	ed.scroll.minPxPerSec = float32(size.X) / ed.audio.seconds
	ed.scroll.maxZoomExp = float32(math.Log2(float64(ed.scroll.maxPxPerSec) / float64(ed.scroll.minPxPerSec)))
}

func (ed *Editor) handleClick(posX float32) {
	pxPerSec := max(ed.scroll.minPxPerSec, ed.scroll.pxPerSec)
	seconds := (posX / pxPerSec) + (float32(ed.scroll.leftB) / float32(ed.audio.sampleRate))
	// TODO: handle error here
	seekVal, _ := ed.p.Search(seconds)
	ed.playhead = seekVal
}

const ZOOM_RATE = 0.0008
const PAN_RATE = 0.2

func (ed *Editor) handleScroll(scroll f32.Point, pos f32.Point) {
	ed.scroll.originX = pos.X

	panSamples := int(scroll.X * PAN_RATE * float32(ed.getSamplesPerPx()))
	ed.scroll.leftB += panSamples
	maxLeft := ed.audio.pcmMonoLen - ed.audio.samplesPerPx*ed.size.X
	ed.scroll.leftB = clamp(0, ed.scroll.leftB, maxLeft)

	ed.scroll.zoomExpDelta += scroll.Y * ZOOM_RATE
	ed.scroll.zoomExpDelta = clamp(0.0, ed.scroll.zoomExpDelta, ed.scroll.maxZoomExp)
	ed.scroll.pxPerSec = ed.scroll.minPxPerSec * float32(math.Exp2(float64(ed.scroll.zoomExpDelta)))
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

func (ed *Editor) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	player := ed.p
	isPlaying := player.IsPlaying()
	ed.handlePointerEvents(gtx)
	ed.handleKey(gtx, isPlaying)

	backgroundComp(gtx, ed.th.Editor.Bg)

	wavesYBorder := ed.size.Y/2 - ed.margin
	offsetBy(gtx, image.Pt(0, ed.margin), func() {
		soundWavesComp(gtx, ed.th, float32(wavesYBorder), ed.getRenderableWaves())
	})
	secondsRulerComp(gtx, ed.th, ed.margin, ed.audio, ed.scroll)

	playheadComp(gtx, ed.th, ed.playhead, ed.audio, ed.scroll)
	if isPlaying {
		if ed.playhead < ed.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(ed.playheadUpdate)})
		}
		ed.listenToPlayerUpdates()
	}
	return layout.Dimensions{}
}
