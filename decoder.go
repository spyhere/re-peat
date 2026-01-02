package main

import (
	"os"

	"github.com/tosone/minimp3"
)

func decodeFile(filePath string) (*minimp3.Decoder, []byte, error) {
	var err error
	var dec *minimp3.Decoder
	var data []byte

	var file []byte
	if file, err = os.ReadFile(filePath); err != nil {
		return dec, data, err
	}

	if dec, data, err = minimp3.DecodeFull(file); err != nil {
		return dec, data, err
	}
	return dec, data, nil
}
