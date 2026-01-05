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
		p:              player,
		samples:        monoSamples,
		seconds:        float64(frames) / float64(dec.SampleRate),
		margin:         400,
		padding:        90,
		playheadUpdate: time.Millisecond * 50,
	}, nil
}

// Entity to visualise wave forms of sound track
type WavesRenderer struct {
	playhead       int
	playheadUpdate time.Duration
	sampleRate     int
	// Manual setting, otherwise it is calculated using max screen size
	pxPerSec float64
	pcmLen   int
	samples  []float32
	// Temporal caching
	waves [][2]float32
	p     *player.Player
	// Total seconds of composition
	seconds float64
	margin  int
	padding int
	// Max size of current widget
	size   image.Point
	deltaX float32
	deltaY float32
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
	var pxPerSec float64
	if r.pxPerSec > 0 {
		pxPerSec = r.pxPerSec
	} else {
		pxPerSec = float64(r.size.X) / r.seconds
	}
	return int(float64(r.sampleRate) / pxPerSec)
}

func (r *WavesRenderer) getRenderableWaves() [][2]float32 {
	if len(r.waves) > 0 {
		return r.waves
	}
	samples := r.samples
	samplesPerPx := r.getSamplesPerPx()
	res := make([][2]float32, len(samples)/samplesPerPx)

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
}

func (r *WavesRenderer) handleClick(posX float32) {
	seekVal, _ := r.p.Search(posX * 100.0 / float32(r.size.X))
	r.playhead = int(seekVal)
}

func (r *WavesRenderer) handleScroll(point f32.Point) {
	r.deltaX = point.X
	r.deltaY = point.Y
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
