package player

import (
	"bytes"
	"errors"
	"io"

	"github.com/ebitengine/oto/v3"
	"github.com/spyhere/re-peat/internal/constants"
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

func (p *Player) Search(seconds float32) (int64, error) {
	value := int64(seconds * float32(p.dataLen) / p.totalSec)
	value -= value % constants.BYTES_PER_SAMPLE
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

/*
You MUST(!) to pause the player when cReader tells it's done,
otherwise there can be an edge case bug:
When the player is playing the last portion of bytes and you are
trying to seek, then player's native "Seek" method blocks forever.
It's impossible to do it in this method, since we cannot read from
the channel without emptying it and I need checking for state without
blocking the thread (select clause).
*/
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
		return n, err
	}
	c.bytes += int64(n)
	return n, err
}

func (c *countingReader) Seek(offset int64, whence int) (int64, error) {
	c.bytes = offset
	return c.r.Seek(offset, whence)
}
