package fonts

import (
	"os"
	"path/filepath"
	"strings"

	"gioui.org/font/opentype"
	"gioui.org/text"
)

const fonstDir = "./fonts"

func LoadFonts(faces []text.FontFace) ([]text.FontFace, error) {
	fontPath, err := filepath.Abs(fonstDir)
	if err != nil {
		return []text.FontFace{}, err
	}
	fonts, err := os.ReadDir(fontPath)
	if err != nil {
		return []text.FontFace{}, err
	}
	for _, it := range fonts {
		absPath := filepath.Join(fontPath, it.Name())
		if !strings.HasSuffix(absPath, ".ttf") {
			continue
		}
		data, err := os.ReadFile(absPath)
		if err != nil {
			return []text.FontFace{}, err
		}
		newFace, err := opentype.Parse(data)
		if err != nil {
			return []text.FontFace{}, err
		}
		faces = append(faces, text.FontFace{
			Font: newFace.Font(),
			Face: newFace,
		})
	}

	return faces, nil
}
