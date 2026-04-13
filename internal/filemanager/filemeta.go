package filemanager

import (
	"fmt"
	"strconv"
	"time"
)

func NewFileMeta(name string, size int64, updTime time.Time) FileMeta {
	return FileMeta{
		Name:      name,
		Size:      size,
		UpdatedAt: updTime,
	}
}

type FileMeta struct {
	Name      string
	Size      int64
	UpdatedAt time.Time
}

func (f FileMeta) SizeString() string {
	if f.Size == 0.0 {
		return ""
	}
	bytes := f.Size
	if bytes < 1000*1000 {
		kb := float64(bytes) / 1000
		return fmt.Sprintf("%.1f Kb", kb)
	} else if bytes < 1000*1000*1000 {
		mb := float64(bytes) / (1000 * 1000)
		return fmt.Sprintf("%.1f Mb", mb)
	}
	return strconv.Itoa(int(f.Size))
}
func (f FileMeta) UpdatedAtString() string {
	if f.UpdatedAt.IsZero() {
		return ""
	}
	return f.UpdatedAt.Format("Monday, 2 January 2006 at 15:04")
}
