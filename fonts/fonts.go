package fonts

import (
	"log"
	"math"
	"path/filepath"
	"strings"

	"embed"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"
)

//go:embed ttfs/*.ttf
var ttfs embed.FS

func LoadFonts(faces []text.FontFace) ([]text.FontFace, error) {
	fonts, err := ttfs.ReadDir("ttfs")
	if err != nil {
		return []text.FontFace{}, err
	}
	for _, it := range fonts {
		if it.IsDir() {
			continue
		}

		data, err := ttfs.ReadFile(filepath.Join("ttfs", it.Name()))
		if err != nil {
			return []text.FontFace{}, err
		}
		newFace, err := opentype.Parse(data)
		if err != nil {
			return []text.FontFace{}, err
		}

		fontName := strings.ToLower(it.Name())
		newFont := newFace.Font()
		switch {
		case strings.Contains(fontName, "extrabold"):
			newFont.Weight = font.ExtraBold
		case strings.Contains(fontName, "semibold"):
			newFont.Weight = font.SemiBold
		case strings.Contains(fontName, "bold"):
			newFont.Weight = font.Bold
		case strings.Contains(fontName, "black"):
			newFont.Weight = font.Black
		case strings.Contains(fontName, "medium"):
			newFont.Weight = font.Medium
		case strings.Contains(fontName, "extralight"):
			newFont.Weight = font.ExtraLight
		case strings.Contains(fontName, "light"):
			newFont.Weight = font.Light
		case strings.Contains(fontName, "thin"):
			newFont.Weight = font.Thin
		default:
			newFont.Weight = font.Normal
		}

		if strings.Contains(fontName, "italic") {
			newFont.Style = font.Italic
		} else {
			newFont.Style = font.Regular
		}

		faces = append(faces, text.FontFace{
			Font: newFont,
			Face: newFace,
		})
	}

	return faces, nil
}

var FontsCollection = make(fCollection)

const numOfStyles = 2

type fCollection map[font.Typeface]map[font.Weight][numOfStyles]bool

func (c fCollection) RememberFonts(fontFaces []text.FontFace) {
	for _, it := range fontFaces {
		if it.Font.Style >= numOfStyles {
			log.Printf("Style %v has more styles than expected - %v > %v. Ignoring\n", it.Font.Typeface, it.Font.Style, numOfStyles)
			continue
		}
		if c[it.Font.Typeface] == nil {
			c[it.Font.Typeface] = make(map[font.Weight][numOfStyles]bool)
		}
		styles := c[it.Font.Typeface][it.Font.Weight]
		styles[it.Font.Style] = true
		c[it.Font.Typeface][it.Font.Weight] = styles
	}
}
func (c fCollection) hasExact(f font.Typeface, w font.Weight, s font.Style) bool {
	if len(c) == 0 {
		return false
	}
	if s > numOfStyles {
		return false
	}
	if _, hasFace := c[f]; !hasFace {
		return false
	}
	if _, hasWeight := c[f][w]; !hasWeight {
		return false
	}
	if c[f][w][s] == true {
		return true
	}
	return false
}

func (c fCollection) debugPresence(f font.Typeface, w font.Weight, s font.Style) {
	if !c.hasExact(f, w, s) {
		log.Printf("%v %v %v doesn't exist\n", f, w, s)
	}
}

func (c fCollection) closestWeight(target font.Weight, face font.Typeface) font.Weight {
	best := target
	bestDist := math.MaxInt32
	for it := range c[face] {
		dW := int(it - target)
		if dW < 0 {
			dW = -dW
		}
		if dW < bestDist {
			bestDist = dW
			best = it
		}
	}
	return best
}
func (c fCollection) checkForStyle(face font.Typeface, weight font.Weight, style font.Style) font.Style {
	if style > numOfStyles {
		return font.Regular
	}
	if c[face][weight][style] == false {
		log.Printf("%v %v: has no %v style. Fallback -> Regular\n", face, weight.String(), style.String())
		if c[face][weight][font.Regular] == false {
			log.Printf("%v: has no Regular style. Gio expected to fallback to other typeface\n", face)
		}
		style = font.Regular
	}
	return style
}
func (c fCollection) useFont(typeFace font.Typeface, weight font.Weight, style font.Style) font.Font {
	c.debugPresence(typeFace, weight, style)
	fontFace := font.Font{
		Typeface: typeFace,
		Weight:   weight,
		Style:    style,
	}
	if len(c) == 0 {
		log.Println("Font collection is empty")
		return fontFace
	}
	if _, hasFace := c[typeFace]; !hasFace {
		log.Println(typeFace, "is not loaded. Fallback.")
		return fontFace
	}
	if _, hasWeight := c[typeFace][weight]; !hasWeight {
		newW := c.closestWeight(weight, typeFace)
		log.Printf("%v %v: is not loaded. Fallback to %v\n", typeFace, weight.String(), newW.String())
		fontFace.Weight = newW
		fontFace.Style = c.checkForStyle(typeFace, newW, style)
		return fontFace
	}
	fontFace.Style = c.checkForStyle(typeFace, weight, style)
	return fontFace
}

func Go(w font.Weight, s font.Style) font.Font {
	return FontsCollection.useFont("Go", w, s)
}
func GoMedium(w font.Weight, s font.Style) font.Font {
	return FontsCollection.useFont("Go Medium", w, s)
}
func GoMono(w font.Weight, s font.Style) font.Font {
	return FontsCollection.useFont("Go Mono", w, s)
}
func GoSmallcaps(w font.Weight, s font.Style) font.Font {
	return FontsCollection.useFont("Go Smallcaps", w, s)
}
func Roboto(w font.Weight, s font.Style) font.Font {
	return FontsCollection.useFont("Roboto", w, s)
}
