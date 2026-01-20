package editor

import (
	"encoding/binary"
	"fmt"
)

const maxUin16 float32 = 32767.0

// TODO: move to helpers
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
