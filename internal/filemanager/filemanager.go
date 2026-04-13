package filemanager

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/x/explorer"
)

func NewFileManager(window *app.Window) *FileManager {
	return &FileManager{
		e:      explorer.NewExplorer(window),
		window: window,
	}
}

type FileManager struct {
	e        *explorer.Explorer
	window   *app.Window
	choosing bool
}

func (f *FileManager) Load(cb func(string, error), extensions ...string) {
	f.choosing = true
	go func(cb func(string, error)) {
		file, err := f.e.ChooseFile(extensions...)
		defer func() {
			if file != nil {
				file.Close()
			}
			f.choosing = false
			f.window.Invalidate()
		}()

		if err != nil {
			cb("", err)
			return
		}

		if fName, ok := file.(*os.File); ok {
			f.window.Invalidate()
			cb(fName.Name(), nil)
		} else {
			cb("", fmt.Errorf("is not a file"))
		}
	}(cb)
}

func (f *FileManager) Save() {
	log.Fatal("Save is not implemented")
}

func (f *FileManager) SaveAs() {
	log.Fatal("Save is not implemented")
}

func (f *FileManager) IsChoosing() bool {
	return f.choosing
}
