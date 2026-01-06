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
		sampleRate:     dec.SampleRate,
		pcmLen:         len(pcm),
		pcmMonoLen:     len(monoSamples),
		p:              player,
		samples:        monoSamples,
		seconds:        float32(frames) / float32(dec.SampleRate),
		margin:         400,
		padding:        90,
		playheadUpdate: time.Millisecond * 50,
		maxPxPerSec:    200,
		zoom:           zoom{},
	}, nil
}

// Entity to visualise wave forms of sound track
type WavesRenderer struct {
	playhead       int
	playheadUpdate time.Duration
	sampleRate     int
	minPxPerSec    float32
	maxPxPerSec    float32
	pcmLen         int
	pcmMonoLen     int
	samples        []float32
	// Temporal caching
	waves [][2]float32
	p     *player.Player
	// Total seconds of composition
	seconds float32
	margin  int
	padding int
	// Max size of current widget
	size image.Point
	zoom zoom
}

type zoom struct {
	minX     float32
	maxX     float32
	deltaX   float32
	deltaY   float32
	pxPerSec float32
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
	pxPerSec := r.minPxPerSec

	if r.zoom.pxPerSec != 0 {
		pxPerSec = r.zoom.pxPerSec
	}
	return int(float32(r.sampleRate) / pxPerSec)
}

func (r *WavesRenderer) guardZoom(leftB int, rightB int) {
	if leftB == 0 {
		r.zoom.minX = 0
	}
	if rightB == r.pcmMonoLen {
		r.zoom.maxX = r.zoom.deltaX
	} else {
		r.zoom.maxX = 1e38
	}
}

func (r *WavesRenderer) getRenderableWaves() [][2]float32 {
	samplesPerPx := r.getSamplesPerPx()
	maxSamples := samplesPerPx * r.size.X
	leftB := int(min(max(0, r.zoom.deltaX*200.0), float32(r.pcmMonoLen-maxSamples)))
	rightB := leftB + maxSamples
	r.guardZoom(leftB, rightB)
	// cache layer here

	samples := r.samples[leftB:rightB]
	res := make([][2]float32, r.size.X)

	var idx int
	var min float32 = 1
	var max float32 = -1
	count := samplesPerPx
	for _, it := range samples {
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
	r.waves = res
	return res
}

func (r *WavesRenderer) SetSize(size image.Point) {
	r.size = size
	r.minPxPerSec = float32(size.X) / r.seconds
}

func (r *WavesRenderer) handleClick(posX float32) {
	seekVal, _ := r.p.Search(posX * 100.0 / float32(r.size.X))
	r.playhead = int(seekVal)
}

func (r *WavesRenderer) handleScroll(point f32.Point) {
	r.zoom.deltaX += point.X
	r.zoom.deltaX = min(max(r.zoom.minX, r.zoom.deltaX), r.zoom.maxX)

	r.zoom.deltaY += point.Y
	minPx := r.minPxPerSec * 100.0
	maxPx := r.maxPxPerSec * 100.0
	r.zoom.deltaY = min(max(minPx, r.zoom.deltaY), maxPx)
	r.zoom.pxPerSec = r.zoom.deltaY * 0.01
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
					if r.playhead >= r.pcmLen {
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
		r.playhead = r.pcmLen
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
			r.handleScroll(e.Scroll)
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

	playheadComp(gtx, r.playhead, r.pcmLen)
	if isPlaying {
		if r.playhead < r.pcmLen {
			gtx.Source.Execute(op.InvalidateCmd{At: gtx.Now.Add(r.playheadUpdate)})
		}
		r.listenToPlayerUpdates()
	}
	return layout.Dimensions{}
}
