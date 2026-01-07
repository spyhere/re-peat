package waverenderer

import (
	"fmt"
	"image"
	"image/color"
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
			pcmLen:      len(pcm),
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
	playhead       int
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
		return []float32{}
	}
	if chanNum > 2 {
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
	pxPerSec := r.scroll.minPxPerSec

	if r.scroll.deltaY != 0 {
		pxPerSec = r.scroll.deltaY
	}
	return int(float32(r.audio.sampleRate) / pxPerSec)
}

func (r *WavesRenderer) guardZoom(leftB int, rightB int) {
	if leftB == 0 {
		r.scroll.minX = 0
	}
	if rightB == r.audio.pcmMonoLen {
		r.scroll.maxX = r.scroll.deltaX
	} else {
		r.scroll.maxX = 1e38
	}
}

func (r *WavesRenderer) getRenderableWaves() [][2]float32 {
	samplesPerPx := r.getSamplesPerPx()
	maxSamples := samplesPerPx * r.size.X
	deltaSamples := int(r.scroll.deltaX) * samplesPerPx
	// TODO: zoom on scroll.originX
	leftB := int(min(max(0, deltaSamples), r.audio.pcmMonoLen-maxSamples))
	rightB := leftB + maxSamples
	r.guardZoom(leftB, rightB)
	if leftB == r.scroll.leftB && rightB == r.scroll.rightB {
		return r.cached
	}
	r.scroll.leftB = leftB
	r.scroll.rightB = rightB

	monoSamples := r.monoSamples[leftB:rightB]
	res := make([][2]float32, r.size.X)

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
	r.cached = res
	return res
}

func (r *WavesRenderer) SetSize(size image.Point) {
	r.size = size
	r.scroll.minPxPerSec = float32(size.X) / r.audio.seconds
}

func (r *WavesRenderer) handleClick(posX float32) {
	pxPerSec := max(r.scroll.minPxPerSec, r.scroll.deltaY)
	seconds := (posX / pxPerSec) + (float32(r.scroll.leftB) / float32(r.audio.sampleRate))
	// TODO: handle error here
	seekVal, _ := r.p.Search(seconds)
	r.playhead = int(seekVal)
}

const ZOOM_RATE = 0.01
const PAN_RATE = 0.2

func (r *WavesRenderer) handleScroll(scroll f32.Point, pos f32.Point) {
	r.scroll.originX = pos.X

	r.scroll.deltaX += scroll.X * PAN_RATE
	r.scroll.deltaX = clamp(r.scroll.minX, r.scroll.deltaX, r.scroll.maxX)

	r.scroll.deltaY += scroll.Y * ZOOM_RATE
	r.scroll.deltaY = clamp(r.scroll.minPxPerSec, r.scroll.deltaY, r.scroll.maxPxPerSec)
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
		r.playhead = int(player.GetReadAmount())
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

func (r *WavesRenderer) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
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
	// TODO: draw seconds measure

	playheadComp(gtx, r.playhead, r.audio, r.scroll)
	if isPlaying {
		if r.playhead < r.audio.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(r.playheadUpdate)})
		}
		r.listenToPlayerUpdates()
	}
	return layout.Dimensions{}
}
