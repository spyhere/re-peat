package player

import (
	"bytes"
	"io"

	"github.com/ebitengine/oto/v3"
	"github.com/spyhere/re-peat/internal/constants"
	"github.com/tosone/minimp3"
)

type Player struct {
	player *oto.Player
	reader *bytes.Reader
	// Amount of PCM bytes
	dataLen  int
	totalSec float32
}

func (p *Player) Play() {
	p.player.Play()
}

func (p *Player) Pause() {
	p.player.Pause()
}

func (p *Player) IsPlaying() bool {
	return p.player.IsPlaying()
}

func (p *Player) SetVolume(volume float64) {
	p.player.SetVolume(volume)
}

func (p *Player) Search(seconds float32) (int64, error) {
	value := int64(seconds * float32(p.dataLen) / p.totalSec)
	value -= value % constants.BytesPerSample
	return p.player.Seek(value, io.SeekStart)
}

func (p *Player) Set(pcm int64) (int64, error) {
	return p.player.Seek(pcm, io.SeekStart)
}

func (p *Player) BufferedSize() int {
	return p.player.BufferedSize()
}

func (p *Player) GetReadAmount() int64 {
	current, _ := p.reader.Seek(0, io.SeekCurrent)
	return current - int64(p.player.BufferedSize())
}

func NewPlayer(dec *minimp3.Decoder, data []byte) (*Player, error) {
	var err error
	op := &oto.NewContextOptions{
		SampleRate:   dec.SampleRate,
		ChannelCount: dec.Channels,
		Format:       oto.FormatSignedInt16LE,
	}
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return &Player{}, err
	}
	<-readyChan
	reader := bytes.NewReader(data)
	player := otoCtx.NewPlayer(reader)
	return &Player{
		player:   player,
		reader:   reader,
		dataLen:  len(data),
		totalSec: float32(len(data)) / float32(dec.SampleRate) * 0.25,
	}, nil
}
