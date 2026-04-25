package filemanager

import (
	"time"

	"github.com/spyhere/re-peat/internal/common"
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
	return common.ParseSize(f.Size)
}
func (f FileMeta) UpdatedAtString() string {
	if f.UpdatedAt.IsZero() {
		return ""
	}
	return f.UpdatedAt.Format("Monday, 2 January 2006 at 15:04")
}
