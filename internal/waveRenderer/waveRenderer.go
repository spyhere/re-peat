package waverenderer

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/spyhere/re-peat/internal/constants"
	"github.com/spyhere/re-peat/internal/player"
	"github.com/tosone/minimp3"
)

func NewWavesRenderer(dec *minimp3.Decoder, pcm []byte, player *player.Player) (*WavesRenderer, error) {
	normSamples, err := getNormalisedSamples(pcm)
	if err != nil {
		return &WavesRenderer{}, err
	}
	fmt.Println("Audio data is normalised")
	frames := len(normSamples) / dec.Channels
	monoSamples := makeSamplesMono(normSamples, dec.Channels)
	fmt.Println("WaveRenderer received mono samples")
	return &WavesRenderer{
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
	}, nil
}

// Entity to visualise wave forms of sound track
type WavesRenderer struct {
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

func (r *WavesRenderer) getSamplesPerPx() int {
	pxPerSec := max(r.scroll.minPxPerSec, r.scroll.pxPerSec)
	return int(float32(r.audio.sampleRate) / pxPerSec)
}

func (r *WavesRenderer) getRenderableWaves() [][2]float32 {
	prevSamplesPerPx := r.audio.samplesPerPx
	samplesPerPx := r.getSamplesPerPx()
	r.audio.samplesPerPx = samplesPerPx
	maxSamples := samplesPerPx * r.size.X
	sampleAtCursor := r.scroll.leftB + int(r.scroll.originX*float32(prevSamplesPerPx))
	leftB := sampleAtCursor - int(r.scroll.originX*float32(samplesPerPx))
	leftB = clamp(0, leftB, r.audio.pcmMonoLen-maxSamples)
	rightB := leftB + maxSamples
	if leftB == r.scroll.leftB && rightB == r.scroll.rightB {
		return r.cached
	}
	r.scroll.leftB = leftB
	r.scroll.rightB = rightB

	monoSamples := r.monoSamples[leftB:rightB]
	res := r.cached[:r.size.X]

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
	r.cached = res
	return res
}

func (r *WavesRenderer) SetSize(size image.Point) {
	r.size = size
	if cap(r.cached) < size.X {
		r.cached = make([][2]float32, size.X, size.X*2)
	}
	r.scroll.minPxPerSec = float32(size.X) / r.audio.seconds
	r.scroll.maxZoomExp = float32(math.Log2(float64(r.scroll.maxPxPerSec) / float64(r.scroll.minPxPerSec)))
}

func (r *WavesRenderer) handleClick(posX float32) {
	pxPerSec := max(r.scroll.minPxPerSec, r.scroll.pxPerSec)
	seconds := (posX / pxPerSec) + (float32(r.scroll.leftB) / float32(r.audio.sampleRate))
	// TODO: handle error here
	seekVal, _ := r.p.Search(seconds)
	r.playhead = seekVal
}

const ZOOM_RATE = 0.0008
const PAN_RATE = 0.2

func (r *WavesRenderer) handleScroll(scroll f32.Point, pos f32.Point) {
	r.scroll.originX = pos.X

	panSamples := int(scroll.X * PAN_RATE * float32(r.getSamplesPerPx()))
	r.scroll.leftB += panSamples
	maxLeft := r.audio.pcmMonoLen - r.audio.samplesPerPx*r.size.X
	r.scroll.leftB = clamp(0, r.scroll.leftB, maxLeft)

	r.scroll.zoomExpDelta += scroll.Y * ZOOM_RATE
	r.scroll.zoomExpDelta = clamp(0.0, r.scroll.zoomExpDelta, r.scroll.maxZoomExp)
	r.scroll.pxPerSec = r.scroll.minPxPerSec * float32(math.Exp2(float64(r.scroll.zoomExpDelta)))
}

func (r *WavesRenderer) handleKey(gtx layout.Context, isPlaying bool) {
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
					if r.playhead >= r.audio.pcmLen {
						continue
					}
					r.p.Play()
					r.p.WaitUntilReady()
				} else {
					r.p.Pause()
				}
			}
		}
	}
}

func (r *WavesRenderer) listenToPlayerUpdates() {
	player := r.p
	select {
	case _ = <-player.IsDoneCh():
		r.playhead = r.audio.pcmLen
		// We need to pause it after it's done to mitigate the potential bug. See [player.IsDoneCh] comment.
		r.p.Pause()
	default:
		r.playhead = player.GetReadAmount()
	}
}

func (r *WavesRenderer) handlePointerEvents(gtx layout.Context) {
	event.Op(gtx.Ops, r)
	for {
		evt, ok := gtx.Event(pointer.Filter{
			Target:  r,
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
			r.handleScroll(e.Scroll, e.Position)
		case pointer.Press:
			r.handleClick(e.Position.X)
		}
	}
}

func (r *WavesRenderer) Layout(gtx layout.Context, th *material.Theme, e app.FrameEvent) layout.Dimensions {
	player := r.p
	isPlaying := player.IsPlaying()
	r.handlePointerEvents(gtx)
	r.handleKey(gtx, isPlaying)

	backgroundComp(gtx, color.NRGBA{A: 0xff})
	bgArea := image.Rect(0, r.margin-r.padding, r.size.X, r.size.Y-r.margin+r.padding)
	ColorBox(gtx, bgArea, color.NRGBA{R: 0x11, G: 0x77, B: 0x66, A: 0xff})

	wavesYBorder := r.size.Y/2 - r.margin
	offsetBy(gtx, image.Pt(0, r.margin), func() {
		soundWavesComp(gtx, float32(wavesYBorder), r.getRenderableWaves())
	})
	secondsRulerComp(gtx, th, r.margin, r.audio, r.scroll)

	playheadComp(gtx, r.playhead, r.audio, r.scroll)
	if isPlaying {
		if r.playhead < r.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(r.playheadUpdate)})
		}
		r.listenToPlayerUpdates()
	}
	return layout.Dimensions{}
}
