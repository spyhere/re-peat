package main

import (
	"log"
	"os"

	"github.com/gopxl/beep/mp3"
	"github.com/spyhere/re-peat/internal/audio"
)

func loadMonoSamples(path string) (monoSamples []float32, a audio.AudioMeta, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: use file extensions to figure out the file type
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return []float32{}, audio.AudioMeta{}, err
	}
	defer streamer.Close()

	buf := make([][2]float64, 1024)
	monoSamples = monoSamples[:0]
	for {
		n, ok := streamer.Stream(buf)
		if !ok {
			break
		}
		for i := range n {
			lSample := buf[i][0]
			rSample := buf[i][1]
			monoSamples = append(monoSamples, float32((lSample+rSample)*0.5))
		}
	}
	a = audio.NewAudioMeta(int(format.SampleRate), format.NumChannels, len(monoSamples))
	return monoSamples, a, nil
}
