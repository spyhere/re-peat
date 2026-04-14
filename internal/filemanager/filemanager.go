package filemanager

import (
	"fmt"
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

func (f *FileManager) Save(filePath string, data []byte, cb func(error)) {
	go func(cb func(error)) {
		var (
			file *os.File
			err  error
		)
		defer func() {
			if err != nil {
				cb(err)
			}
			if file != nil {
				file.Close()
			}
			f.window.Invalidate()
		}()
		file, err = os.Create(filePath)
		if err != nil {
			return
		}
		_, err = file.Write(data)
		if err != nil {
			return
		}
	}(cb)
}

type namer interface {
	Name() string
}

func (f *FileManager) SaveAs(defaultName string, data []byte, cb func(string, error)) {
	go func(cb func(string, error)) {
		wc, err := f.e.CreateFile(defaultName)
		defer func() {
			if wc != nil {
				wc.Close()
			}
			filePath := ""
			if n, ok := wc.(namer); ok {
				filePath = n.Name()
			}
			if err != nil {
				cb("", err)
			} else {
				cb(filePath, nil)
			}
			f.window.Invalidate()
		}()

		if err != nil {
			return
		}
		_, err = wc.Write(data)
		if err != nil {
			return
		}
	}(cb)
}

func (f *FileManager) IsChoosing() bool {
	return f.choosing
}
