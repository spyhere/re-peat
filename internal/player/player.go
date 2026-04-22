package player

import (
	"math"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
	"github.com/spyhere/re-peat/internal/audio"
)

const defaultSampleRate = 44100

func NewPlayer() *Player {
	speaker.Init(defaultSampleRate, beep.SampleRate(defaultSampleRate).N(time.Second/10))
	return &Player{
		format: beep.Format{
			SampleRate:  defaultSampleRate,
			NumChannels: 2,
		},
	}
}

type Player struct {
	streamer  beep.StreamSeekCloser
	format    beep.Format
	ctrl      *beep.Ctrl
	volume    *effects.Volume
	isPlaying bool
	eof       bool
}

func (p *Player) attachStreamer(str beep.Streamer, f beep.Format) {
	if f.SampleRate != p.format.SampleRate {
		resampled := beep.Resample(4, f.SampleRate, p.format.SampleRate, p.volume)
		str = resampled
	}
	p.format = f
	speaker.Play(beep.Seq(str, beep.Callback(func() {
		p.eof = true
		p.isPlaying = false
	})))
}

func (p *Player) SetAudio(f *os.File) (audio.AudioMeta, error) {
	if p.streamer != nil {
		p.streamer.Close()
	}

	streamer, format, err := audio.Decode(f)
	if err != nil {
		return audio.AudioMeta{}, err
	}

	p.streamer = streamer
	p.ctrl = &beep.Ctrl{Streamer: streamer, Paused: true}
	p.volume = &effects.Volume{
		Streamer: p.ctrl,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}
	p.attachStreamer(p.volume, format)
	return audio.NewAudioMeta(int(format.SampleRate), format.NumChannels, streamer.Len()), nil
}

func (p *Player) SetVolume(volume float64) {
	if p.volume == nil {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	if volume <= 0 {
		p.volume.Silent = true
		return
	}
	p.volume.Silent = false
	v := math.Pow(volume, 2.0)
	p.volume.Volume = math.Log2(v)
}

func (p *Player) Play() {
	if p.eof {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	p.ctrl.Paused = false
	p.isPlaying = true
}

func (p *Player) Pause() {
	if p.eof {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	p.ctrl.Paused = true
	p.isPlaying = false
}

func (p *Player) IsPlaying() bool {
	return p.isPlaying
}
func (p *Player) IsEOF() bool {
	return p.eof
}

func (p *Player) Set(samples int) (int, error) {
	if p.eof == true {
		speaker.Lock()
		p.ctrl.Paused = true
		speaker.Unlock()
		p.eof = false
		p.attachStreamer(p.volume, p.format)
	}
	speaker.Lock()
	defer speaker.Unlock()
	err := p.streamer.Seek(samples)
	if err != nil {
		return 0, err
	}
	return p.streamer.Position(), nil
}

func (p *Player) Search(seconds float32) (int, error) {
	if p.eof == true {
		speaker.Lock()
		p.ctrl.Paused = true
		speaker.Unlock()
		p.eof = false
		p.attachStreamer(p.volume, p.format)
	}
	speaker.Lock()
	defer speaker.Unlock()
	dur := time.Duration(seconds * float32(time.Second))
	samplesN := p.format.SampleRate.N(dur)
	if err := p.streamer.Seek(samplesN); err != nil {
		return 0, err
	}
	return p.streamer.Position(), nil
}

func (p *Player) GetReadAmount() int {
	speaker.Lock()
	defer speaker.Unlock()
	return p.streamer.Position()
}

func (p *Player) GetCurrentSecond() float64 {
	return math.Round(float64(p.format.SampleRate.D(p.GetReadAmount())/time.Second)*100) / 100
}
