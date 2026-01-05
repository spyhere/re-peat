package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/tosone/minimp3"
)

const maxUin16 float32 = 32767.0

func getNormalisedSamples(data []byte) ([]float32, error) {
	if len(data)%2 != 0 {
		return []float32{}, fmt.Errorf("Read samples are not uint16: %d\n", len(data))
	}
	normalised := make([]float32, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		// Normalize to -1..1
		normalised[i/2] = float32(sample) / maxUin16
	}
	return normalised, nil
}

// Entity to visualise wave forms of sound track
type WavesRenderer struct {
	CaretPos    int
	caretUpdate time.Duration
	SampleRate  int
	// Manual setting, otherwise it is calculated using max screen size
	PxPerSec float64
	PCMLen   int
	Samples  []float32
	// Temporal caching
	Waves  [][2]float32
	Player *Player
	// Total seconds of composition
	Seconds float64
	margin  int
	padding int
	// Max size of current widget
	Size      image.Point
	list      layout.List
	clickable widget.Clickable
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
	if r.PxPerSec > 0 {
		pxPerSec = r.PxPerSec
	} else {
		pxPerSec = float64(r.Size.X) / r.Seconds
	}
	return int(float64(r.SampleRate) / pxPerSec)
}

func (r *WavesRenderer) GetRenderableWaves() [][2]float32 {
	if len(r.Waves) > 0 {
		return r.Waves
	}
	samples := r.Samples
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
	r.Waves = res
	return res
}

func (r *WavesRenderer) SetSize(size image.Point) {
	r.Size = size
}

func (r *WavesRenderer) HandleClick(gtx layout.Context) {
	if r.clickable.Clicked(gtx) {
		clickHistory := r.clickable.History()
		pressX := clickHistory[len(clickHistory)-1].Position.X
		seekVal, _ := r.Player.Search(float64(pressX) * 100.0 / float64(r.Size.X))
		pressX = int(float64(seekVal) * float64(r.Size.X) / float64(r.PCMLen))
		r.CaretPos = pressX
	}
}

func (r *WavesRenderer) HandleKey(gtx layout.Context, isPlaying bool) {
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
					r.Player.Play()
					r.Player.WaitUntilReady()
				} else {
					r.Player.Pause()
				}
			}
		}
	}
}

func (r *WavesRenderer) Layout(gtx layout.Context, e app.FrameEvent) layout.Dimensions {
	player := r.Player
	isPlaying := player.IsPlaying()

	ColorBox(gtx, image.Rectangle{Max: image.Pt(r.Size.X, r.Size.Y)}, color.NRGBA{A: 0xff})
	bgArea := image.Rect(0, r.margin-r.padding, r.Size.X, r.Size.Y-r.margin+r.padding)
	ColorBox(gtx, bgArea, color.NRGBA{R: 0x11, G: 0x77, B: 0x66, A: 0xff})

	yMid := r.Size.Y / 2
	wavesYBorder := yMid - r.margin
	wavesYBorderF := float32(wavesYBorder)
	waves := r.GetRenderableWaves()
	listOffset := op.Offset(image.Pt(0, r.margin)).Push(gtx.Ops)
	r.list.Layout(gtx, len(waves), func(gtx layout.Context, idx int) layout.Dimensions {
		high := wavesYBorder - int(waves[idx][1]*wavesYBorderF)
		low := wavesYBorder - int(waves[idx][0]*wavesYBorderF)
		return ColorBox(gtx, image.Rect(0, high, 1, low), color.NRGBA{G: 0x32, B: 0x55, A: 0xff})
	})
	listOffset.Pop()

	r.HandleClick(gtx)
	activeA := op.Offset(bgArea.Min).Push(gtx.Ops)
	r.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(bgArea.Dx(), bgArea.Dy())}
	})
	activeA.Pop()

	r.HandleKey(gtx, isPlaying)
	if isPlaying {
		select {
		case _ = <-player.IsDoneCh():
			r.CaretPos = r.Size.X - 3
		default:
			rAmount := player.GetReadAmount()
			r.CaretPos = int(float64(rAmount) * float64(r.Size.X) / float64(r.PCMLen))
			gtx.Source.Execute(op.InvalidateCmd{At: e.Now.Add(r.caretUpdate)})
		}
	}
	ColorBox(gtx, image.Rect(r.CaretPos, 0, r.CaretPos+1, r.Size.Y), color.NRGBA{R: 0xff, G: 0xdd, B: 0xdd, A: 0xff})
	return layout.Dimensions{}
}

func NewWavesRenderer(dec *minimp3.Decoder, pcm []byte, player *Player) (*WavesRenderer, error) {
	normSamples, err := getNormalisedSamples(pcm)
	if err != nil {
		return &WavesRenderer{}, err
	}
	fmt.Println("Audio data is normalised")
	frames := len(normSamples) / dec.Channels
	monoSamples := makeSamplesMono(normSamples, dec.Channels)
	fmt.Println("WaveRenderer received mono samples")
	return &WavesRenderer{
		SampleRate:  dec.SampleRate,
		PCMLen:      len(pcm),
		Player:      player,
		Samples:     monoSamples,
		Seconds:     float64(frames) / float64(dec.SampleRate),
		list:        layout.List{},
		clickable:   widget.Clickable{},
		margin:      400,
		padding:     90,
		caretUpdate: time.Millisecond * 50,
	}, nil
}

func ColorBox(gtx layout.Context, size image.Rectangle, color color.NRGBA) layout.Dimensions {
	defer clip.Rect(size).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size.Size()}
}
