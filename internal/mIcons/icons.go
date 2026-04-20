package micons

import (
	"log"

	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func newIcon(i []byte) *widget.Icon {
	wI, err := widget.NewIcon(i)
	if err != nil {
		log.Println(err)
	}
	return wI
}

var (
	ContentAddCircle = newIcon(icons.ContentAddCircle)
	Delete           = newIcon(icons.ActionDelete)
	Check            = newIcon(icons.NavigationCheck)
	Search           = newIcon(icons.ActionSearch)
	Cancel           = newIcon(icons.NavigationCancel)
	Close            = newIcon(icons.NavigationClose)
	Play             = newIcon(icons.AVPlayArrow)
	Replay           = newIcon(icons.AVReplay)
	Pause            = newIcon(icons.AVPause)
	Filter           = newIcon(icons.ContentFilterList)
	Edit             = newIcon(icons.EditorModeEdit)
	Warning          = newIcon(icons.AlertWarning)
	Tick             = newIcon(icons.NavigationCheck)
	Comment          = newIcon(icons.EditorModeComment)
	CommentInsert    = newIcon(icons.EditorInsertComment)
	Folder           = newIcon(icons.FileFolder)
	Save             = newIcon(icons.ContentSave)
	Info             = newIcon(icons.ActionInfo)
)
