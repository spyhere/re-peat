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
)
