package main

import (
	"time"

	"github.com/hajimehoshi/oto"
	"github.com/tosone/minimp3"
)

func player(dec *minimp3.Decoder, data []byte) error {
	var err error
	var context *oto.Context
	if context, err = oto.NewContext(dec.SampleRate, dec.Channels, 2, 1024); err != nil {
		return err
	}

	var player = context.NewPlayer()
	player.Write(data)

	<-time.After(time.Second)

	dec.Close()
	if err = player.Close(); err != nil {
		return err
	}
	return nil
}
