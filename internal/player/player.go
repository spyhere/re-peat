package player

import (
	"bytes"
	"errors"
	"io"

	"github.com/ebitengine/oto/v3"
	"github.com/tosone/minimp3"
)

type Player struct {
	player  *oto.Player
	cReader *countingReader
	// Amount of PCM bytes
	dataLen int
}

func (p *Player) Play() {
	p.player.Play()
}

func (p *Player) Pause() {
	p.player.Pause()
	p.cReader.isActive = false
}

func (p *Player) IsPlaying() bool {
	return p.player.IsPlaying()
}

func (p *Player) SetVolume(volume float64) {
	p.player.SetVolume(volume)
}

const BYTES_PER_SAMPLE int64 = 2

func (p *Player) Search(offset float64) (int64, error) {
	value := int64(offset * float64(p.dataLen) / 100.0)
	value -= value % BYTES_PER_SAMPLE
	return p.player.Seek(value, io.SeekStart)
}

func (p *Player) BufferedSize() int {
	return p.player.BufferedSize()
}

func (p *Player) GetReadAmount() int64 {
	return p.cReader.bytes - int64(p.player.BufferedSize())
}

func (p *Player) WaitUntilReady() bool {
	return <-p.cReader.ready
}

func (p *Player) IsDoneCh() chan bool {
	return p.cReader.done
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
	cReader := &countingReader{
		r:     bytes.NewReader(data),
		ready: make(chan bool),
		done:  make(chan bool),
	}
	player := otoCtx.NewPlayer(cReader)
	return &Player{
		player:  player,
		cReader: cReader,
		dataLen: len(data),
	}, nil
}

type countingReader struct {
	r *bytes.Reader
	// What's the current offset
	bytes int64
	// Is reading at the moment
	isActive bool
	// User of this reader has started to read
	ready chan bool
	// EOF is beaing reached
	done chan bool
}

func (c *countingReader) Read(p []byte) (int, error) {
	if !c.isActive {
		c.isActive = true
		c.ready <- true
	}
	// Should read after the check otherwise it can be janky sometimes
	n, err := c.r.Read(p)
	if err != nil {
		if errors.Is(err, io.EOF) {
			c.done <- true
			c.isActive = false
		}
		return 0, err
	}
	c.bytes += int64(n)
	return n, err
}

func (c *countingReader) Seek(offset int64, whence int) (int64, error) {
	c.bytes = offset
	return c.r.Seek(offset, whence)
}
