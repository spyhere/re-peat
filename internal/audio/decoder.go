package audio

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/wav"
)

func Decode(f *os.File) (beep.StreamSeekCloser, beep.Format, error) {
	var (
		streamer beep.StreamSeekCloser
		format   beep.Format
		err      error
	)
	switch filepath.Ext(f.Name()) {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	case ".flac":
		streamer, format, err = flac.Decode(f)
	default:
		err = fmt.Errorf("Such format is not supported. Only supporting .mp3, .wav and .flac")
	}
	return streamer, format, err
}

// TODO: Pass slice here to avoid reallocations
func LoadMonoSamples(path string) (monoSamples []float32, a AudioMeta, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := Decode(file)
	if err != nil {
		return []float32{}, AudioMeta{}, err
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
	a = NewAudioMeta(int(format.SampleRate), format.NumChannels, len(monoSamples))
	return monoSamples, a, nil
}
