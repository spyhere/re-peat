package common

import (
	"cmp"
	"fmt"
	"image"
	"log"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// TODO: Make it easier to work with "GetCursorType" ("Cursor")
func SetCursor(gtx layout.Context, cursor pointer.Cursor) {
	pointer.Cursor(cursor).Add(gtx.Ops)
}

func OffsetBy(gtx layout.Context, amount image.Point, w func(gtx layout.Context)) {
	defer op.Offset(amount).Push(gtx.Ops).Pop()
	w(gtx)
}

func CenteredX(gtx layout.Context, w func() layout.Dimensions) layout.Dimensions {
	macro, dimensions := MakeMacro(gtx, func(gtx layout.Context) layout.Dimensions {
		return w()
	})
	xCenter := gtx.Constraints.Max.X / 2
	wCenter := dimensions.Size.X / 2
	OffsetBy(gtx, image.Pt(xCenter-wCenter, 0), func(gtx layout.Context) {
		macro.Add(gtx.Ops)
	})
	return dimensions
}

func RegisterTag(gtx layout.Context, tag event.Tag, area image.Rectangle) {
	defer clip.Rect(area).Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, tag)
}

func HandlePointerEvents(gtx layout.Context, tag event.Tag, pKind pointer.Kind, cb func(e pointer.Event)) {
	for {
		evt, ok := gtx.Event(pointer.Filter{
			Target:  tag,
			Kinds:   pKind,
			ScrollX: pointer.ScrollRange{Min: -1e9, Max: 1e9},
			ScrollY: pointer.ScrollRange{Min: -1e9, Max: 1e9},
		})
		if !ok {
			break
		}
		e, ok := evt.(pointer.Event)
		if !ok {
			continue
		}
		cb(e)
	}
}

func HandleKeyEvents(gtx layout.Context, cb func(e key.Event), filters ...event.Filter) {
	for {
		evt, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		e, ok := evt.(key.Event)
		if !ok {
			continue
		}
		cb(e)
	}
}

func MakeMacro(gtx layout.Context, cb func(gtx layout.Context) layout.Dimensions) (op.CallOp, layout.Dimensions) {
	macro := op.Record(gtx.Ops)
	dims := cb(gtx)
	return macro.Stop(), dims
}

// Lock "this" between "from" and "to"
func Clamp[T cmp.Ordered](from T, this T, to T) T {
	if to < from {
		log.Panicf("Clamp has received TO: %v smaller than FROM: %v", to, from)
	}
	return min(max(from, this), to)
}

// TODO: Get rig of Atoi on every frame
func PrcToPx(origin int, prc string) int {
	prc = strings.Split(prc, "%")[0]
	prcInt, err := strconv.Atoi(prc)
	if err != nil {
		log.Fatal(err)
	}
	return prcInt * origin / 100
}

func Snap(v float32) float32 {
	return float32(math.Round(float64(v)))
}

func strlen(input string) int {
	return utf8.RuneCountInString(input)
}

// Since non-lating letters are taking more then 1 byte `strlen` and manual idx in range is required
func StrTrunc(name string, limit int) string {
	if limit == 0 || strlen(name) < limit {
		return name
	}
	var newName strings.Builder
	idx := 0
	for _, r := range name {
		if idx >= limit-3 {
			break
		}
		newName.WriteRune(r)
		idx++
	}
	newName.WriteString("...")
	return newName.String()
}

func FormatSeconds(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("00:%02d", int(seconds))
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%02d:%02d", int(minutes), int(math.Mod(seconds, 60)))
	}
	hours := minutes / 60
	return fmt.Sprintf("%d:%02d:%02d", int(hours), int(math.Mod(minutes, 60)), int(math.Mod(seconds, 60)))
}

func ParseSeconds(secondsStr string) (float64, error) {
	if secondsStr == "" || !strings.Contains(secondsStr, ":") {
		return 0, fmt.Errorf("Given incorrect string to parse from seconds")
	}
	parts := strings.Split(secondsStr, ":")
	m, s := parts[0], parts[1]
	minutes, err := strconv.Atoi(m)
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return float64(minutes*60 + seconds), nil
}

func ParseSize(size int64) string {
	if size == 0.0 {
		return ""
	}
	bytes := size
	if bytes < 1000*1000 {
		kb := float64(bytes) / 1000
		return fmt.Sprintf("%.1f Kb", kb)
	} else if bytes < 1000*1000*1000 {
		mb := float64(bytes) / (1000 * 1000)
		return fmt.Sprintf("%.1f Mb", mb)
	}
	return strconv.Itoa(int(size))
}
